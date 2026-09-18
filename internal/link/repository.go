package link

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/url"
	"time"

	"github.com/oschwald/geoip2-golang"
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

	CreateQRCode(ctx context.Context, linkID int64, shortURL string) (QRCode, error)
	GetQRCodeByLinkID(ctx context.Context, linkID int64) (QRCode, bool, error)
	DeleteQRCodeByLinkID(ctx context.Context, linkID int64) error
	GetQRCodeByID(ctx context.Context, id int64) (QRCode, bool, error)

	RecordClick(event ClickEvent)
	GetClickByID(ctx context.Context, id int64) (ClickEvent, bool, error)
	ListClicksByLinkID(ctx context.Context, linkID int64) ([]ClickEvent, error)
	GetClicksByLinkIDWithFilters(ctx context.Context, linkID int64, filters map[string]any) ([]ClickEvent, error)
	GetAggregatedClicks(ctx context.Context, linkID int64, groupBy string, startDate, endDate *time.Time) ([]AggregatedClick, error)
	Close()
}

type sqliteRepository struct {
	db     *sql.DB
	clicks *clickTracker
}

func NewRepository(database *sql.DB, uaParser UserAgentParser, geoIP *geoip2.Reader) Repository {
	return &sqliteRepository{
		db:     database,
		clicks: newClickTracker(database, uaParser, geoIP),
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

func parseUTMParams(referer string) map[string]string {
	params := map[string]string{}
	u, err := url.Parse(referer)
	if err != nil {
		return params
	}
	q := u.Query()
	for _, key := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"} {
		if val := q.Get(key); val != "" {
			params[key] = val
		}
	}
	return params
}

func isQRScan(referer string) bool {
	u, err := url.Parse(referer)
	if err != nil {
		return false
	}
	q := u.Query()
	return q.Get("qr") != "" || q.Get("scan") != ""
}

func (r *sqliteRepository) CreateQRCode(ctx context.Context, linkID int64, shortURL string) (QRCode, error) {
	query := `
		INSERT INTO qr_codes (link_id, short_url)
		VALUES (?, ?)
		RETURNING id, link_id, short_url, created_at
	`
	var qc QRCode
	err := r.db.QueryRowContext(ctx, query, linkID, shortURL).Scan(
		&qc.ID, &qc.LinkID, &qc.ShortURL, &qc.CreatedAt,
	)
	return qc, err
}

func (r *sqliteRepository) GetQRCodeByLinkID(ctx context.Context, linkID int64) (QRCode, bool, error) {
	var qc QRCode
	err := r.db.QueryRowContext(ctx, `
		SELECT id, link_id, short_url, created_at FROM qr_codes WHERE link_id = ?
	`, linkID).Scan(&qc.ID, &qc.LinkID, &qc.ShortURL, &qc.CreatedAt)

	if err == sql.ErrNoRows {
		return QRCode{}, false, nil
	}
	return qc, err == nil, err
}

func (r *sqliteRepository) DeleteQRCodeByLinkID(ctx context.Context, linkID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM qr_codes WHERE link_id = ?`, linkID)
	return err
}

func (r *sqliteRepository) GetQRCodeByID(ctx context.Context, id int64) (QRCode, bool, error) {
	var qc QRCode
	err := r.db.QueryRowContext(ctx, `SELECT id, link_id, short_url, created_at FROM qr_codes WHERE id = ?`, id).Scan(&qc.ID, &qc.LinkID, &qc.ShortURL, &qc.CreatedAt)
	if err == sql.ErrNoRows {
		return QRCode{}, false, nil
	}
	return qc, err == nil, err
}

func (r *sqliteRepository) RecordClick(event ClickEvent) {
	ua := r.clicks.uaParser.Parse(event.UserAgent)
	event.Browser = ua.Name
	event.OS = ua.OS

	switch {
	case ua.Bot:
		event.Device = "Bot"
	case ua.Mobile:
		event.Device = "Mobile"
	case ua.Tablet:
		event.Device = "Tablet"
	case ua.Desktop:
		event.Device = "Desktop"
	default:
		event.Device = ua.Device
	}

	utmParams := parseUTMParams(event.Referer)
	event.UTMParams = utmParams

	if event.QRCodeID != nil {
		event.QRScan = true
	} else {
		event.QRScan = isQRScan(event.Referer)
	}

	if event.IP != "" && r.clicks.geoIP != nil {
		location, _ := r.clicks.geoIP.City(net.ParseIP(event.IP))
		event.Country = location.Country.Names["en"]
		event.Region = location.Subdivisions[0].Names["en"]
		event.City = location.City.Names["en"]
	}

	r.clicks.record(event)
}

func (r *sqliteRepository) GetClickByID(ctx context.Context, id int64) (ClickEvent, bool, error) {
	var c ClickEvent
	var utmParamsJSON []byte
	var qrCodeID *int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id, link_id, referer, user_agent, clicked_at, browser, os, device, country, region, city, utm_params, qr_scan, ip, qr_code_id FROM link_clicks WHERE id = ?
	`, id).Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt, &c.Browser, &c.OS, &c.Device, &c.Country, &c.Region, &c.City, &utmParamsJSON, &c.QRScan, &c.IP, &qrCodeID)

	if err == sql.ErrNoRows {
		return ClickEvent{}, false, nil
	}
	if utmParamsJSON != nil {
		json.Unmarshal(utmParamsJSON, &c.UTMParams)
	}
	if qrCodeID != nil {
		c.QRCodeID = qrCodeID
	}
	return c, err == nil, err
}

func (r *sqliteRepository) ListClicksByLinkID(ctx context.Context, linkID int64) ([]ClickEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, link_id, referer, user_agent, clicked_at, browser, os, device, country, region, city, utm_params, qr_scan, ip, qr_code_id FROM link_clicks
		WHERE link_id = ? ORDER BY clicked_at DESC
	`, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []ClickEvent
	for rows.Next() {
		var c ClickEvent
		var utmParamsJSON []byte
		var qrCodeID *int64
		if err := rows.Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt, &c.Browser, &c.OS, &c.Device, &c.Country, &c.Region, &c.City, &utmParamsJSON, &c.QRScan, &c.IP, &qrCodeID); err != nil {
			return nil, err
		}
		if utmParamsJSON != nil {
			json.Unmarshal(utmParamsJSON, &c.UTMParams)
		}
		if qrCodeID != nil {
			c.QRCodeID = qrCodeID
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

func (r *sqliteRepository) GetClicksByLinkIDWithFilters(ctx context.Context, linkID int64, filters map[string]interface{}) ([]ClickEvent, error) {
	query := `
        SELECT id, link_id, referer, user_agent, clicked_at, browser, os, device, country, region, city, utm_params, qr_scan, ip, qr_code_id
        FROM link_clicks
        WHERE link_id = ?
    `
	args := []interface{}{linkID}

	if browser, ok := filters["browser"].(string); ok && browser != "" {
		query += " AND browser = ?"
		args = append(args, browser)
	}
	if osName, ok := filters["os"].(string); ok && osName != "" {
		query += " AND os = ?"
		args = append(args, osName)
	}
	if country, ok := filters["country"].(string); ok && country != "" {
		query += " AND country = ?"
		args = append(args, country)
	}

	query += " ORDER BY clicked_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []ClickEvent
	for rows.Next() {
		var c ClickEvent
		var utmParamsJSON []byte
		var qrCodeID *int64
		if err := rows.Scan(&c.ID, &c.LinkID, &c.Referer, &c.UserAgent, &c.ClickedAt, &c.Browser, &c.OS, &c.Device, &c.Country, &c.Region, &c.City, &utmParamsJSON, &c.QRScan, &c.IP, &qrCodeID); err != nil {
			return nil, err
		}
		if utmParamsJSON != nil {
			_ = json.Unmarshal(utmParamsJSON, &c.UTMParams)
		}
		if qrCodeID != nil {
			c.QRCodeID = qrCodeID
		}
		clicks = append(clicks, c)
	}
	return clicks, rows.Err()
}

func (r *sqliteRepository) GetAggregatedClicks(ctx context.Context, linkID int64, groupBy string, startDate, endDate *time.Time) ([]AggregatedClick, error) {
	var dateFormat string
	switch groupBy {
	case "hour":
		dateFormat = "%Y-%m-%d %H:00:00"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	query := `
        SELECT strftime(?, clicked_at) AS period, COUNT(*) as count
        FROM link_clicks
        WHERE link_id = ?
    `
	args := []any{dateFormat, linkID}

	if startDate != nil {
		query += " AND clicked_at >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND clicked_at <= ?"
		args = append(args, *endDate)
	}

	query += " GROUP BY period ORDER BY period ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AggregatedClick
	for rows.Next() {
		var agg AggregatedClick
		if err := rows.Scan(&agg.Date, &agg.Count); err != nil {
			return nil, err
		}
		results = append(results, agg)
	}
	return results, rows.Err()
}
