package link

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func setupTestDB(t *testing.T) *sql.DB {
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

func TestClickTracker_ConcurrencyAndShutdown(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tracker := newClickTracker(db)

	const numWorkers = 10
	const clicksPerWorker = 200

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	// Spin up workers sending events concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal // Sync start for max concurrency

			for j := 0; j < clicksPerWorker; j++ {
				tracker.record(ClickEvent{
					LinkID:    1,
					Referer:   fmt.Sprintf("http://example.com/%d", workerID),
					UserAgent: "TestAgent/1.0",
					ClickedAt: time.Now(),
				})
			}
		}(i)
	}

	close(startSignal) // Trigger workers

	// Trigger shutdown concurrently while workers are sending
	time.Sleep(5 * time.Millisecond)
	tracker.close()

	// Ensure no goroutine deadlocks or panics on closed channels
	wg.Wait()

	// Verify idempotency on repeated close calls
	tracker.close()
}
