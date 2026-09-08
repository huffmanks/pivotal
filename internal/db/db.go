package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func Open(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating database directory '%s': %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("setting pragma '%s': %w", pragma, err)
		}
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE CHECK(length(slug) <= 255),
		destination_url TEXT NOT NULL,
		is_custom INTEGER NOT NULL DEFAULT 0,
		click_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug ON links(slug);

	CREATE TABLE IF NOT EXISTS link_clicks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
		referer TEXT,
		user_agent TEXT,
		clicked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schema)
	return err
}
