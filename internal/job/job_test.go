package job

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"pivotal/internal/link"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func setupJobTest(t *testing.T) (*sql.DB, *Manager, *JobExecutor) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE CHECK(length(slug) <= 255),
			destination_url TEXT NOT NULL,
			title TEXT,
			is_custom INTEGER NOT NULL DEFAULT 0,
			click_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'expired')),
			disabled_at DATETIME,
			enabled_at DATETIME,
			redirect_type TEXT NOT NULL DEFAULT '302' CHECK (redirect_type IN ('301', '302')),
			fallback_url TEXT
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug ON links(slug);
		CREATE TABLE link_clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
			referer TEXT,
			user_agent TEXT,
			clicked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			qr_code_id INTEGER
		);
		CREATE TABLE qr_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
			short_url TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(link_id)
		);
		CREATE TABLE IF NOT EXISTS jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL CHECK(type IN ('expire_links', 'enforce_visit_limits', 'enforce_unique_visitor_limits', 'enforce_qr_scan_limits', 'health_check', 'refresh_metadata', 'process_routing_changes', 'send_notification')),
			payload TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'processing', 'completed', 'failed')),
			priority INTEGER NOT NULL DEFAULT 5,
			max_retries INTEGER NOT NULL DEFAULT 3,
			retry_count INTEGER NOT NULL DEFAULT 0,
			dedup_key TEXT,
			scheduled_at DATETIME,
			started_at DATETIME,
			completed_at DATETIME,
			failed_at DATETIME,
			error_message TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(type, dedup_key)
		);
		CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
		CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type);
		CREATE INDEX IF NOT EXISTS idx_jobs_scheduled ON jobs(scheduled_at) WHERE scheduled_at IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_jobs_priority ON jobs(priority);
	`)
	if err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	linkRepo := link.NewRepository(db, nil, nil)
	linkSvc, _ := link.NewService(linkRepo, 100, "http://localhost:3011")

	repository := NewRepository(db)
	executor := NewJobExecutor(linkSvc, repository)
	manager := NewManager(repository, executor, 2)

	return db, manager, executor
}

func TestQueue_EnqueueAndDequeue(t *testing.T) {
	db, manager, _ := setupJobTest(t)
	defer db.Close()

	ctx := context.Background()
	job, err := manager.Enqueue(ctx, &Job{
		Type:     JobTypeExpireLinks,
		Payload:  "check_expired",
		Priority: 5,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if job.ID == 0 {
		t.Error("expected non-zero job ID")
	}
	if job.Status != JobStatusPending {
		t.Errorf("expected status %q, got %q", JobStatusPending, job.Status)
	}
}

func TestQueue_RetryAndExhaustRetries(t *testing.T) {
	db, manager, executor := setupJobTest(t)
	defer db.Close()

	executor.customHandler = func(ctx context.Context, job *Job) JobResult {
		return JobResult{Success: false, Retryable: true, Error: "simulated error"}
	}

	manager.Start()
	defer manager.Stop()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:       JobTypeHealthCheck,
		Payload:    "health",
		MaxRetries: 2,
		DedupKey:   "retry_test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(3 * time.Second)

	repositoryd, found, err := manager.GetJob(ctx, 1)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if repositoryd.RetryCount < 2 {
		t.Errorf("expected at least 2 retries, got %d", repositoryd.RetryCount)
	}
	if repositoryd.Status != JobStatusFailed {
		t.Errorf("expected status %q, got %q", JobStatusFailed, repositoryd.Status)
	}
}

func TestQueue_CompletedJobNotRetried(t *testing.T) {
	db, manager, executor := setupJobTest(t)
	defer db.Close()

	executor.customHandler = func(ctx context.Context, job *Job) JobResult {
		return JobResult{Success: true}
	}

	manager.Start()
	defer manager.Stop()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:     JobTypeHealthCheck,
		Payload:  "health",
		DedupKey: "completed_test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	repositoryd, found, err := manager.GetJob(ctx, 1)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if repositoryd.Status != JobStatusCompleted {
		t.Errorf("expected status %q, got %q", JobStatusCompleted, repositoryd.Status)
	}
}

func TestQueue_DedupKeyPreventsDuplicates(t *testing.T) {
	db, manager, _ := setupJobTest(t)
	defer db.Close()

	ctx := context.Background()
	job1, err := manager.Enqueue(ctx, &Job{
		Type:     JobTypeExpireLinks,
		Payload:  "check",
		DedupKey: "dedup_1",
	})
	if err != nil {
		t.Fatalf("first Enqueue failed: %v", err)
	}

	job2, err := manager.Enqueue(ctx, &Job{
		Type:     JobTypeExpireLinks,
		Payload:  "check_updated",
		DedupKey: "dedup_1",
	})
	if err != nil {
		t.Fatalf("second Enqueue failed: %v", err)
	}

	if job1.ID != job2.ID {
		t.Errorf("expected same job ID for dedup, got %d and %d", job1.ID, job2.ID)
	}
}

func TestQueue_FailedJobNotRetryable(t *testing.T) {
	db, manager, executor := setupJobTest(t)
	defer db.Close()

	executor.customHandler = func(ctx context.Context, job *Job) JobResult {
		return JobResult{Success: false, Retryable: false, Error: "permanent failure"}
	}

	manager.Start()
	defer manager.Stop()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:       JobTypeHealthCheck,
		Payload:    "health",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	repositoryd, found, err := manager.GetJob(ctx, 1)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if repositoryd.Status != JobStatusFailed {
		t.Errorf("expected status %q, got %q", JobStatusFailed, repositoryd.Status)
	}
	if repositoryd.RetryCount != 0 {
		t.Errorf("expected 0 retries for non-retryable failure, got %d", repositoryd.RetryCount)
	}
}

func TestQueue_GracefulShutdown(t *testing.T) {
	db, manager, executor := setupJobTest(t)
	defer db.Close()

	executor.customHandler = func(ctx context.Context, job *Job) JobResult {
		time.Sleep(100 * time.Millisecond)
		return JobResult{Success: true}
	}

	manager.Start()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:    JobTypeHealthCheck,
		Payload: "health",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	manager.Stop()

	repositoryd, found, err := manager.GetJob(ctx, 1)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if repositoryd.Status == JobStatusPending {
		t.Error("expected job to have been processed or at least started before shutdown")
	}
}

func TestManager_Notifications(t *testing.T) {
	db, manager, _ := setupJobTest(t)
	defer db.Close()

	manager.Start()
	defer manager.Stop()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:       JobTypeSendNotification,
		Payload:    "test notification",
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	notif := <-manager.Notifications()
	if notif.JobID != 1 {
		t.Errorf("expected notification JobID 1, got %d", notif.JobID)
	}
}

func TestRepository_Enqueue(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:     JobTypeExpireLinks,
		Payload:  "test",
		Priority: 5,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if job.ID != 1 {
		t.Errorf("expected ID 1, got %d", job.ID)
	}
	if job.Status != JobStatusPending {
		t.Errorf("expected status %q, got %q", JobStatusPending, job.Status)
	}
}

func TestRepository_GetNextPending(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	_, err := repository.Enqueue(ctx, &Job{
		Type:     JobTypeExpireLinks,
		Payload:  "test",
		Priority: 10,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	job, found, err := repository.GetNextPending(ctx)
	if err != nil {
		t.Fatalf("GetNextPending failed: %v", err)
	}
	if !found {
		t.Fatal("expected job to be found")
	}
	if job.Priority != 10 {
		t.Errorf("expected priority 10, got %d", job.Priority)
	}
}

func TestRepository_MarkProcessingAndComplete(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:    JobTypeExpireLinks,
		Payload: "test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if err := repository.MarkProcessing(ctx, job.ID); err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}

	updated, found, err := repository.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if updated.Status != JobStatusProcessing {
		t.Errorf("expected status %q, got %q", JobStatusProcessing, updated.Status)
	}

	if err := repository.MarkCompleted(ctx, job.ID); err != nil {
		t.Fatalf("MarkCompleted failed: %v", err)
	}

	completed, found, err := repository.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if completed.Status != JobStatusCompleted {
		t.Errorf("expected status %q, got %q", JobStatusCompleted, completed.Status)
	}
}

func TestRepository_MarkFailed(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:    JobTypeExpireLinks,
		Payload: "test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if err := repository.MarkProcessing(ctx, job.ID); err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}

	errMsg := "something went wrong"
	if err := repository.MarkFailed(ctx, job.ID, errMsg); err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}

	failed, found, err := repository.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if failed.Status != JobStatusFailed {
		t.Errorf("expected status %q, got %q", JobStatusFailed, failed.Status)
	}
	if failed.ErrorMessage != nil && *failed.ErrorMessage != errMsg {
		t.Errorf("expected error %q, got %q", errMsg, *failed.ErrorMessage)
	}
	if failed.ErrorMessage == nil {
		t.Error("expected non-nil error message")
	}
}

func TestRepository_DedupOnEnqueue(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job1, err := repository.Enqueue(ctx, &Job{
		Type:     JobTypeHealthCheck,
		Payload:  "health_check",
		DedupKey: "dedup_key_1",
	})
	if err != nil {
		t.Fatalf("first Enqueue failed: %v", err)
	}

	job2, err := repository.Enqueue(ctx, &Job{
		Type:     JobTypeHealthCheck,
		Payload:  "health_check_updated",
		DedupKey: "dedup_key_1",
	})
	if err != nil {
		t.Fatalf("second Enqueue failed: %v", err)
	}

	if job1.ID != job2.ID {
		t.Errorf("expected same ID for dedup, got %d and %d", job1.ID, job2.ID)
	}
}

func TestRepository_ListByStatus(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repository.Enqueue(ctx, &Job{
			Type:    JobTypeExpireLinks,
			Payload: "test",
		})
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
	}

	jobs, err := repository.ListByStatus(ctx, JobStatusPending, 10)
	if err != nil {
		t.Fatalf("ListByStatus failed: %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 pending jobs, got %d", len(jobs))
	}
}

func TestRepository_ListByStatus_Limit(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := repository.Enqueue(ctx, &Job{
			Type:    JobTypeExpireLinks,
			Payload: "test",
		})
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
	}

	jobs, err := repository.ListByStatus(ctx, JobStatusPending, 2)
	if err != nil {
		t.Fatalf("ListByStatus failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestRepository_CleanupCompleted(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	_, err := repository.Enqueue(ctx, &Job{
		Type:    JobTypeExpireLinks,
		Payload: "test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	job, _, _ := repository.GetByID(ctx, 1)
	repository.MarkCompleted(ctx, job.ID)

	err = repository.CleanupCompleted(ctx, time.Now())
	if err != nil {
		t.Fatalf("CleanupCompleted failed: %v", err)
	}

	_, found, err := repository.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found {
		t.Error("expected completed job to be cleaned up")
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	_, found, err := repository.GetByID(ctx, 999)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found {
		t.Error("expected job not found")
	}
}

func TestExecutor_Handle_UnknownType(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	result := executor.Handle(context.Background(), &Job{
		Type:    JobType("unknown"),
		Payload: "test",
	})
	if result.Success {
		t.Error("expected failure for unknown job type")
	}
	if result.Retryable {
		t.Error("expected non-retryable for unknown job type")
	}
}

func TestExecutor_Handle_ExpireLinks(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	result := executor.Handle(context.Background(), &Job{
		Type:    JobTypeExpireLinks,
		Payload: "check",
	})
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestExecutor_Handle_HealthCheck(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	result := executor.Handle(context.Background(), &Job{
		Type:    JobTypeHealthCheck,
		Payload: "check",
	})
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestJobStatusValues(t *testing.T) {
	if JobStatusPending != "pending" {
		t.Errorf("expected %q, got %q", "pending", JobStatusPending)
	}
	if JobStatusProcessing != "processing" {
		t.Errorf("expected %q, got %q", "processing", JobStatusProcessing)
	}
	if JobStatusCompleted != "completed" {
		t.Errorf("expected %q, got %q", "completed", JobStatusCompleted)
	}
	if JobStatusFailed != "failed" {
		t.Errorf("expected %q, got %q", "failed", JobStatusFailed)
	}
}

func TestJobTypeValues(t *testing.T) {
	expected := []string{
		"expire_links",
		"enforce_visit_limits",
		"enforce_unique_visitor_limits",
		"enforce_qr_scan_limits",
		"health_check",
		"refresh_metadata",
		"process_routing_changes",
		"send_notification",
	}
	actual := []string{
		string(JobTypeExpireLinks),
		string(JobTypeEnforceVisitLimits),
		string(JobTypeEnforceUniqueVisitorLimits),
		string(JobTypeEnforceQRScanLimits),
		string(JobTypeHealthCheck),
		string(JobTypeRefreshMetadata),
		string(JobTypeProcessRoutingChanges),
		string(JobTypeSendNotification),
	}
	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("expected %q, got %q", exp, actual[i])
		}
	}
}

func TestQueue_IsRunning(t *testing.T) {
	db, manager, _ := setupJobTest(t)
	defer db.Close()

	if manager.IsRunning() {
		t.Error("expected manager to not be running before Start()")
	}

	manager.Start()
	defer manager.Stop()

	if !manager.IsRunning() {
		t.Error("expected manager to be running after Start()")
	}
}

func TestQueue_ProcessJob_NonRetryableFailure(t *testing.T) {
	db, manager, executor := setupJobTest(t)
	defer db.Close()

	executor.customHandler = func(ctx context.Context, job *Job) JobResult {
		return JobResult{Success: false, Retryable: false, Error: "permanent error"}
	}

	manager.Start()
	defer func() {
		time.Sleep(3 * time.Second)
		manager.Stop()
	}()

	ctx := context.Background()
	_, err := manager.Enqueue(ctx, &Job{
		Type:       JobTypeHealthCheck,
		Payload:    "health",
		MaxRetries: 3,
		DedupKey:   "non_retryable_1",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(3 * time.Second)

	repositoryd, found, err := manager.GetJob(ctx, 1)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if !found {
		t.Fatal("job not found")
	}
	if repositoryd.Status != JobStatusFailed {
		t.Errorf("expected status %q, got %q", JobStatusFailed, repositoryd.Status)
	}
}

func TestRepository_MarkProcessing_PendingJobOnly(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:    JobTypeExpireLinks,
		Payload: "test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	repository.MarkProcessing(ctx, job.ID)
}

func TestExecutor_SendNotification(t *testing.T) {
	db, _, executor := setupJobTest(t)
	defer db.Close()

	ctx := context.Background()
	result := executor.Handle(ctx, &Job{
		Type:    JobTypeSendNotification,
		Payload: "test message",
	})
	if !result.Success {
		t.Errorf("expected success, got %s", result.Error)
	}

	notif := <-executor.Notifications()
	if notif.Message != "test message" {
		t.Errorf("expected message %q, got %q", "test message", notif.Message)
	}
}

func TestExecutor_Handle_WithNilLinkSvc(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	result := executor.Handle(context.Background(), &Job{
		Type:    JobTypeRefreshMetadata,
		Payload: "refresh",
	})
	_ = result
}

func TestRepository_IncrementRetry(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:       JobTypeHealthCheck,
		Payload:    "health",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	repository.MarkProcessing(ctx, job.ID)
	repository.MarkFailed(ctx, job.ID, "error")

	err = repository.IncrementRetry(ctx, job.ID)
	if err != nil {
		t.Fatalf("IncrementRetry failed: %v", err)
	}
}

func TestRepository_GetNextPending_FailedJob(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	_, err := repository.Enqueue(ctx, &Job{
		Type:     JobTypeHealthCheck,
		Payload:  "health",
		DedupKey: "failed_test",
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	job, _, _ := repository.GetByID(ctx, 1)
	repository.MarkFailed(ctx, job.ID, "error")

	failedJob, found, err := repository.GetNextPending(ctx)
	if err != nil {
		t.Fatalf("GetNextPending failed: %v", err)
	}
	if !found {
		t.Fatal("expected failed job to be found for retry")
	}
	if job.ID != failedJob.ID {
		t.Errorf("expected same job ID, got %d and %d", job.ID, failedJob.ID)
	}
}

func TestRepository_MarkCompleted_ResetRetry(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	ctx := context.Background()

	job, err := repository.Enqueue(ctx, &Job{
		Type:       JobTypeHealthCheck,
		Payload:    "health",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	repository.MarkProcessing(ctx, job.ID)
	repository.MarkFailed(ctx, job.ID, "error")
	repository.MarkCompleted(ctx, job.ID)

	completed, _, _ := repository.GetByID(ctx, job.ID)
	if completed.RetryCount != 0 {
		t.Errorf("expected retry_count reset to 0, got %d", completed.RetryCount)
	}
}

func TestExecutor_Handle_ProcessRoutingChanges(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	result := executor.Handle(context.Background(), &Job{
		Type:    JobTypeProcessRoutingChanges,
		Payload: "routing",
	})
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestExecutor_Handle_SendNotificationChannelFull(t *testing.T) {
	db, _, _ := setupJobTest(t)
	defer db.Close()

	repository := NewRepository(db)
	executor := NewJobExecutor(nil, repository)

	for i := 0; i < 200; i++ {
		result := executor.Handle(context.Background(), &Job{
			Type:    JobTypeSendNotification,
			Payload: "msg",
		})
		if !result.Success {
			t.Errorf("expected success for notification %d, got error: %s", i, result.Error)
		}
	}
}
