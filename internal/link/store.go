package link

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"url-shortener/internal/db"
)

const linkColumns = `
	id,
	slug,
	destination_url,
	is_custom,
	click_count,
	created_at,
	expires_at
`

const (
	clickBatchSize     = 100
	clickBufferSize    = 2000
	clickFlushInterval = 2 * time.Second
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
		clickChan: make(chan ClickEvent, clickBufferSize),
	}

	s.wg.Add(1)
	go s.processClicks()

	return s, nil
}

func (s *Store) Create(
	ctx context.Context,
	slug string,
	destinationURL string,
	isCustom bool,
	expiresAt *time.Time,
) (Link, error) {
	for {
		query := `
			INSERT INTO links (
				slug,
				destination_url,
				is_custom,
				expires_at
			)
			VALUES (?, ?, ?, ?)
			RETURNING ` + linkColumns

		link, err := db.Scan[Link](
			s.db.QueryRowContext(
				ctx,
				query,
				slug,
				destinationURL,
				isCustom,
				expiresAt,
			),
		)

		if err == nil {
			s.cache.Add(link.Slug, link)
			return link, nil
		}

		if isCustom || !db.IsUniqueConstraintError(err) {
			return Link{}, err
		}

		slug = GenerateBase62ID(6)
	}
}

func (s *Store) Resolve(ctx context.Context, fullPath string) (Link, string, bool) {
	if link, ok := s.cache.Get(fullPath); ok {
		if link.ExpiresAt == nil || link.ExpiresAt.After(time.Now()) {
			return link, "", true
		}

		s.cache.Remove(fullPath)
	}

	currentPath := fullPath

	for {
		query := `
			SELECT ` + linkColumns + `
			FROM links
			WHERE slug = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		`

		resolved, err := db.Scan[Link](
			s.db.QueryRowContext(ctx, query, currentPath),
		)

		if err == nil {
			extraPath := strings.TrimPrefix(fullPath, currentPath)
			s.cache.Add(currentPath, resolved)
			return resolved, extraPath, true
		}

		if err != sql.ErrNoRows {
			return Link{}, "", false
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

	ticker := time.NewTicker(clickFlushInterval)
	defer ticker.Stop()

	batch := make([]ClickEvent, 0, clickBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := s.flushClicks(batch); err != nil {
			return
		}

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

			if len(batch) >= clickBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func (s *Store) flushClicks(batch []ClickEvent) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, click := range batch {
		if err := db.Insert(tx, "link_clicks", click); err != nil {
			return err
		}

		if _, err := tx.Exec(`
			UPDATE links
			SET click_count = click_count + 1
			WHERE id = ?
		`, click.LinkID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
