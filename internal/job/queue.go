package job

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrJobNotFound      = errors.New("job not found")
	ErrJobAlreadyInWork = errors.New("job already being processed")
	ErrWorkerStopped    = errors.New("worker stopped")
)

type Queue struct {
	repository Repository
	workers    int
	handler    JobHandler
	jobCtx     context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
	processor  *jobProcessor
}

type jobProcessor struct {
	queue *Queue
}

func NewQueue(repository Repository, workers int, handler JobHandler) *Queue {
	if workers < 1 {
		workers = 1
	}
	return &Queue{
		repository: repository,
		workers:    workers,
		handler:    handler,
	}
}

func (q *Queue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return
	}

	q.jobCtx, q.cancel = context.WithCancel(context.Background())
	q.running = true
	q.processor = &jobProcessor{queue: q}

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	log.Printf("job queue started with %d workers", q.workers)
}

func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	q.cancel()
	q.wg.Wait()
	q.running = false
	log.Println("job queue stopped")
}

func (q *Queue) worker(id int) {
	defer q.wg.Done()
	log.Printf("worker %d started", id)

	for {
		select {
		case <-q.jobCtx.Done():
			log.Printf("worker %d stopping", id)
			return
		default:
		}

		job, found, err := q.repository.GetNextPending(q.jobCtx)
		if err != nil {
			log.Printf("worker %d: error fetching job: %v", id, err)
			select {
			case <-q.jobCtx.Done():
				return
			case <-time.After(1 * time.Second):
			}
			continue
		}
		if !found {
			select {
			case <-q.jobCtx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}

		if err := q.repository.MarkProcessing(q.jobCtx, job.ID); err != nil {
			log.Printf("worker %d: failed to mark job %d as processing: %v", id, job.ID, err)
			select {
			case <-q.jobCtx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}

		q.processJob(id, job)
	}
}

func (q *Queue) processJob(workerID int, job *Job) {
	log.Printf("worker %d: processing job %d of type %s", workerID, job.ID, job.Type)

	result := q.handler(q.jobCtx, job)

	if result.Success {
		if err := q.repository.MarkCompleted(q.jobCtx, job.ID); err != nil {
			log.Printf("worker %d: failed to mark job %d completed: %v", workerID, job.ID, err)
		} else {
			log.Printf("worker %d: job %d completed", workerID, job.ID)
		}
	} else {
		if job.RetryCount < job.MaxRetries && result.Retryable {
			if err := q.repository.IncrementRetry(q.jobCtx, job.ID); err != nil {
				log.Printf("worker %d: failed to increment retry for job %d: %v", workerID, job.ID, err)
			} else {
				log.Printf("worker %d: job %d failed, will retry (%d/%d): %v", workerID, job.ID, job.RetryCount+1, job.MaxRetries, result.Error)
			}
		} else {
			errMsg := result.Error
			if errMsg == "" {
				errMsg = "job failed"
			}
			if err := q.repository.MarkFailed(q.jobCtx, job.ID, errMsg); err != nil {
				log.Printf("worker %d: failed to mark job %d as failed: %v", workerID, job.ID, err)
			} else {
				log.Printf("worker %d: job %d failed permanently: %v", workerID, job.ID, result.Error)
			}
		}
	}
}

func (q *Queue) Enqueue(ctx context.Context, job *Job) (Job, error) {
	return q.repository.Enqueue(ctx, job)
}

func (q *Queue) IsRunning() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.running
}

func (q *Queue) Wait() {
	q.wg.Wait()
}

func (q *Queue) Context() context.Context {
	return q.jobCtx
}
