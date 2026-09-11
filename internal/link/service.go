package link

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"url-shortener/internal/db"

	lru "github.com/hashicorp/golang-lru/v2"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrInvalidURL   = errors.New("invalid destination URL")
	ErrReservedSlug = errors.New("slug is a reserved path")
	ErrSlugExists   = errors.New("slug already exists")
	ErrNotFound     = errors.New("resource not found")
)

type Service interface {
	Create(ctx context.Context, req CreateLinkRequest) (Link, error)
	GetByID(ctx context.Context, id int64) (Link, error)
	List(ctx context.Context) ([]Link, error)
	Resolve(ctx context.Context, fullPath string) (Link, string, error)

	RecordClick(event ClickEvent)
	GetClickByID(ctx context.Context, id int64) (ClickEvent, error)
	ListClicks(ctx context.Context, linkID int64) ([]ClickEvent, error)
	Close()
}

type service struct {
	repo  Repository
	cache *lru.Cache[string, Link]
}

func NewService(repo Repository, cacheSize int) (Service, error) {
	cache, err := lru.New[string, Link](cacheSize)
	if err != nil {
		return nil, err
	}
	return &service{repo: repo, cache: cache}, nil
}

func (s *service) Create(ctx context.Context, req CreateLinkRequest) (Link, error) {
	parsed, err := url.ParseRequestURI(req.DestinationURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return Link{}, ErrInvalidURL
	}

	slug := strings.Trim(strings.TrimSpace(req.Slug), "/")
	isCustom := slug != ""

	if isCustom && isReservedSlug(slug) {
		return Link{}, ErrReservedSlug
	}

	for {
		if slug == "" {
			var err error
			slug, err = generateBase62ID(6)
			if err != nil {
				return Link{}, err
			}
		}

		created, err := s.repo.Create(ctx, slug, req.DestinationURL, isCustom, req.ExpiresAt)
		if err == nil {
			s.cache.Add(created.Slug, created)
			return created, nil
		}

		if isCustom || !db.IsUniqueConstraintError(err) {
			if db.IsUniqueConstraintError(err) {
				return Link{}, ErrSlugExists
			}
			return Link{}, err
		}

		slug = ""
	}
}

func (s *service) GetByID(ctx context.Context, id int64) (Link, error) {
	l, found, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Link{}, err
	}
	if !found {
		return Link{}, ErrNotFound
	}
	return l, nil
}

func (s *service) List(ctx context.Context) ([]Link, error) {
	return s.repo.List(ctx)
}

func (s *service) Resolve(ctx context.Context, fullPath string) (Link, string, error) {
	if l, ok := s.cache.Get(fullPath); ok {
		if l.ExpiresAt == nil || l.ExpiresAt.After(time.Now()) {
			return l, "", nil
		}
		s.cache.Remove(fullPath)
	}

	currentPath := fullPath
	for {
		l, found, err := s.repo.GetBySlug(ctx, currentPath)
		if err != nil {
			return Link{}, "", err
		}

		if found {
			extraPath := strings.TrimPrefix(fullPath, currentPath)
			s.cache.Add(currentPath, l)
			return l, extraPath, nil
		}

		idx := strings.LastIndex(currentPath, "/")
		if idx == -1 {
			break
		}
		currentPath = currentPath[:idx]
	}

	return Link{}, "", ErrNotFound
}

func (s *service) RecordClick(event ClickEvent) {
	s.repo.RecordClick(event)
}

func (s *service) GetClickByID(ctx context.Context, id int64) (ClickEvent, error) {
	c, found, err := s.repo.GetClickByID(ctx, id)
	if err != nil {
		return ClickEvent{}, err
	}
	if !found {
		return ClickEvent{}, ErrNotFound
	}
	return c, nil
}

func (s *service) ListClicks(ctx context.Context, linkID int64) ([]ClickEvent, error) {
	return s.repo.ListClicksByLinkID(ctx, linkID)
}

func (s *service) Close() {
	s.repo.Close()
}

func generateBase62ID(length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random ID: %w", err)
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}

func isReservedSlug(slug string) bool {
	switch slug {
	case "api", "_health", "_app":
		return true
	}
	return strings.HasPrefix(slug, "api/")
}
