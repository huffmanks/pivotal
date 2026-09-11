package db

import (
	"testing"
)

func TestDB_MigrationIdempotency(t *testing.T) {
	// 1. Open database (triggers initial migration)
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// 2. Verify schema version after first run
	v1, err := currentSchemaVersion(db)
	if err != nil {
		t.Fatalf("failed to read schema version: %v", err)
	}
	if v1 == 0 {
		t.Errorf("expected schema version > 0, got %d", v1)
	}

	// 3. Execute Migrate directly again to ensure safety/idempotency
	if err := Migrate(db); err != nil {
		t.Fatalf("subsequent migration failed: %v", err)
	}

	// 4. Confirm schema version remained unchanged
	v2, err := currentSchemaVersion(db)
	if err != nil {
		t.Fatalf("failed to re-read schema version: %v", err)
	}
	if v1 != v2 {
		t.Errorf("schema version changed unexpectedly: before %d, after %d", v1, v2)
	}

	// 5. Query expected table columns to verify schema state
	var expiresAtExists bool
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
		if name == "expires_at" {
			expiresAtExists = true
		}
	}

	if !expiresAtExists {
		t.Error("expected 'expires_at' column in 'links' table from migration 002")
	}
}
