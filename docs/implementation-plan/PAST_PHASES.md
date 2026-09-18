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
