## Phase 4 — Background Jobs & Link Automation

Introduce the background processing needed for automation.

Implement a reliable queue/worker/cron/scheduled-job mechanism appropriate for the existing architecture.

Use it for:

- Expiring links.
- Enforcing visit limits.
- Enforcing unique-visitor limits.
- Enforcing QR-scan limits.
- Periodic link health checks.
- Refreshing URL metadata.
- Processing scheduled routing changes.
- Sending notifications.

Jobs must be:

- Retryable.
- Idempotent where possible.
- Safe to run multiple times.
- Non-blocking to normal redirects.

Do not make redirect requests wait for background processing.

Expose an API for monitoring and managing background jobs and queues, including:

- Viewing active, pending, completed, and failed jobs.
- Viewing queue status and worker status.
- Manually triggering or retrying jobs.
- Cancelling or stopping queued or running jobs where supported.
- Clearing or managing failed/completed jobs where appropriate.

The API should provide enough visibility and control to operate the job system without directly accessing its underlying storage or process.

### Phase 4 Completion

- Jobs survive transient failures.
- Retries do not create duplicate side effects.
- Jobs can safely be rerun.
- Failed jobs do not break normal redirects.
- Existing asynchronous click processing remains reliable.
- Jobs and queues can be monitored and managed through the API.
- Tests cover retries, duplicate execution, failures, graceful shutdown, and job management.
