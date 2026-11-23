# ADR 0005: Frontend Serving and Deployment

- **Status:** Accepted for MVP foundation
- **Date:** 2026-11-23

## Context

The frontend is HTML/CSS/Vanilla JS with no framework. The README suggests Render Static Sites; the assignment brief allows Render deployment. We need a simple MVP deployment model with minimal CORS/auth complexity.

## Decision

- For MVP, **serve frontend static files from the Go backend** under the same origin when deployed as a single service.
- Keep frontend assets in `frontend/` and allow the backend to serve them in production; during development, a simple static file server is acceptable.
- Design API endpoints under `/api/v1` to allow future separation into a dedicated static frontend deployment if needed.
- Target **Render** for deployment; use PostgreSQL for centralized persistence and environment-based configuration.

## Consequences

- Same-origin deployment simplifies CORS and cookie/session handling.
- Backend becomes responsible for static asset serving in MVP; this is acceptable given low complexity.
- If later split into Render Static Site + API, CORS, auth, and asset versioning must be revisited.
- Deployment docs must include environment variables, migrations, and health checks.
