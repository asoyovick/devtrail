# Changelog

All notable changes to DevTrail will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) while in the `0.x.x` series until the first stable public release.

## [Unreleased]

### Added
- Initial repository assessment and Phase 1 architecture proposal (`docs/requirements/phase1-assessment.md`).
- Initial project README describing DevTrail vision, scope, and planned architecture.
- Repository workflow baseline: `.gitignore`, `CONTRIBUTING.md`, and branching workflow documentation.
- Phase 2 architecture decision records for module layout, auth, event model/sync, privacy defaults, and frontend serving (`docs/decisions/`).
- Initial Phase 2 foundation scaffold: backend health endpoint, CLI entrypoint, frontend shell placeholder, config example, and Makefile targets.
- Backend platform database foundation: pgx pool connection, `/readyz` endpoint, initial SQL migrations, and migration Makefile targets.
- Backend HTTP platform: request ID, structured request logging, panic recovery middleware, and `/api/v1/healthz` endpoint.
- Auth foundation: `auth_sessions` migration, password hashing, and secure token generation/hash helpers.
- User auth endpoints: register, login, logout, and `/me` with opaque session tokens and ownership-scoped auth middleware.
- Project CRUD endpoints: create, list, get, update, and archive with per-user ownership checks.
- Development session endpoints: start, list, get, pause, resume, and end with project ownership checks and state transition validation.
- Activity event foundation: `activity_events` migration, idempotent batch ingestion endpoint, and activity timeline listing endpoint.
- Debugging journal foundation: `investigations` migration and CRUD endpoints with user/project ownership checks.
- Engineering decisions foundation: `decisions` migration and CRUD endpoints with user/project ownership checks.
- Technical notes foundation: `notes` migration and CRUD/delete endpoints with user ownership checks.
- Test run history foundation: `test_runs` migration and create/list/get endpoints with user/project ownership checks.
- Git activity foundation: `git_commits` migration and create/list endpoints with idempotent upsert by project+commit hash.
- Reporting foundation: daily/weekly/project report endpoints with JSON and Markdown output from recorded data.
- Insights foundation: rule-based, evidence-backed insights for recurring test failures and improvement areas.
- Backend hardening: request body limits, security headers, CORS configuration, and in-memory rate limiting middleware.
