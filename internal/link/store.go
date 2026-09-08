package link

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Store struct {
	db        *sql.DB
	cache     *lru.Cache[string, Link]
	clickChan chan ClickEvent
	wg        sync.WaitGroup
}

func NewStore(db *sql.DB, cacheSize int) (*Store, error) {
	cache, err := lru.New[string, Link](cacheSize)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db:        db,
		cache:     cache,
		clickChan: make(chan ClickEvent, 2000),
	}

	s.wg.Add(1)
	go s.processClicks()

	return s, nil
}

func (s *Store) Create(ctx context.Context, slug, destinationURL string, isCustom bool) (Link, error) {
	query := `INSERT INTO links (slug, destination_url, is_custom) VALUES (?, ?, ?)`
	res, err := s.db.ExecContext(ctx, query, slug, destinationURL, isCustom)
	if err != nil {
		return Link{}, err
	}

	id, _ := res.LastInsertId()
	link := Link{
		ID:             id,
		Slug:           slug,
		DestinationURL: destinationURL,
		IsCustom:       isCustom,
		CreatedAt:      time.Now(),
	}

	s.cache.Add(slug, link)
	return link, nil
}

func (s *Store) Resolve(ctx context.Context, fullPath string) (Link, string, bool) {
	if link, ok := s.cache.Get(fullPath); ok {
		return link, "", true
	}

	currentPath := fullPath
	for {
		var link Link
		query := `SELECT id, slug, destination_url, is_custom, click_count, created_at FROM links WHERE slug = ?`
		err := s.db.QueryRowContext(ctx, query, currentPath).Scan(
			&link.ID, &link.Slug, &link.DestinationURL, &link.IsCustom, &link.ClickCount, &link.CreatedAt,
		)

		if err == nil {
			extraPath := strings.TrimPrefix(fullPath, currentPath)
			s.cache.Add(currentPath, link)
			return link, extraPath, true
		}

		idx := strings.LastIndex(currentPath, "/")
		if idx == -1 {
			break
		}
		currentPath = currentPath[:idx]
	}

	return Link{}, "", false
}

func (s *Store) RecordClick(event ClickEvent) {
	select {
	case s.clickChan <- event:
	default:
	}
}

func (s *Store) Close() {
	close(s.clickChan)
	s.wg.Wait()
}

func (s *Store) processClicks() {
	defer s.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var batch []ClickEvent

	flush := func() {
		if len(batch) == 0 {
			return
		}

		tx, err := s.db.Begin()
		if err != nil {
			batch = batch[:0]
			return
		}

		stmtClick, _ := tx.Prepare("INSERT INTO link_clicks (link_id, referer, user_agent, clicked_at) VALUES (?, ?, ?, ?)")
		stmtCount, _ := tx.Prepare("UPDATE links SET click_count = click_count + 1 WHERE id = ?")

		if stmtClick != nil {
			defer stmtClick.Close()
		}
		if stmtCount != nil {
			defer stmtCount.Close()
		}

		for _, click := range batch {
			if stmtClick != nil {
				stmtClick.Exec(click.LinkID, click.Referer, click.UserAgent, click.ClickedAt)
			}
			if stmtCount != nil {
				stmtCount.Exec(click.LinkID)
			}
		}

		_ = tx.Commit()
		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-s.clickChan:
			if !ok {
				flush()
				return
			}
			batch = append(batch, event)
			if len(batch) >= 100 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func BuildTargetURL(rawDest string, incomingQuery map[string][]string, extraPath string) (string, error) {
	return rawDest, nil
}
