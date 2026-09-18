package link

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func TestLinkService_Create(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if link.DestinationURL != "https://example.com" {
		t.Errorf("expected destination URL %q, got %q", "https://example.com", link.DestinationURL)
	}
	if link.Status != LinkStatusActive {
		t.Errorf("expected status %q, got %q", LinkStatusActive, link.Status)
	}
	if link.RedirectType != "302" {
		t.Errorf("expected redirect type 302, got %q", link.RedirectType)
	}

	link2, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com/about", Title: "About Page"})
	if err != nil {
		t.Fatalf("Create with title failed: %v", err)
	}
	if link2.Title != "About Page" {
		t.Errorf("expected title %q, got %q", "About Page", link2.Title)
	}

	link3, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		RedirectType:   new("301"),
	})
	if err != nil {
		t.Fatalf("Create with 301 failed: %v", err)
	}
	if link3.RedirectType != "301" {
		t.Errorf("expected redirect type 301, got %q", link3.RedirectType)
	}

	link4, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		FallbackURL:    new("https://fallback.com"),
	})
	if err != nil {
		t.Fatalf("Create with fallback failed: %v", err)
	}
	if link4.FallbackURL != "https://fallback.com" {
		t.Errorf("expected fallback URL %q, got %q", "https://fallback.com", link4.FallbackURL)
	}

	_, err = svc.Create(ctx, CreateLinkRequest{DestinationURL: "not-a-url"})
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestLinkService_Update(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	updated, err := svc.Update(ctx, link.ID, UpdateLinkRequest{DestinationURL: new("https://newdestination.com")})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.DestinationURL != "https://newdestination.com" {
		t.Errorf("expected destination %q, got %q", "https://newdestination.com", updated.DestinationURL)
	}

	updated, err = svc.Update(ctx, link.ID, UpdateLinkRequest{Title: new("New Title")})
	if err != nil {
		t.Fatalf("Update title failed: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title %q, got %q", "New Title", updated.Title)
	}

	updated, err = svc.Update(ctx, link.ID, UpdateLinkRequest{RedirectType: new("301")})
	if err != nil {
		t.Fatalf("Update redirect type failed: %v", err)
	}
	if updated.RedirectType != "301" {
		t.Errorf("expected redirect type 301, got %q", updated.RedirectType)
	}

	updated, err = svc.Update(ctx, link.ID, UpdateLinkRequest{FallbackURL: new("https://fallback.com")})
	if err != nil {
		t.Fatalf("Update fallback failed: %v", err)
	}
	if updated.FallbackURL != "https://fallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://fallback.com", updated.FallbackURL)
	}
}

func TestLinkService_DisableEnable(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	disabled, err := svc.Disable(ctx, link.ID, nil)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
	if disabled.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, disabled.Status)
	}

	disabled2, err := svc.Disable(ctx, link.ID, new("https://fallback.com"))
	if err != nil {
		t.Fatalf("Disable with fallback failed: %v", err)
	}
	if disabled2.FallbackURL != "https://fallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://fallback.com", disabled2.FallbackURL)
	}
	if disabled2.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, disabled2.Status)
	}

	enabled, err := svc.Enable(ctx, link.ID)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
	if enabled.Status != LinkStatusActive {
		t.Errorf("expected status %q, got %q", LinkStatusActive, enabled.Status)
	}
}

func TestLinkService_Resolve(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	activeLink, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	resolved, _, err := svc.Resolve(ctx, activeLink.Slug)
	if err != nil {
		t.Fatalf("Resolve active link failed: %v", err)
	}
	if resolved.Status != LinkStatusActive {
		t.Errorf("expected status %q, got %q", LinkStatusActive, resolved.Status)
	}

	svc.Disable(ctx, activeLink.ID, nil)
	resolved, _, err = svc.Resolve(ctx, activeLink.Slug)
	if err != nil {
		t.Fatalf("Resolve disabled link failed: %v", err)
	}
	if resolved.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, resolved.Status)
	}
}

func TestLinkService_ResolveWithRedirect(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link301, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		RedirectType:   new("301"),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	l, dest, code, err := svc.ResolveWithRedirect(ctx, link301.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect failed: %v", err)
	}
	if code != httpStatusMovedPermanently {
		t.Errorf("expected 301, got %d", code)
	}
	if l.ID != link301.ID {
		t.Errorf("expected link %d, got %d", link301.ID, l.ID)
	}
	if dest != "https://example.com" {
		t.Errorf("expected destination %q, got %q", "https://example.com", dest)
	}

	link302, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	l, dest, code, err = svc.ResolveWithRedirect(ctx, link302.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect failed: %v", err)
	}
	if code != httpStatusFound {
		t.Errorf("expected 302, got %d", code)
	}
	if dest != "https://example.com" {
		t.Errorf("expected destination %q, got %q", "https://example.com", dest)
	}

	svc.Disable(ctx, link302.ID, new("https://fallback.com"))
	l, dest, code, err = svc.ResolveWithRedirect(ctx, link302.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect disabled+fallback failed: %v", err)
	}
	if code != httpStatusFound {
		t.Errorf("expected 302, got %d", code)
	}
	if dest != "https://fallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://fallback.com", dest)
	}

	svc.Disable(ctx, link301.ID, nil)

	svc.Update(ctx, link301.ID, UpdateLinkRequest{FallbackURL: new("https://fallback.com")})
	svc.Update(ctx, link301.ID, UpdateLinkRequest{FallbackURL: new("https://fallback.com")})
	l, dest, code, err = svc.ResolveWithRedirect(ctx, link301.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect disabled with fallback failed: %v", err)
	}
	if code != httpStatusMovedPermanently {
		t.Errorf("expected 301, got %d", code)
	}
	if dest != "https://fallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://fallback.com", dest)
	}
	if l.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, l.Status)
	}
}

func TestLinkService_Delete(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := svc.Delete(ctx, link.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = svc.GetByID(ctx, link.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestLinkService_Expiry(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	expiredTime := time.Now().Add(-time.Hour)
	link, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		ExpiresAt:      &expiredTime,
	})
	if err != nil {
		t.Fatalf("Create expired link failed: %v", err)
	}

	got, err := svc.GetByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != LinkStatusExpired {
		t.Errorf("expected status %q for expired link, got %q", LinkStatusExpired, got.Status)
	}
}

func TestLinkService_ResolveWithRedirect_ExpiredWithoutFallback(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	expiredTime := time.Now().Add(-time.Hour)
	link, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		ExpiresAt:      &expiredTime,
	})
	if err != nil {
		t.Fatalf("Create expired link failed: %v", err)
	}

	l, dest, code, err := svc.ResolveWithRedirect(ctx, link.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect expired without fallback failed: %v", err)
	}
	if l.Status != LinkStatusExpired {
		t.Errorf("expected status %q, got %q", LinkStatusExpired, l.Status)
	}
	if dest != "https://example.com" {
		t.Errorf("expected destination %q, got %q", "https://example.com", dest)
	}
	if code != httpStatusFound {
		t.Errorf("expected 302, got %d", code)
	}
}

func TestLinkService_ResolveWithRedirect_ExpiredWithFallback(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	expiredTime := time.Now().Add(-time.Hour)
	link, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		ExpiresAt:      &expiredTime,
		FallbackURL:    new("https://fallback.com"),
	})
	if err != nil {
		t.Fatalf("Create expired link failed: %v", err)
	}

	l, dest, _, err := svc.ResolveWithRedirect(ctx, link.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect expired with fallback failed: %v", err)
	}
	if l.Status != LinkStatusExpired {
		t.Errorf("expected status %q, got %q", LinkStatusExpired, l.Status)
	}
	if dest != "https://fallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://fallback.com", dest)
	}
}

func TestLinkService_ResolveWithRedirect_DisabledWithoutFallback(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{DestinationURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	svc.Disable(ctx, link.ID, nil)

	l, dest, code, err := svc.ResolveWithRedirect(ctx, link.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect disabled without fallback failed: %v", err)
	}
	if l.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, l.Status)
	}
	if dest != "https://example.com" {
		t.Errorf("expected destination %q, got %q", "https://example.com", dest)
	}
	if code != httpStatusFound {
		t.Errorf("expected 302, got %d", code)
	}
}

func TestLinkService_ResolveWithRedirect_DisabledWithConfiguredFallback(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	link, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		FallbackURL:    new("https://myfallback.com"),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	svc.Disable(ctx, link.ID, nil)

	l, dest, _, err := svc.ResolveWithRedirect(ctx, link.Slug)
	if err != nil {
		t.Fatalf("ResolveWithRedirect disabled with configured fallback failed: %v", err)
	}
	if l.Status != LinkStatusDisabled {
		t.Errorf("expected status %q, got %q", LinkStatusDisabled, l.Status)
	}
	if dest != "https://myfallback.com" {
		t.Errorf("expected fallback %q, got %q", "https://myfallback.com", dest)
	}
}

func TestLinkService_Resolve_ExpiredLink(t *testing.T) {
	db, svc := setupFullTest(t)
	defer db.Close()
	defer svc.Close()

	ctx := context.Background()

	expiredTime := time.Now().Add(-time.Hour)
	link, err := svc.Create(ctx, CreateLinkRequest{
		DestinationURL: "https://example.com",
		ExpiresAt:      &expiredTime,
	})
	if err != nil {
		t.Fatalf("Create expired link failed: %v", err)
	}

	resolved, _, err := svc.Resolve(ctx, link.Slug)
	if err != nil {
		t.Fatalf("Resolve expired link failed: %v", err)
	}
	if resolved.Status != LinkStatusExpired {
		t.Errorf("expected status %q, got %q", LinkStatusExpired, resolved.Status)
	}
}

func setupClickDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE links (id INTEGER PRIMARY KEY, click_count INTEGER DEFAULT 0);
		CREATE TABLE link_clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link_id INTEGER NOT NULL,
			referer TEXT,
			user_agent TEXT,
			clicked_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO links (id, click_count) VALUES (1, 0);
	`)
	if err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	return db
}

func setupFullTest(t *testing.T) (*sql.DB, *service) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE CHECK(length(slug) <= 255),
			destination_url TEXT NOT NULL,
			title TEXT,
			is_custom INTEGER NOT NULL DEFAULT 0,
			click_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'expired')),
			disabled_at DATETIME,
			enabled_at DATETIME,
			redirect_type TEXT NOT NULL DEFAULT '302' CHECK (redirect_type IN ('301', '302')),
			fallback_url TEXT
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug ON links(slug);
		CREATE TABLE IF NOT EXISTS link_clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
			referer TEXT,
			user_agent TEXT,
			clicked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	repo := NewRepository(db)
	cache, _ := lru.New[string, Link](100)
	svc := &service{repo: repo, cache: cache}

	return db, svc
}
