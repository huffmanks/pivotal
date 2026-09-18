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
