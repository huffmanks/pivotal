package link

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

const (
	clickBatchSize     = 100
	clickBufferSize    = 2000
	clickFlushInterval = 2 * time.Second
)

type clickTracker struct {
	db     *sql.DB
	events chan ClickEvent
	mu     sync.RWMutex
	wg     sync.WaitGroup
	closed bool
}

func newClickTracker(db *sql.DB) *clickTracker {
	tracker := &clickTracker{
		db:     db,
		events: make(chan ClickEvent, clickBufferSize),
	}

	tracker.wg.Add(1)
	go tracker.process()

	return tracker
}

func (t *clickTracker) record(event ClickEvent) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return
	}

	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()

	select {
	case t.events <- event:
	case <-timer.C:
		log.Printf("warning: click tracker buffer full, dropped click for link_id: %d", event.LinkID)
	}
}

func (t *clickTracker) close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	close(t.events)
	t.mu.Unlock()

	t.wg.Wait()
}

func (t *clickTracker) process() {
	defer t.wg.Done()

	ticker := time.NewTicker(clickFlushInterval)
	defer ticker.Stop()

	batch := make([]ClickEvent, 0, clickBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := t.flush(batch); err != nil {
			log.Printf("error flushing clicks: %v", err)
			return
		}
		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-t.events:
			if !ok {
				flush()
				return
			}
			batch = append(batch, event)
			if len(batch) >= clickBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (t *clickTracker) flush(batch []ClickEvent) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, click := range batch {
		if _, err := tx.Exec(`
            INSERT INTO link_clicks (link_id, referer, user_agent, clicked_at)
            VALUES (?, ?, ?, ?)
        `, click.LinkID, click.Referer, click.UserAgent, click.ClickedAt); err != nil {
			return err
		}

		if _, err := tx.Exec(`
            UPDATE links SET click_count = click_count + 1 WHERE id = ?
        `, click.LinkID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
