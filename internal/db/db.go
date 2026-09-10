package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Open(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)

	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf(
				"creating database directory '%s': %w",
				dir,
				err,
			)
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
			db.Close()
			return nil, fmt.Errorf(
				"setting pragma '%s': %w",
				pragma,
				err,
			)
		}
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			version INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("creating schema_version table: %w", err)
	}

	var version int

	err := db.QueryRow(`
		SELECT version
		FROM schema_version
		LIMIT 1
	`).Scan(&version)

	if err == sql.ErrNoRows {
		version = 0
	} else if err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	type migration struct {
		version  int
		filename string
	}

	var pending []migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf(
				"invalid migration filename: %s",
				entry.Name(),
			)
		}

		migrationVersion, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf(
				"invalid migration version in %s: %w",
				entry.Name(),
				err,
			)
		}

		if migrationVersion > version {
			pending = append(pending, migration{
				version:  migrationVersion,
				filename: entry.Name(),
			})
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].version < pending[j].version
	})

	for _, migration := range pending {
		if err := runMigration(
			db,
			migration.version,
			migration.filename,
		); err != nil {
			return err
		}

		version = migration.version
	}

	return nil
}

func runMigration(db *sql.DB, version int, filename string) error {
	sqlBytes, err := migrations.ReadFile(
		"migrations/" + filename,
	)
	if err != nil {
		return fmt.Errorf(
			"reading migration %s: %w",
			filename,
			err,
		)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf(
			"starting migration %d: %w",
			version,
			err,
		)
	}

	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf(
			"running migration %d (%s): %w",
			version,
			filename,
			err,
		)
	}

	_, err = tx.Exec(`
		INSERT INTO schema_version (id, version)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET version = excluded.version
	`, version)

	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf(
			"recording migration %d: %w",
			version,
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"committing migration %d: %w",
			version,
			err,
		)
	}

	return nil
}
