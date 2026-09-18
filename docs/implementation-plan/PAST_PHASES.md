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
