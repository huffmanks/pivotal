just test
// go test -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out ./...
...
...
...
2026/09/18 16:35:57 worker 0: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 1: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 0: processing job 1 of type health_check
2026/09/18 16:35:57 worker 1: processing job 1 of type health_check
2026/09/18 16:35:57 worker 0: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 1: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 0: processing job 1 of type health_check
2026/09/18 16:35:57 worker 1: processing job 1 of type health_check
2026/09/18 16:35:57 worker 0: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 1: job 1 failed permanently: permanent error
2026/09/18 16:35:57 worker 0: processing job 1 of type health_check
2026/09/18 16:35:57 worker 1: processing job 1 of type health_check
2026/09/18 16:35:57 scheduler stopped
2026/09/18 16:35:57 worker 0: failed to mark job 1 as processing: context canceled
2026/09/18 16:35:57 worker 1: error fetching job: sqlite3: interrupted
2026/09/18 16:35:57 job queue stopped
2026/09/18 16:35:57 job manager stopped
--- PASS: TestQueue_ProcessJob_NonRetryableFailure (6.02s)
=== RUN TestRepository_MarkProcessing_PendingJobOnly
job_test.go:741: Enqueue failed: sql: Scan error on column index 7, name "dedup_key": converting NULL to string is unsupported
--- FAIL: TestRepository_MarkProcessing_PendingJobOnly (0.01s)
=== RUN TestExecutor_SendNotification
--- PASS: TestExecutor_SendNotification (0.01s)
=== RUN TestExecutor_Handle_WithNilLinkSvc
--- PASS: TestExecutor_Handle_WithNilLinkSvc (0.01s)
=== RUN TestRepository_IncrementRetry
job_test.go:793: Enqueue failed: sql: Scan error on column index 7, name "dedup_key": converting NULL to string is unsupported
--- FAIL: TestRepository_IncrementRetry (0.01s)
=== RUN TestRepository_GetNextPending_FailedJob
--- PASS: TestRepository_GetNextPending_FailedJob (0.02s)
=== RUN TestRepository_MarkCompleted_ResetRetry
job_test.go:849: Enqueue failed: sql: Scan error on column index 7, name "dedup_key": converting NULL to string is unsupported
--- FAIL: TestRepository_MarkCompleted_ResetRetry (0.01s)
=== RUN TestExecutor_Handle_ProcessRoutingChanges
2026/09/18 16:35:57 job process_routing_changes: processed routing changes
--- PASS: TestExecutor_Handle_ProcessRoutingChanges (0.01s)
=== RUN TestExecutor_Handle_SendNotificationChannelFull
--- PASS: TestExecutor_Handle_SendNotificationChannelFull (0.01s)
FAIL
coverage: 17.0% of statements in ./...
FAIL pivotal/internal/job 12.063s
=== RUN TestClickTracker_ConcurrencyAndShutdown
--- PASS: TestClickTracker_ConcurrencyAndShutdown (0.30s)
=== RUN TestLinkService_Create
--- PASS: TestLinkService_Create (0.01s)
=== RUN TestLinkService_Update
--- PASS: TestLinkService_Update (0.02s)
=== RUN TestLinkService_DisableEnable
--- PASS: TestLinkService_DisableEnable (0.01s)
=== RUN TestLinkService_Resolve
--- PASS: TestLinkService_Resolve (0.01s)
=== RUN TestLinkService_ResolveWithRedirect
--- PASS: TestLinkService_ResolveWithRedirect (0.02s)
=== RUN TestLinkService_Delete
--- PASS: TestLinkService_Delete (0.01s)
=== RUN TestLinkService_Expiry
--- PASS: TestLinkService_Expiry (0.01s)
=== RUN TestLinkService_ResolveWithRedirect_ExpiredWithoutFallback
--- PASS: TestLinkService_ResolveWithRedirect_ExpiredWithoutFallback (0.01s)
=== RUN TestLinkService_ResolveWithRedirect_ExpiredWithFallback
--- PASS: TestLinkService_ResolveWithRedirect_ExpiredWithFallback (0.01s)
=== RUN TestLinkService_ResolveWithRedirect_DisabledWithoutFallback
--- PASS: TestLinkService_ResolveWithRedirect_DisabledWithoutFallback (0.01s)
=== RUN TestLinkService_ResolveWithRedirect_DisabledWithConfiguredFallback
--- PASS: TestLinkService_ResolveWithRedirect_DisabledWithConfiguredFallback (0.01s)
=== RUN TestLinkService_Resolve_ExpiredLink
--- PASS: TestLinkService_Resolve_ExpiredLink (0.01s)
=== RUN TestLinkService_QRCode
--- PASS: TestLinkService_QRCode (0.03s)
=== RUN TestLinkService_QRCode_DeleteLink
--- PASS: TestLinkService_QRCode_DeleteLink (0.01s)
=== RUN TestLinkService_QRCode_DestinationChange
--- PASS: TestLinkService_QRCode_DestinationChange (0.01s)
PASS
coverage: 18.9% of statements in ./...
ok pivotal/internal/link 1.844s coverage: 18.9% of statements in ./...
pivotal/internal/middleware coverage: 0.0% of statements
pivotal/internal/platform/utils coverage: 0.0% of statements
pivotal/internal/server coverage: 0.0% of statements
pivotal/web coverage: 0.0% of statements
FAIL
error: recipe `test` failed on line 31 with exit code 1
