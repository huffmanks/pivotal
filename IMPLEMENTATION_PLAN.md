# URL Shortener Implementation Plan

## Instructions

Implement these phases strictly in the exact order listed below.

At any point, identify the **first incomplete phase** and work only on that phase. Later phases are roadmap context, not implementation requirements for the current phase.

Do not implement, pre-build, or add abstractions solely for later phases unless they are genuinely required by the current phase.

Complete the current phase fully before moving to the next phase.

For every phase:

- Preserve all existing functionality.
- Existing links must continue to redirect correctly.
- New migrations must work on both fresh and existing databases.
- New functionality must be usable through the existing API/UI where appropriate.
- Add appropriate automated tests.
- Keep normal redirect performance a priority.
- Background processing must never unnecessarily block redirects.
- Handle expected failure cases safely.
- Prefer simple, maintainable solutions over speculative abstractions.
- Reuse existing architecture, patterns, and conventions.
- Do not rewrite the application from scratch.
- Prioritize API/server, persistence, redirect behavior, and domain logic over frontend work.
- Keep frontend changes minimal and functional rather than spending significant effort on presentation.

When a phase is complete, verify it before proceeding.

---

## Phase 1 — Core Link Management

Implement the fundamental link lifecycle before adding analytics, QR codes, automation, or advanced routing.

### Link Management

- Edit the destination URL after creation.
- Delete links.
- Disable links without deleting them.
- Re-enable disabled links.
- Add a human-readable link title/name.
- Support 301 and 302 redirects.
- Add an optional expiration date/time.
- Clearly expose link status:
  - Active
  - Disabled
  - Expired

### Fallback Redirects

Allow disabled or expired links to optionally redirect to a configured fallback URL instead of failing.

The redirect system must correctly determine whether the normal destination or fallback destination should be used.

### Phase 1 Completion

- Existing links still redirect correctly.
- Active, disabled, and expired states work consistently.
- Redirect status codes are configurable.
- Fallback behavior is deterministic and tested.
- API supports the complete link lifecycle.
- Required UI controls exist.
- Migrations work on fresh and existing databases.
- Tests cover normal, disabled, expired, fallback, edit, delete, and redirect-code behavior.

---

## Phase 2 — Analytics Foundation

Build the analytics/event tracking foundation before adding advanced routing or automation.

Track, where available and appropriate:

- Total visits/clicks.
- Unique visitors.
- Timestamp.
- Referrer.
- User agent.
- Browser and version.
- Operating system and version.
- Device type.
- Device model where reasonably available.
- Country.
- Region/state.
- City where reasonably available.
- HTTP status/result.
- Destination.
- Short link.
- QR scan versus normal visit where distinguishable.
- UTM parameters.
- Routing rule used.

Do not collect or retain unnecessary sensitive information.

The analytics system should support aggregation by:

- Link.
- Date/time.
- Referrer.
- Device.
- Country/region.
- Campaign.
- UTM parameters.
- QR versus normal visits.

Make analytics available in the dashboard.

### Phase 2 Completion

- Analytics are recorded reliably without unnecessarily slowing redirects.
- Existing asynchronous/background click processing is reused or extended where appropriate.
- Analytics data has a clear, extensible event model.
- Aggregations needed by the dashboard are supported.
- Privacy-conscious data collection is enforced.
- Existing links and redirects continue to work.
- Tests cover event recording, aggregation, missing metadata, and failure handling.

---

## Phase 3 — QR Codes

Add dynamic QR codes.

- Generate a QR code for every short link.
- QR codes must point to the short URL, never directly to the destination.
- Changing the destination must not require a new QR code.
- Allow QR codes to be downloaded.
- Track QR scans separately where possible.
- Display QR scan metrics in analytics.

Make QR codes first-class objects associated with links where appropriate so future QR-specific functionality can be added cleanly.

Do not duplicate redirect or analytics logic specifically for QR codes.

### Phase 3 Completion

- Every eligible link can have a QR code.
- QR codes remain valid when destinations change.
- QR scans are distinguishable from normal visits where technically possible.
- QR metrics appear in analytics.
- QR generation/download failures are handled safely.
- Tests cover generation, association, downloads, and scan attribution.

---

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

---

## Phase 5 — Visit & Scan Limits

Add configurable automatic limits to links.

Allow links to automatically disable after:

- A specified number of total visits.
- A specified number of unique visitors.
- A specified number of QR scans.

When a limit is reached:

1. Mark the link appropriately.
2. Disable normal routing.
3. Redirect to the configured fallback URL if one exists.
4. Otherwise use the appropriate disabled/expired response.

Display current usage and configured limits in the dashboard.

Limits must be enforced consistently regardless of whether enforcement occurs synchronously or through background processing.

---

## Phase 6 — URL Unfurling & Metadata

When a destination URL is created or changed, fetch and inspect it asynchronously.

Store/display useful metadata such as:

- Page title.
- Description.
- Canonical URL.
- Domain.
- Favicon.
- Preview image where available.
- HTTP status.
- Final destination after redirects.

Show this information in the dashboard.

Allow metadata to be refreshed.

A failed metadata request must never prevent creation or editing of a valid short link.

Metadata fetching must remain independent from normal redirect processing.

---

## Phase 7 — Advanced Routing

Add routing rules after the core redirect and analytics systems are stable.

### Device Routing

Support:

- Desktop.
- Mobile.
- Tablet.

Allow each device type to have a different destination.

Provide a default destination when no device rule matches.

### Time-Based Routing

Support:

- Business hours.
- Custom time ranges.
- Specific date ranges.
- Scheduled destination changes.

Define and enforce a clear precedence order when multiple rules could match.

Always have a predictable fallback destination.

Routing decisions should remain deterministic and should not require background processing during a normal redirect.

---

## Phase 8 — Scheduling

Build scheduling on top of the background processing and routing systems.

Support:

- Link activation date/time.
- Link deactivation date/time.
- Scheduled destination changes.
- Scheduled routing rules.
- Automatic expiration.
- Fallback destinations after schedules end.

The dashboard should make the current and upcoming state of a scheduled link obvious.

Scheduling should integrate with the existing link status and routing models rather than creating competing state systems.

---

## Phase 9 — Campaigns & UTM Management

### Campaigns

Allow links to belong to campaigns.

Campaigns should have:

- Name.
- Description/metadata where appropriate.
- Associated links.
- Total clicks.
- Unique visitors.
- Traffic over time.
- Referrer breakdown.
- Device breakdown.
- Geographic breakdown.
- QR scan metrics.
- UTM metrics.

Campaign analytics should aggregate the underlying link analytics rather than duplicate analytics records.

### UTM Management

Support:

- `utm_source`
- `utm_medium`
- `utm_campaign`
- `utm_term`
- `utm_content`

Users should be able to configure these without manually constructing query strings.

Correctly preserve existing destination query parameters.

Expose UTM information within analytics and campaigns.

---

## Phase 10 — Link Health Monitoring

Implement periodic health monitoring using the background-job system.

Check for:

- HTTP errors.
- 404/410 responses.
- Connection failures.
- DNS failures.
- TLS/SSL failures.
- Redirect loops.
- Excessive redirect chains.
- Unexpected destination changes.
- Other significant destination-health problems.

Store:

- Current health status.
- Last checked time.
- Last status change.
- Relevant error information.
- Previous health state.

Display health clearly in the dashboard.

Do not make normal redirects dependent on a health check.

Health checks must be isolated from normal redirect traffic and must respect reasonable timeouts.

---

## Phase 11 — Notifications

Add notifications for important events:

- Link becomes unhealthy.
- Link becomes healthy again.
- Destination changes unexpectedly.
- Link expires.
- Visit limit reached.
- Unique visitor limit reached.
- QR scan limit reached.

Use existing notification infrastructure if available.

Keep notification delivery decoupled from link processing so different notification channels can be added later.

Notification failures must not affect redirects, link creation, or other core functionality.

---

## Phase 12 — Organization & Discovery

Implement organization and discovery features after the underlying link metadata, campaigns, statuses, and analytics are established.

### Domain/Path Grouping

Allow links to be naturally grouped/filterable by:

- Domain.
- Subdomain.
- Path.
- Subpath.

Example:

example.com
├── /admissions
├── /events
├── /giving
└── /news

other.example.com
├── /docs
└── /resources

### Tags

Allow multiple tags per link.

Support filtering/searching by:

- Domain.
- Path.
- Tags.
- Campaign.
- Status.
- Date.
- Other useful metadata.

Search and filtering should build on existing link metadata rather than introducing duplicate representations of the same information.
