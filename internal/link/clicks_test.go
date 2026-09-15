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

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal

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

	close(startSignal)

	time.Sleep(5 * time.Millisecond)
	tracker.close()

	wg.Wait()

	tracker.close()
}
