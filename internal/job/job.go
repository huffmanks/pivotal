package job

import (
	"context"
	"log"
	"time"
)

type Manager struct {
	queue      *Queue
	scheduler  *Scheduler
	executor   *JobExecutor
	repository Repository
}

func NewManager(repository Repository, executor *JobExecutor, workers int) *Manager {
	q := NewQueue(repository, workers, executor.Handle)
	return &Manager{
		repository: repository,
		executor:   executor,
		queue:      q,
	}
}

func (m *Manager) Start() {
	m.queue.Start()

	m.scheduler = NewScheduler(m.queue, 1*time.Minute)
	m.scheduler.Start()

	log.Println("job manager started")
}

func (m *Manager) Stop() {
	if m.scheduler != nil {
		m.scheduler.Stop()
	}
	m.queue.Stop()
	m.executor.Close()
	log.Println("job manager stopped")
}

func (m *Manager) Enqueue(ctx context.Context, job *Job) (Job, error) {
	return m.queue.Enqueue(ctx, job)
}

func (m *Manager) GetJob(ctx context.Context, id int64) (Job, bool, error) {
	return m.repository.GetByID(ctx, id)
}

func (m *Manager) Notifications() <-chan JobNotification {
	return m.executor.Notifications()
}

func (m *Manager) IsRunning() bool {
	return m.queue.IsRunning()
}
