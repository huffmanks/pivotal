package job

import (
	"context"
	"database/sql"
	"time"
)

type Repository interface {
	Enqueue(ctx context.Context, job *Job) (Job, error)
	GetNextPending(ctx context.Context) (*Job, bool, error)
	GetByID(ctx context.Context, id int64) (Job, bool, error)
	UpdateStatus(ctx context.Context, id int64, status JobStatus, errorMsg *string) error
	MarkProcessing(ctx context.Context, id int64) error
	MarkCompleted(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	IncrementRetry(ctx context.Context, id int64) error
	ListByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error)
	CleanupCompleted(ctx context.Context, olderThan time.Time) error
}

type sqliteRepository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) Repository {
	return &sqliteRepository{db: database}
}

func (s *sqliteRepository) Enqueue(ctx context.Context, job *Job) (Job, error) {
	if job.Status == "" {
		job.Status = JobStatusPending
	}
	if job.Priority == 0 {
		job.Priority = 5
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	if job.DedupKey != "" {
		query := `
			INSERT INTO jobs (type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(type, dedup_key) DO UPDATE SET
				payload = excluded.payload,
				priority = excluded.priority,
				max_retries = excluded.max_retries,
				scheduled_at = excluded.scheduled_at,
				status = CASE WHEN excluded.status = 'pending' THEN jobs.status ELSE excluded.status END,
				retry_count = CASE WHEN excluded.status = 'pending' THEN jobs.retry_count ELSE excluded.retry_count END
			RETURNING id, type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at, started_at, completed_at, failed_at, error_message, created_at
		`
		var j Job
		err := s.db.QueryRowContext(ctx, query,
			job.Type, job.Payload, job.Status, job.Priority, job.MaxRetries, job.RetryCount,
			job.DedupKey, job.ScheduledAt,
		).Scan(
			&j.ID, &j.Type, &j.Payload, &j.Status, &j.Priority, &j.MaxRetries, &j.RetryCount,
			&j.DedupKey, &j.ScheduledAt, &j.StartedAt, &j.CompletedAt, &j.FailedAt,
			&j.ErrorMessage, &j.CreatedAt,
		)
		return j, err
	}

	query := `
		INSERT INTO jobs (type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at, started_at, completed_at, failed_at, error_message, created_at
	`
	var j Job
	dedupVal := interface{}(nil)
	if job.DedupKey != "" {
		dedupVal = job.DedupKey
	}
	err := s.db.QueryRowContext(ctx, query,
		job.Type, job.Payload, job.Status, job.Priority, job.MaxRetries, job.RetryCount,
		dedupVal, job.ScheduledAt,
	).Scan(
		&j.ID, &j.Type, &j.Payload, &j.Status, &j.Priority, &j.MaxRetries, &j.RetryCount,
		&j.DedupKey, &j.ScheduledAt, &j.StartedAt, &j.CompletedAt, &j.FailedAt,
		&j.ErrorMessage, &j.CreatedAt,
	)
	return j, err
}

func (s *sqliteRepository) GetNextPending(ctx context.Context) (*Job, bool, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	query := `
		SELECT id, type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at, started_at, completed_at, failed_at, error_message, created_at
		FROM jobs
	WHERE (status = 'pending' OR status = 'failed')
	  AND (scheduled_at IS NULL OR scheduled_at <= ?)
		  AND (retry_count < max_retries OR max_retries = 0)
		ORDER BY CASE WHEN status = 'failed' THEN 0 ELSE 1 END, priority DESC, created_at ASC
		LIMIT 1
	`
	var j Job
	err := s.db.QueryRowContext(ctx, query, now).Scan(
		&j.ID, &j.Type, &j.Payload, &j.Status, &j.Priority, &j.MaxRetries, &j.RetryCount,
		&j.DedupKey, &j.ScheduledAt, &j.StartedAt, &j.CompletedAt, &j.FailedAt,
		&j.ErrorMessage, &j.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &j, true, nil
}

func (s *sqliteRepository) GetByID(ctx context.Context, id int64) (Job, bool, error) {
	query := `
		SELECT id, type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at, started_at, completed_at, failed_at, error_message, created_at
		FROM jobs WHERE id = ?
	`
	var j Job
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID, &j.Type, &j.Payload, &j.Status, &j.Priority, &j.MaxRetries, &j.RetryCount,
		&j.DedupKey, &j.ScheduledAt, &j.StartedAt, &j.CompletedAt, &j.FailedAt,
		&j.ErrorMessage, &j.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return Job{}, false, nil
	}
	return j, err == nil, err
}

func (s *sqliteRepository) UpdateStatus(ctx context.Context, id int64, status JobStatus, errorMsg *string) error {
	if errorMsg != nil {
		_, err := s.db.ExecContext(ctx, `
			UPDATE jobs SET status = ?, error_message = ?, completed_at = CASE WHEN ? = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END, failed_at = CASE WHEN ? = 'failed' THEN CURRENT_TIMESTAMP ELSE failed_at END
			WHERE id = ?
		`, status, *errorMsg, status, status, id)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs SET status = ?, completed_at = CASE WHEN ? = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END, failed_at = CASE WHEN ? = 'failed' THEN CURRENT_TIMESTAMP ELSE failed_at END
		WHERE id = ?
	`, status, status, status, id)
	return err
}

func (s *sqliteRepository) MarkProcessing(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs SET status = 'processing', started_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (status = 'pending' OR status = 'failed')
	`, id)
	return err
}

func (s *sqliteRepository) MarkCompleted(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs SET status = 'completed', completed_at = CURRENT_TIMESTAMP, retry_count = 0
		WHERE id = ?
	`, id)
	return err
}

func (s *sqliteRepository) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs SET status = 'failed', failed_at = CURRENT_TIMESTAMP, error_message = ?
		WHERE id = ?
	`, errMsg, id)
	return err
}

func (s *sqliteRepository) IncrementRetry(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs SET retry_count = retry_count + 1, status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END
		WHERE id = ?
	`, id)
	return err
}

func (s *sqliteRepository) ListByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error) {
	query := `
		SELECT id, type, payload, status, priority, max_retries, retry_count, dedup_key, scheduled_at, started_at, completed_at, failed_at, error_message, created_at
		FROM jobs WHERE status = ? ORDER BY created_at DESC LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID, &j.Type, &j.Payload, &j.Status, &j.Priority, &j.MaxRetries, &j.RetryCount,
			&j.DedupKey, &j.ScheduledAt, &j.StartedAt, &j.CompletedAt, &j.FailedAt,
			&j.ErrorMessage, &j.CreatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (s *sqliteRepository) CleanupCompleted(ctx context.Context, olderThan time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE status = 'completed' AND completed_at < ?`, olderThan.Format("2006-01-02 15:04:05"))
	return err
}
