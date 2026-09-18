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

### Phase 4 Completion

- Jobs survive transient failures.
- Retries do not create duplicate side effects.
- Jobs can safely be rerun.
- Failed jobs do not break normal redirects.
- Existing asynchronous click processing remains reliable.
- Tests cover retries, duplicate execution, failures, and graceful shutdown.
