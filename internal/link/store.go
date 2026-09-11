package link

import (
	"context"
	"database/sql"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"url-shortener/internal/db"
)

type Store struct {
	db     *sql.DB
	cache  *lru.Cache[string, Link]
	clicks *clickTracker
}

func NewStore(db *sql.DB, cacheSize int) (*Store, error) {
	cache, err := lru.New[string, Link](cacheSize)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:     db,
		cache:  cache,
		clicks: newClickTracker(db),
	}, nil
}

func (s *Store) Create(
	ctx context.Context,
	slug string,
	destinationURL string,
	isCustom bool,
	expiresAt *time.Time,
) (Link, error) {
	for {
		var err error
		if slug == "" {
			slug, err = GenerateBase62ID(6)
			if err != nil {
				return Link{}, err
			}
		}

		query := `
            INSERT INTO links (
                slug,
                destination_url,
                is_custom,
                expires_at
            )
            VALUES (?, ?, ?, ?)
            RETURNING id, slug, destination_url, is_custom, click_count, created_at, expires_at
        `

		var l Link
		err = s.db.QueryRowContext(
			ctx,
			query,
			slug,
			destinationURL,
			isCustom,
			expiresAt,
		).Scan(
			&l.ID,
			&l.Slug,
			&l.DestinationURL,
			&l.IsCustom,
			&l.ClickCount,
			&l.CreatedAt,
			&l.ExpiresAt,
		)

		if err == nil {
			s.cache.Add(l.Slug, l)
			return l, nil
		}

		if isCustom || !db.IsUniqueConstraintError(err) {
			return Link{}, err
		}

		slug = ""
	}
}

func (s *Store) GetByID(ctx context.Context, id int64) (Link, bool, error) {
	var l Link
	err := s.db.QueryRowContext(ctx, `
        SELECT id, slug, destination_url, is_custom, click_count, created_at, expires_at
        FROM links
        WHERE id = ?
    `, id).Scan(
		&l.ID,
		&l.Slug,
		&l.DestinationURL,
		&l.IsCustom,
		&l.ClickCount,
		&l.CreatedAt,
		&l.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	if err != nil {
		return Link{}, false, err
	}

	return l, true, nil
}

func (s *Store) List(ctx context.Context) ([]Link, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, slug, destination_url, is_custom, click_count, created_at, expires_at
        FROM links
        ORDER BY id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(
			&l.ID,
			&l.Slug,
			&l.DestinationURL,
			&l.IsCustom,
			&l.ClickCount,
			&l.CreatedAt,
			&l.ExpiresAt,
		); err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, rows.Err()
}

func (s *Store) ListClicksByLinkID(ctx context.Context, linkID int64) ([]ClickEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, link_id, referer, user_agent, clicked_at
        FROM link_clicks
        WHERE link_id = ?
        ORDER BY clicked_at DESC
    `, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []ClickEvent
	for rows.Next() {
		var c ClickEvent
		if err := rows.Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt); err != nil {
			return nil, err
		}
		clicks = append(clicks, c)
	}

	return clicks, rows.Err()
}

func (s *Store) GetClickByID(ctx context.Context, id int64) (ClickEvent, bool, error) {
	var c ClickEvent
	err := s.db.QueryRowContext(ctx, `
        SELECT id, link_id, referer, user_agent, clicked_at
        FROM link_clicks
        WHERE id = ?
    `, id).Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt)

	if err == sql.ErrNoRows {
		return ClickEvent{}, false, nil
	}
	if err != nil {
		return ClickEvent{}, false, err
	}

	return c, true, nil
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
		var l Link
		err := s.db.QueryRowContext(ctx, `
            SELECT id, slug, destination_url, is_custom, click_count, created_at, expires_at
            FROM links
            WHERE slug = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
        `, currentPath).Scan(
			&l.ID,
			&l.Slug,
			&l.DestinationURL,
			&l.IsCustom,
			&l.ClickCount,
			&l.CreatedAt,
			&l.ExpiresAt,
		)

		if err == nil {
			extraPath := strings.TrimPrefix(fullPath, currentPath)
			s.cache.Add(currentPath, l)
			return l, extraPath, true
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
	s.clicks.record(event)
}

func (s *Store) Close() {
	s.clicks.close()
}
