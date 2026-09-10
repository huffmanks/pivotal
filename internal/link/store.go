package link

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

const linkColumns = `
	id,
	slug,
	destination_url,
	is_custom,
	click_count,
	created_at
`

// ADD above
// TODO expires_at

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

func (s *Store) Create(ctx context.Context, slug, destinationURL string, isCustom bool) (Link, error) {
	query := `
		INSERT INTO links (slug, destination_url, is_custom)
		VALUES (?, ?, ?)
		RETURNING ` + linkColumns

	link, err := scanLink(
		s.db.QueryRowContext(ctx, query, slug, destinationURL, isCustom),
	)
	if err != nil {
		return Link{}, err
	}

	s.cache.Add(link.Slug, link)

	return link, nil
}

func (s *Store) List(ctx context.Context) ([]Link, error) {
	query := `
		SELECT ` + linkColumns + `
		FROM links
		ORDER BY id DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]Link, 0)

	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (s *Store) GetByID(ctx context.Context, id int64) (Link, bool, error) {
	query := `
		SELECT ` + linkColumns + `
		FROM links
		WHERE id = ?
	`

	link, err := scanLink(
		s.db.QueryRowContext(ctx, query, id),
	)

	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}

	if err != nil {
		return Link{}, false, err
	}

	return link, true, nil
}

func (s *Store) Resolve(ctx context.Context, fullPath string) (Link, string, bool) {
	if link, ok := s.cache.Get(fullPath); ok {
		return link, "", true
	}

	currentPath := fullPath

	for {
		query := `
			SELECT ` + linkColumns + `
			FROM links
			WHERE slug = ?
		`

		link, err := scanLink(
			s.db.QueryRowContext(ctx, query, currentPath),
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

	batch := make([]ClickEvent, 0, 100)

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

			if len(batch) >= 100 {
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

	stmtClick, err := tx.Prepare(`
		INSERT INTO link_clicks (
			link_id,
			referer,
			user_agent,
			clicked_at
		)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtClick.Close()

	stmtCount, err := tx.Prepare(`
		UPDATE links
		SET click_count = click_count + 1
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmtCount.Close()

	for _, click := range batch {
		if _, err := stmtClick.Exec(
			click.LinkID,
			click.Referer,
			click.UserAgent,
			click.ClickedAt,
		); err != nil {
			return err
		}

		if _, err := stmtCount.Exec(click.LinkID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func scanLink(scanner interface {
	Scan(dest ...any) error
}) (Link, error) {
	var link Link

	err := scanner.Scan(
		&link.ID,
		&link.Slug,
		&link.DestinationURL,
		&link.IsCustom,
		&link.ClickCount,
		&link.CreatedAt,
		// TODO &link.ExpiresAt,
	)

	return link, err
}
