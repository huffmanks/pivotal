# URL Shortener Implementation Plan

## Instructions

Implement only what is required for the current phase. Build abstractions only when they are genuinely needed.

Complete the current phase fully and verify it.

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
- Focus implementation on the API, server, persistence, redirect behavior, and domain logic. Keep frontend work minimal and functional.
- Treat `web/` as completely out of scope. Leave its contents unchanged.

Proceed with `docs/implementation-plan/CURRENT_PHASE.md`
