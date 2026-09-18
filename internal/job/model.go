package job

import (
	"context"
	"time"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type JobType string

const (
	JobTypeExpireLinks             JobType = "expire_links"
	JobTypeEnforceVisitLimits      JobType = "enforce_visit_limits"
	JobTypeEnforceUniqueVisitorLimits JobType = "enforce_unique_visitor_limits"
	JobTypeEnforceQRScanLimits     JobType = "enforce_qr_scan_limits"
	JobTypeHealthCheck             JobType = "health_check"
	JobTypeRefreshMetadata         JobType = "refresh_metadata"
	JobTypeProcessRoutingChanges   JobType = "process_routing_changes"
	JobTypeSendNotification        JobType = "send_notification"
)

type Job struct {
	ID           int64     `json:"id" db:"id"`
	Type         JobType   `json:"type" db:"type"`
	Payload      string    `json:"payload" db:"payload"`
	Status       JobStatus `json:"status" db:"status"`
	Priority     int       `json:"priority" db:"priority"`
	MaxRetries   int       `json:"max_retries" db:"max_retries"`
	RetryCount   int       `json:"retry_count" db:"retry_count"`
	DedupKey     string    `json:"dedup_key,omitempty" db:"dedup_key"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt    *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	FailedAt     *time.Time `json:"failed_at,omitempty" db:"failed_at"`
ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type JobResult struct {
	JobID     int64
	Success   bool
	Retryable bool
	Error     string
}

type JobHandler func(ctx context.Context, job *Job) JobResult

type JobOption func(*Job)

func WithPriority(p int) JobOption {
	return func(j *Job) {
		j.Priority = p
	}
}

func WithMaxRetries(r int) JobOption {
	return func(j *Job) {
		j.MaxRetries = r
	}
}

func WithDedupKey(key string) JobOption {
	return func(j *Job) {
		j.DedupKey = key
	}
}

func WithSchedule(t time.Time) JobOption {
	return func(j *Job) {
		j.ScheduledAt = &t
	}
}
