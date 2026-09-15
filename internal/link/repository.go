package link

import (
	"context"
	"database/sql"
	"time"
)

type Repository interface {
	Create(ctx context.Context, slug, dest, title string, isCustom bool, exp *time.Time, redirectType, fallbackURL string) (Link, error)
	GetByID(ctx context.Context, id int64) (Link, bool, error)
	GetBySlug(ctx context.Context, slug string) (Link, bool, error)
	List(ctx context.Context) ([]Link, error)
	Update(ctx context.Context, id int64, slug, dest, title string, isCustom bool, exp *time.Time, redirectType, fallbackURL string) (Link, bool, error)
	Delete(ctx context.Context, id int64) (string, bool, error)
	Disable(ctx context.Context, id int64, fallbackURL *string) (Link, bool, error)
	Enable(ctx context.Context, id int64) (Link, bool, error)

	RecordClick(event ClickEvent)
	GetClickByID(ctx context.Context, id int64) (ClickEvent, bool, error)
	ListClicksByLinkID(ctx context.Context, linkID int64) ([]ClickEvent, error)
	Close()
}

type sqliteRepository struct {
	db     *sql.DB
	clicks *clickTracker
}

func NewRepository(database *sql.DB) Repository {
	return &sqliteRepository{
		db:     database,
		clicks: newClickTracker(database),
	}
}

func (r *sqliteRepository) Create(ctx context.Context, slug, dest, title string, isCustom bool, exp *time.Time, redirectType, fallbackURL string) (Link, error) {
	query := `
        INSERT INTO links (slug, destination_url, title, is_custom, expires_at, redirect_type, fallback_url)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        RETURNING id, slug, destination_url, title, is_custom, click_count, created_at, expires_at, status, disabled_at, enabled_at, redirect_type, fallback_url
    `
	var l Link
	err := r.db.QueryRowContext(ctx, query, slug, dest, title, isCustom, exp, redirectType, fallbackURL).Scan(
		&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt, &l.ExpiresAt,
		&l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL,
	)
	return l, err
}

func (r *sqliteRepository) GetByID(ctx context.Context, id int64) (Link, bool, error) {
	var l Link
	err := r.db.QueryRowContext(ctx, `
        SELECT id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
               status, disabled_at, enabled_at, redirect_type, fallback_url
        FROM links WHERE id = ?
    `, id).Scan(&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt,
		&l.ExpiresAt, &l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL)

	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	return l, err == nil, err
}

func (r *sqliteRepository) GetBySlug(ctx context.Context, slug string) (Link, bool, error) {
	var l Link
	err := r.db.QueryRowContext(ctx, `
        SELECT id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
               status, disabled_at, enabled_at, redirect_type, fallback_url
        FROM links WHERE slug = ?
    `, slug).Scan(&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt,
		&l.ExpiresAt, &l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL)

	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	return l, err == nil, err
}

func (r *sqliteRepository) List(ctx context.Context) ([]Link, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
               status, disabled_at, enabled_at, redirect_type, fallback_url
        FROM links ORDER BY id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt,
			&l.ExpiresAt, &l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *sqliteRepository) RecordClick(event ClickEvent) {
	r.clicks.record(event)
}

func (r *sqliteRepository) GetClickByID(ctx context.Context, id int64) (ClickEvent, bool, error) {
	var c ClickEvent
	err := r.db.QueryRowContext(ctx, `
        SELECT id, link_id, referer, user_agent, clicked_at FROM link_clicks WHERE id = ?
    `, id).Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt)

	if err == sql.ErrNoRows {
		return ClickEvent{}, false, nil
	}
	return c, err == nil, err
}

func (r *sqliteRepository) ListClicksByLinkID(ctx context.Context, linkID int64) ([]ClickEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, link_id, referer, user_agent, clicked_at FROM link_clicks
        WHERE link_id = ? ORDER BY clicked_at DESC
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

func (r *sqliteRepository) Close() {
	r.clicks.close()
}

func (r *sqliteRepository) Update(ctx context.Context, id int64, slug, dest, title string, isCustom bool, exp *time.Time, redirectType, fallbackURL string) (Link, bool, error) {
	query := `
		UPDATE links
		SET slug = ?, destination_url = ?, title = ?, is_custom = ?, expires_at = ?, redirect_type = ?, fallback_url = ?
		WHERE id = ?
		RETURNING id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
		          status, disabled_at, enabled_at, redirect_type, fallback_url
	`
	var l Link
	err := r.db.QueryRowContext(ctx, query, slug, dest, title, isCustom, exp, redirectType, fallbackURL, id).Scan(
		&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt, &l.ExpiresAt,
		&l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL,
	)
	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	return l, err == nil, err
}

func (r *sqliteRepository) Delete(ctx context.Context, id int64) (string, bool, error) {
	var slug string
	err := r.db.QueryRowContext(ctx, `DELETE FROM links WHERE id = ? RETURNING slug`, id).Scan(&slug)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return slug, err == nil, err
}

func (r *sqliteRepository) Disable(ctx context.Context, id int64, fallbackURL *string) (Link, bool, error) {
	if fallbackURL != nil && *fallbackURL != "" {
		query := `
			UPDATE links SET status = 'disabled', disabled_at = CURRENT_TIMESTAMP, fallback_url = ?
			WHERE id = ?
			RETURNING id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
			          status, disabled_at, enabled_at, redirect_type, fallback_url
		`
		var l Link
		err := r.db.QueryRowContext(ctx, query, *fallbackURL, id).Scan(
			&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt, &l.ExpiresAt,
			&l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL,
		)
		if err == sql.ErrNoRows {
			return Link{}, false, nil
		}
		return l, err == nil, err
	}

	query := `
		UPDATE links SET status = 'disabled', disabled_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
		          status, disabled_at, enabled_at, redirect_type, fallback_url
	`
	var l Link
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt, &l.ExpiresAt,
		&l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL,
	)
	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	return l, err == nil, err
}

func (r *sqliteRepository) Enable(ctx context.Context, id int64) (Link, bool, error) {
	query := `
		UPDATE links SET status = 'active', enabled_at = CURRENT_TIMESTAMP, disabled_at = NULL
		WHERE id = ?
		RETURNING id, slug, destination_url, title, is_custom, click_count, created_at, expires_at,
		          status, disabled_at, enabled_at, redirect_type, fallback_url
	`
	var l Link
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.Slug, &l.DestinationURL, &l.Title, &l.IsCustom, &l.ClickCount, &l.CreatedAt, &l.ExpiresAt,
		&l.Status, &l.DisabledAt, &l.EnabledAt, &l.RedirectType, &l.FallbackURL,
	)
	if err == sql.ErrNoRows {
		return Link{}, false, nil
	}
	return l, err == nil, err
}
