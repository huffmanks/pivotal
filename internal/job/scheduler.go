package job

import (
	"context"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	queue    *Queue
	interval time.Duration
	stop     chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

func NewScheduler(queue *Queue, interval time.Duration) *Scheduler {
	return &Scheduler{
		queue:    queue,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}
	s.running = true

	s.wg.Add(1)
	go s.run()
	log.Printf("scheduler started with interval %s", s.interval)
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.scheduleJobs()
		}
	}
}

func (s *Scheduler) scheduleJobs() {
	now := time.Now().UTC()

	expireJob := &Job{
		Type:      JobTypeExpireLinks,
		Payload:   "check_expired_links",
		Status:    JobStatusPending,
		Priority:  5,
		MaxRetries: 3,
		DedupKey:  "expire_links_periodic",
	}
	if _, err := s.queue.Enqueue(context.Background(), expireJob); err != nil {
		log.Printf("scheduler: failed to enqueue expire_links: %v", err)
	}

	healthJob := &Job{
		Type:      JobTypeHealthCheck,
		Payload:   "check_all_links_health",
		Status:    JobStatusPending,
		Priority:  3,
		MaxRetries: 2,
		DedupKey:  "health_check_periodic",
		ScheduledAt: &now,
	}
	if _, err := s.queue.Enqueue(context.Background(), healthJob); err != nil {
		log.Printf("scheduler: failed to enqueue health_check: %v", err)
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	close(s.stop)
	s.wg.Wait()
	s.running = false
	log.Println("scheduler stopped")
}
