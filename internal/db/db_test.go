package db

import (
	"testing"
)

func TestDB_MigrationIdempotency(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	v1, err := currentSchemaVersion(db)
	if err != nil {
		t.Fatalf("failed to read schema version: %v", err)
	}
	if v1 == 0 {
		t.Errorf("expected schema version > 0, got %d", v1)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("subsequent migration failed: %v", err)
	}

	v2, err := currentSchemaVersion(db)
	if err != nil {
		t.Fatalf("failed to re-read schema version: %v", err)
	}
	if v1 != v2 {
		t.Errorf("schema version changed unexpectedly: before %d, after %d", v1, v2)
	}

	expectedColumns := map[string]bool{
		"id":              false,
		"slug":            false,
		"destination_url": false,
		"is_custom":       false,
		"click_count":     false,
		"created_at":      false,
		"expires_at":      false,
		"status":          false,
		"disabled_at":     false,
		"enabled_at":      false,
		"redirect_type":   false,
		"fallback_url":    false,
	}

	rows, err := db.Query("PRAGMA table_info(links);")
	if err != nil {
		t.Fatalf("failed to inspect links table schema: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typeStr string
		var notNull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &typeStr, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed scanning column info: %v", err)
		}
		if _, ok := expectedColumns[name]; ok {
			expectedColumns[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("error iterating table info rows: %v", err)
	}

	for name, found := range expectedColumns {
		if !found {
			t.Errorf("expected column %q in links table", name)
		}
	}

	expectedClickColumns := map[string]bool{
		"id":         false,
		"link_id":    false,
		"referer":    false,
		"user_agent": false,
		"clicked_at": false,
		"browser":    false,
		"os":         false,
		"device":     false,
		"country":    false,
		"region":     false,
		"city":       false,
		"utm_params": false,
		"qr_scan":    false,
		"ip":         false,
		"qr_code_id": false,
	}

	rows2, err := db.Query("PRAGMA table_info(link_clicks);")
	if err != nil {
		t.Fatalf("failed to inspect link_clicks table schema: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var cid int
		var name, typeStr string
		var notNull, pk int
		var dfltValue any
		if err := rows2.Scan(&cid, &name, &typeStr, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed scanning column info: %v", err)
		}
		if _, ok := expectedClickColumns[name]; ok {
			expectedClickColumns[name] = true
		}
	}
	if err := rows2.Err(); err != nil {
		t.Fatalf("error iterating table info rows: %v", err)
	}

	for name, found := range expectedClickColumns {
		if !found {
			t.Errorf("expected column %q in link_clicks table", name)
		}
	}

	_, err = db.Query("PRAGMA table_info(qr_codes);")
	if err != nil {
		t.Fatalf("failed to inspect qr_codes table schema: %v", err)
	}
}
