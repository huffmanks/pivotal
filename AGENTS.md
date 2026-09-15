# Project Overview

A lightweight URL shortener with a Go backend, SQLite database, and Svelte frontend.

## Structure

- `main.go` — application entry point
- `internal/` — backend application code
- `migrations/` — database migrations
- `web/` — Svelte frontend
- `README.md` — project documentation

## Guidelines

- Follow the existing architecture and patterns.
- Keep changes small, focused, and reviewable.
- Prefer simple solutions over unnecessary abstractions.
- Preserve existing behavior unless the current task requires otherwise.
- Add new migrations instead of modifying existing ones.
- Keep redirect handling fast.
- Implement phases in order.
- Use `IMPLEMENTATION_PLAN.md` as the source of truth for the current phase.
- Do not implement later-phase requirements early.
- Run relevant tests after completing each phase.
