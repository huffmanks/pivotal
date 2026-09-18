package link

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestClickTracker_ConcurrencyAndShutdown(t *testing.T) {
	db := setupClickDB(t)
	defer db.Close()

	tracker := newClickTracker(db)

	const numWorkers = 10
	const clicksPerWorker = 200

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for workerID := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal

			for range clicksPerWorker {
				tracker.record(ClickEvent{
					LinkID:    1,
					Referer:   fmt.Sprintf("http://example.com/%d", workerID),
					UserAgent: "TestAgent/1.0",
					ClickedAt: time.Now(),
				})
			}
		}(workerID)
	}

	close(startSignal)

	time.Sleep(5 * time.Millisecond)
	tracker.close()

	wg.Wait()

	tracker.close()
}
