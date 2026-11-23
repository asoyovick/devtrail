# ADR 0001: Module Layout and Go Module Paths

- **Status:** Accepted for MVP foundation
- **Date:** 2026-11-23

## Context

DevTrail is a monorepo with three components: backend, CLI, and frontend. The repository currently contains `backend/go.mod` with `module devtrail` and Go 1.22.2. The README proposes separate `backend/go.mod` and `cli/go.mod`.

We need a module layout that supports independent backend/CLI development while keeping the MVP simple.

## Decision

- Use **separate Go modules** for `backend/` and `cli/`.
- Keep the frontend as static assets (no Go module).
- For now, use module paths:
  - `backend/go.mod`: `devtrail/backend`
  - `cli/go.mod`: `devtrail/cli` (when created)
- When the canonical repository remote is established, migrate module paths to the fully qualified VCS path (e.g., `github.com/<owner>/devtrail/backend`) in a dedicated change.

## Consequences

- Backend and CLI dependencies evolve independently.
- Local development is simple; `go` commands run inside each module.
- Module path migration will require import updates later; do it before external consumers or CI release workflows depend on paths.
- Shared code between backend and CLI should be avoided in MVP; if needed later, extract a carefully scoped shared module with clear justification.
