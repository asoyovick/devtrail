# DevTrail — Phase 1: Repository Assessment and Architecture Proposal

**Date:** 2026-11-23
**Status:** Assessment complete; implementation not started
**Scope:** Phase 1 only — repository assessment, architecture, schema, API, CLI, frontend, roadmap, risks

---

## 1. Repository Assessment

### 1.1 Current State

The repository is an early-stage skeleton, not a functioning application.

| Area | State | Evidence |
|---|---|---|
| Git repository | **Not initialized** | `git status` returns `fatal: not a git repository` |
| README | Present, detailed | `README.md` (514 lines) describes vision, architecture, roadmap, MVP scope |
| Backend | Skeleton only | `backend/go.mod` exists; `backend/internal/*` domain directories exist but are empty; no Go source files |
| CLI | Missing | No `cli/` directory |
| Frontend | Missing | No `frontend/` directory |
| Docs | Missing before this assessment | No `docs/` directory |
| Database migrations | Missing | No `migrations/` directory |
| Configuration | Missing | No config files, env examples, or config packages |
| Tests | Missing | No `*_test.go` files |
| CI/CD / deployment | Missing | No Render config, Dockerfile, Makefile, or GitHub workflows |
| License / contributing | Missing | No `LICENSE`, `CONTRIBUTING.md` |

### 1.2 Existing Backend Skeleton

```
backend/
├── go.mod                  # module devtrail, go 1.22.2
└── internal/
    ├── activity/           # empty
    ├── debugging/          # empty
    ├── decision/           # empty
    ├── git/                # empty
    ├── insights/           # empty
    ├── project/            # empty
    ├── reports/            # empty
    ├── session/            # empty
    ├── testing/            # empty
    ├── user/               # empty
    └── workspace/          # empty
```

**Observations:**

- `go.mod` declares module `devtrail` with Go `1.22.2`. The module path is local/bare; before publishing or multi-module use, decide whether to keep a single module at repo root, separate modules per component, or use a fully qualified module path. For the MVP monorepo, a clear module strategy is needed (see Risks).
- Domain directories exist under `internal/`, which aligns with Go conventions and the README. They are placeholders only.
- `workspace/` exists in the skeleton but the product brief emphasizes personal-first development; workspace/team concepts should be deferred unless needed for ownership boundaries.
- Missing domains implied by the brief: authentication/session-token handling (may live under `user` or `auth`), notes/technical notes, synchronization/ingestion, CLI sync tokens, migrations/config/cmd.

### 1.3 README vs. Product Brief Alignment

The README is broadly aligned with the product brief supplied in the assignment. Notable differences to resolve before implementation:

| Topic | README | Assignment brief | Recommended direction |
|---|---|---|---|
| Frontend deployment | "Render Static Sites" | "Render for deployment" | Prefer single Render deployment strategy; static frontend can be served by Go backend or Render Static Site. Decide in Phase 2. |
| Workspace | Included in backend structure and diagram | Not a core MVP module; personal-first | Keep data model owner-scoped; defer workspace/team UI and APIs. |
| Navigation | README proposes page set; brief specifies exact nav structure | Brief gives canonical nav | **Follow the brief's navigation structure** for the web app. |
| Intelligence | Includes "AI-assisted debugging analysis" as long-term | "Do not introduce AI into the initial MVP" | Rule-based, evidence-based insights only in MVP. |
| Reports | Daily/Weekly/Project, Markdown/JSON | Same | Aligned. |
| Technical Notes | Mentioned in nav in brief | Not prominent in README feature list | Add Notes as a first-class knowledge module; keep lightweight. |
| Git Activity page | Brief includes "Git Activity" under Development | README structure less explicit | Include as dedicated view/API surface. |

### 1.4 What Can Be Reused

- **README.md** — good product context; should be updated as architecture decisions are finalized, but do not overwrite blindly.
- **Backend directory layout intent** — `internal/<domain>` is a reasonable starting point.
- **Domain naming** — mostly aligned with required modules.

### 1.5 What Should Not Be Assumed

- No existing code behavior to preserve in `internal/*` (directories are empty).
- No database schema exists.
- No API contracts exist.
- No CLI command behavior exists.
- No frontend components exist.

---

## 2. Architecture Explanation

### 2.1 System Overview

DevTrail is a three-component system:

1. **DevTrail CLI (Go)** — local collector and developer interface. Runs on the developer machine, inspects Git repositories, tracks sessions, stores events in local SQLite, and synchronizes with the backend.
2. **DevTrail Backend (Go REST API)** — central source of truth for authenticated users. Receives synchronized events, manages projects/sessions/investigations/decisions/tests/notes, generates insights and reports.
3. **DevTrail Web Application (HTML/CSS/Vanilla JS)** — browser workspace for exploring project and global engineering history.

```
Developer Machine                              Central Platform
┌──────────────────────────────┐              ┌──────────────────────────────┐
│ DevTrail CLI                 │              │ DevTrail Web App             │
│ - repo registration          │              │ - HTML/CSS/Vanilla JS        │
│ - git inspection             │              │ - REST client                │
│ - session tracking           │              │ - dashboards/views           │
│ - test capture (supported)   │              └──────────────┬───────────────┘
│ - manual investigations      │                             │ HTTPS/JSON
│ - local event queue          │                             ▼
│ - sync client                │              ┌──────────────────────────────┐
│        │                     │   sync/API   │ DevTrail Backend (Go)        │
│        ▼                     ├─────────────►│ - auth/authorization         │
│  SQLite (local store/queue)  │              │ - ingestion & dedup          │
└──────────────────────────────┘              │ - domain services            │
                                              │ - rule-based insights        │
                                              │ - reports/export             │
                                              └──────────────┬───────────────┘
                                                             │
                                                             ▼
                                              ┌──────────────────────────────┐
                                              │ PostgreSQL                   │
                                              │ - users/projects/events      │
                                              │ - investigations/decisions   │
                                              │ - tests/notes/reports cache  │
                                              └──────────────────────────────┘
```

### 2.2 Architectural Principles Applied

- **Personal-first:** all central data is owned by a user; authorization checks enforce ownership on every read/write.
- **Local-first collection:** CLI writes to SQLite first; sync is an explicit/background-capable operation with retry and idempotency.
- **Append-only where appropriate:** activity events are immutable records; mutable domain records (investigations, decisions, notes) keep `created_at`/`updated_at` and may gain edit history later.
- **Evidence-based intelligence:** insights store the query/window/evidence used to generate them; no LLM-generated causal claims in MVP.
- **Simple modularity:** domain packages own validation, business rules, storage interfaces, and HTTP handlers for their area; avoid generic repository/framework abstractions.
- **Privacy by design:** CLI redacts/sanitizes metadata before persistence and sync; backend rejects sensitive fields by schema and validation.

### 2.3 Backend Design

**Style:** modular monolith in Go, organized by domain under `internal/`.

Proposed top-level layout:

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # process entrypoint
├── internal/
│   ├── platform/
│   │   ├── config/                 # env/config loading
│   │   ├── database/               # pgx pool, migrations runner
│   │   ├── httpserver/             # router, middleware, response helpers
│   │   ├── auth/                   # password hashing, tokens/sessions
│   │   ├── validate/               # shared validation helpers
│   │   └── logger/                 # structured logging
│   ├── user/                       # user profile, registration/login
│   ├── project/                    # project CRUD, repo identity
│   ├── session/                    # dev sessions
│   ├── activity/                   # event ingestion/query/timeline
│   ├── gitactivity/                # git-specific views over activity/events
│   ├── debugging/                  # investigations
│   ├── decision/                   # engineering decision records
│   ├── testing/                    # test runs/history
│   ├── note/                       # technical notes
│   ├── insight/                    # rule-based analysis
│   ├── report/                     # daily/weekly/project reports, exports
│   └── sync/                       # CLI sync endpoints, idempotency
├── migrations/
├── configs/
│   └── config.example.env
└── go.mod
```

**Domain package shape (only where justified):**

```
internal/project/
├── model.go        # domain types / DTOs
├── store.go        # SQL queries (concrete or small interface)
├── service.go      # business rules
├── handler.go      # HTTP handlers
└── handler_test.go
```

Do not create all layers mechanically. A simple domain may combine store/service where that is clearer. Interfaces should exist at consumption boundaries (e.g., handler depends on service behavior) only when they aid testing or substitution.

**HTTP conventions:**

- Base path: `/api/v1`.
- JSON request/response bodies.
- Auth via Bearer token (opaque session token or signed JWT; decision in Phase 2). Prefer opaque server-side sessions for simple revocation in MVP.
- Consistent error envelope: `{ "error": { "code": "...", "message": "...", "details": ... } }`.
- Pagination: `limit`, `cursor` or `page`; prefer cursor for timelines/events.
- Timestamps: RFC3339 UTC.
- IDs: UUIDs generated server-side.

### 2.4 CLI Design

Proposed layout:

```
cli/
├── cmd/
│   └── devtrail/
│       └── main.go
├── internal/
│   ├── config/         # CLI config, API URL, auth token storage
│   ├── localdb/        # SQLite schema/migrations/queries
│   ├── repo/           # git repo detection, metadata extraction
│   ├── session/        # start/pause/resume/end
│   ├── activity/       # event model, local recording
│   ├── gitwatch/       # commit/branch/file metadata capture
│   ├── testing/        # test command wrapper/capture
│   ├── debugging/      # investigation capture commands
│   ├── decision/       # decision capture commands
│   ├── sync/           # push/pull, batching, retry, idempotency keys
│   └── output/         # human/JSON output formatting
└── go.mod
```

**Local-first flow:**

1. Command validates input and detects repo/project context.
2. Event/record written to SQLite in a transaction.
3. Sync marks records as pending/uploaded/failed; never blocks core local capture on network failure.
4. Backend ingestion is idempotent using client-generated event IDs.

### 2.5 Frontend Design

Proposed layout:

```
frontend/
├── index.html                  # entry / redirect to overview or login
├── pages/
│   ├── login.html
│   ├── register.html
│   ├── overview.html
│   ├── activity.html
│   ├── projects.html
│   ├── sessions.html
│   ├── git-activity.html
│   ├── testing.html
│   ├── debugging.html
│   ├── decisions.html
│   ├── notes.html
│   ├── insights.html
│   ├── recurring-problems.html
│   ├── improvement-areas.html
│   ├── reports-daily.html
│   ├── reports-weekly.html
│   ├── reports-project.html
│   ├── preferences.html
│   ├── integrations.html
│   └── data-privacy.html
├── css/
│   ├── base.css
│   ├── layout.css
│   ├── components.css
│   └── pages/
├── js/
│   ├── app.js
│   ├── api.js
│   ├── auth.js
│   ├── router-helpers.js
│   ├── state.js
│   ├── components/
│   └── pages/
└── assets/
```

The frontend should use shared layout components rendered by small JS component functions (no framework): sidebar/nav, top bar, cards, tables, filters, empty states, forms, badges, timeline items. Keep API access centralized in `api.js`.

### 2.6 Data Flow Examples

**Git commit capture:**
1. CLI hook or explicit `devtrail git sync` detects new commits.
2. CLI extracts safe metadata: hash, branch, author name/email? (consider privacy; default to local git author only if necessary, prefer omitting), timestamps, message summary? (messages can contain sensitive info; store cautiously or hash/truncate/configure), changed file paths/counts, insertions/deletions if available.
3. CLI stores `git_commit_recorded` event locally.
4. `devtrail sync` pushes pending events to `/api/v1/sync/events`.
5. Backend validates, dedupes by `event_id`, stores event, updates read models/projections as needed.

**Debugging investigation:**
1. Developer creates investigation via CLI or web UI.
2. Record stored as mutable domain entity with timestamps and owner/project association.
3. Activity event emitted: `investigation.created` / `investigation.updated`.
4. Timeline and project views query events and/or domain records.

**Insight generation:**
1. Backend scheduled job or on-demand request runs rule-based queries over events/tests/investigations.
2. Insight record stores type, scope, period, evidence, confidence/qualification, generated_at.
3. UI renders evidence and clearly labels observations vs. recommendations.

---

## 3. Proposed Database Schema

PostgreSQL for centralized persistence. UUID primary keys. Timestamps as `timestamptz` UTC. Use foreign keys, indexes, and check constraints.

### 3.1 Core Identity

#### `users`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| email | citext/text | UNIQUE, NOT NULL | normalize lowercase |
| password_hash | text | NOT NULL | bcrypt/argon2id |
| display_name | text | NOT NULL | |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

#### `auth_sessions` (if opaque sessions chosen)
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| token_hash | text | UNIQUE, NOT NULL | store hash, not raw token |
| expires_at | timestamptz | NOT NULL | |
| revoked_at | timestamptz | NULL | |
| user_agent | text | NULL | optional |
| created_at | timestamptz | NOT NULL | |

### 3.2 Projects and Repositories

#### `projects`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | owner |
| name | text | NOT NULL | |
| description | text | NULL | |
| repo_root_hash | text | NULL | stable identifier derived locally; avoid leaking full path by default |
| default_branch | text | NULL | |
| status | text | NOT NULL | `active`, `archived` |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

Indexes: `(user_id, created_at desc)`, unique `(user_id, name)` optionally.

#### `project_repositories`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| project_id | uuid | FK projects(id), NOT NULL | |
| user_id | uuid | FK users(id), NOT NULL | denormalized for auth/indexing |
| provider | text | NOT NULL | `local_git`, later `github` |
| remote_url | text | NULL | optional, user-controlled |
| local_path_label | text | NULL | user-provided label, not raw machine path by default |
| created_at | timestamptz | NOT NULL | |

### 3.3 Sessions and Events

#### `development_sessions`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NOT NULL | |
| title | text | NULL | optional |
| status | text | NOT NULL | `active`, `paused`, `ended` |
| started_at | timestamptz | NOT NULL | |
| ended_at | timestamptz | NULL | |
| duration_seconds | integer | NULL, >= 0 | computed on end/update |
| source | text | NOT NULL | `cli`, `web` |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

Indexes: `(user_id, started_at desc)`, `(project_id, started_at desc)`.

#### `activity_events`
Append-only event log.

| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | server event row ID |
| event_id | uuid | UNIQUE, NOT NULL | client-generated idempotency key |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NULL | some events may be user-global |
| session_id | uuid | FK development_sessions(id), NULL | |
| event_type | text | NOT NULL | e.g. `session.started`, `git.commit` |
| occurred_at | timestamptz | NOT NULL | |
| source | text | NOT NULL | `cli`, `backend`, `web` |
| payload | jsonb | NOT NULL | sanitized metadata |
| schema_version | integer | NOT NULL | payload versioning |
| ingested_at | timestamptz | NOT NULL | |

Indexes: `(user_id, occurred_at desc)`, `(project_id, occurred_at desc)`, `(session_id, occurred_at desc)`, `(event_type, occurred_at desc)`, GIN on `payload` only if needed.

**Payload rules:** no source code, secrets, env vars, raw sensitive file contents. Payload schemas validated per `event_type`.

### 3.4 Git Activity

Git commits can be represented as events; optionally also maintain a query table for efficient UI.

#### `git_commits`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NOT NULL | |
| session_id | uuid | FK development_sessions(id), NULL | |
| commit_hash | text | NOT NULL | |
| branch | text | NULL | |
| message_summary | text | NULL | optional, sanitized/truncated |
| author_name | text | NULL | consider privacy setting |
| author_email | text | NULL | consider privacy setting |
| committed_at | timestamptz | NOT NULL | |
| files_changed | integer | NULL | |
| insertions | integer | NULL | |
| deletions | integer | NULL | |
| event_id | uuid | UNIQUE, FK-like to activity_events.event_id | traceability |
| created_at | timestamptz | NOT NULL | |

Unique: `(project_id, commit_hash)`.

### 3.5 Debugging Journal

#### `investigations`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NOT NULL | |
| session_id | uuid | FK development_sessions(id), NULL | |
| title | text | NOT NULL | |
| problem_description | text | NOT NULL | |
| observed_behavior | text | NULL | |
| investigation_steps | text | NULL | markdown/plain text |
| attempted_approaches | text | NULL | |
| root_cause | text | NULL | user-documented, not inferred |
| resolution | text | NULL | |
| lessons_learned | text | NULL | |
| prevention_strategies | text | NULL | |
| status | text | NOT NULL | `open`, `resolved`, `abandoned` |
| severity | text | NULL | optional |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |
| resolved_at | timestamptz | NULL | |

#### `investigation_related_files`
| Column | Type | Constraints |
|---|---|---|
| id | uuid | PK |
| investigation_id | uuid | FK investigations(id), NOT NULL |
| file_path | text | NOT NULL | sanitized relative path/label |
| created_at | timestamptz | NOT NULL |

#### `investigation_links`
For associating related problems.
| Column | Type | Constraints |
|---|---|---|
| investigation_id | uuid | FK investigations(id) |
| related_investigation_id | uuid | FK investigations(id) |
| relation_type | text | NOT NULL | `related`, `duplicate`, `regression` |
| created_at | timestamptz | NOT NULL |

PK: `(investigation_id, related_investigation_id)` with check preventing self-link.

### 3.6 Engineering Decisions

#### `decisions`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NOT NULL | |
| title | text | NOT NULL | |
| context | text | NOT NULL | |
| problem | text | NOT NULL | |
| alternatives | text | NULL | markdown/text |
| selected_approach | text | NOT NULL | |
| reasoning | text | NOT NULL | |
| consequences | text | NULL | |
| status | text | NOT NULL | `proposed`, `accepted`, `superseded`, `deprecated` |
| decided_at | timestamptz | NOT NULL | |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

### 3.7 Technical Notes

#### `notes`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NULL | global or project-scoped |
| session_id | uuid | FK development_sessions(id), NULL | |
| title | text | NOT NULL | |
| body | text | NOT NULL | |
| tags | text[] | NOT NULL default '{}' | simple MVP tagging |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

### 3.8 Testing History

#### `test_runs`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NOT NULL | |
| session_id | uuid | FK development_sessions(id), NULL | |
| command | text | NOT NULL | sanitized command label |
| status | text | NOT NULL | `passed`, `failed`, `errored`, `skipped`, `unknown` |
| started_at | timestamptz | NOT NULL | |
| completed_at | timestamptz | NULL | |
| duration_ms | bigint | NULL | |
| total_tests | integer | NULL | |
| passed_tests | integer | NULL | |
| failed_tests | integer | NULL | |
| failure_summary | text | NULL | sanitized, no secrets/raw logs by default |
| event_id | uuid | UNIQUE, NULL | traceability |
| created_at | timestamptz | NOT NULL | |

### 3.9 Insights and Reports

#### `insights`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NULL | global or project |
| insight_type | text | NOT NULL | e.g. `recurring_test_failure` |
| scope | text | NOT NULL | `global`, `project` |
| period_start | timestamptz | NOT NULL | |
| period_end | timestamptz | NOT NULL | |
| summary | text | NOT NULL | |
| observation | text | NOT NULL | factual |
| evidence | jsonb | NOT NULL | query results, counts, IDs |
| recommendation | text | NULL | clearly labeled recommendation |
| confidence | text | NOT NULL | `low`, `medium`, `high` - rule-based rubric |
| generated_at | timestamptz | NOT NULL | |

#### `reports`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| project_id | uuid | FK projects(id), NULL | |
| report_type | text | NOT NULL | `daily`, `weekly`, `project` |
| period_start | timestamptz | NOT NULL | |
| period_end | timestamptz | NOT NULL | |
| format | text | NOT NULL | `json`, `markdown` |
| content | text | NOT NULL | rendered export |
| metadata | jsonb | NOT NULL | counts/evidence refs |
| generated_at | timestamptz | NOT NULL | |

Reports may be generated on demand initially; persisting them is optional/cache-oriented.

### 3.10 Sync and CLI

#### `cli_devices` (optional but useful)
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| name | text | NOT NULL | user-provided device name |
| token_hash | text | UNIQUE, NOT NULL | device/CLI token |
| last_seen_at | timestamptz | NULL | |
| revoked_at | timestamptz | NULL | |
| created_at | timestamptz | NOT NULL | |

#### `sync_cursors`
| Column | Type | Constraints | Notes |
|---|---|---|---|
| id | uuid | PK | |
| user_id | uuid | FK users(id), NOT NULL | |
| device_id | uuid | FK cli_devices(id), NOT NULL | |
| last_event_ingested_at | timestamptz | NULL | |
| last_event_id | uuid | NULL | |
| updated_at | timestamptz | NOT NULL | |

### 3.11 Local SQLite Schema (CLI)

SQLite mirrors a subset for local operation and sync queue.

- `local_config` — API URL, device ID, project mappings, privacy settings.
- `projects` — local registered repos/projects.
- `sessions` — local session state.
- `events` — append-only local events with `sync_status` (`pending`, `syncing`, `synced`, `failed`), `event_id`, `payload_json`.
- `investigations`, `decisions`, `notes`, `test_runs` — local records with sync metadata.
- `sync_state` — last push/pull cursors, last error.

All synced records carry stable UUIDs generated locally at creation time.

---

## 4. Proposed API Structure

All endpoints under `/api/v1`. Authenticated endpoints require `Authorization: Bearer <token>` unless noted. Responses are JSON.

### 4.1 Authentication and User

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/auth/register` | No | Create user account |
| POST | `/auth/login` | No | Authenticate and issue session/token |
| POST | `/auth/logout` | Yes | Revoke current session |
| GET | `/me` | Yes | Current user profile |
| PATCH | `/me` | Yes | Update profile/preferences |
| GET | `/me/privacy` | Yes | Get privacy/data collection settings |
| PATCH | `/me/privacy` | Yes | Update privacy settings |

### 4.2 Projects

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/projects` | Yes | Register/create project |
| GET | `/projects` | Yes | List user's projects |
| GET | `/projects/{projectId}` | Yes | Project details |
| PATCH | `/projects/{projectId}` | Yes | Update project |
| DELETE | `/projects/{projectId}` | Yes | Archive/delete project (soft-delete preferred) |
| GET | `/projects/{projectId}/summary` | Yes | Project dashboard summary |

### 4.3 Sessions

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/sessions` | Yes | Start session |
| GET | `/sessions` | Yes | List sessions (filter by project, status, date) |
| GET | `/sessions/{sessionId}` | Yes | Session detail with activity summary |
| PATCH | `/sessions/{sessionId}/pause` | Yes | Pause active session |
| PATCH | `/sessions/{sessionId}/resume` | Yes | Resume paused session |
| PATCH | `/sessions/{sessionId}/end` | Yes | End session |
| GET | `/sessions/{sessionId}/events` | Yes | Events in session |

### 4.4 Activity / Timeline

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/activity` | Yes | Global activity timeline, filterable |
| GET | `/projects/{projectId}/activity` | Yes | Project activity timeline |
| GET | `/activity/{eventId}` | Yes | Event detail |

### 4.5 Git Activity

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/git/commits` | Yes | Global commit history |
| GET | `/projects/{projectId}/git/commits` | Yes | Project commits |
| GET | `/projects/{projectId}/git/branches` | Yes | Known branch activity (derived) |

### 4.6 Debugging Journal

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/investigations` | Yes | Create investigation |
| GET | `/investigations` | Yes | List/search investigations |
| GET | `/investigations/{id}` | Yes | Investigation detail |
| PATCH | `/investigations/{id}` | Yes | Update investigation |
| POST | `/investigations/{id}/links` | Yes | Link related investigations |
| DELETE | `/investigations/{id}/links/{relatedId}` | Yes | Remove link |
| GET | `/projects/{projectId}/investigations` | Yes | Project investigations |

### 4.7 Engineering Decisions

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/decisions` | Yes | Create decision record |
| GET | `/decisions` | Yes | List decisions |
| GET | `/decisions/{id}` | Yes | Decision detail |
| PATCH | `/decisions/{id}` | Yes | Update decision/status |
| GET | `/projects/{projectId}/decisions` | Yes | Project decisions |

### 4.8 Technical Notes

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/notes` | Yes | Create note |
| GET | `/notes` | Yes | List/search notes |
| GET | `/notes/{id}` | Yes | Note detail |
| PATCH | `/notes/{id}` | Yes | Update note |
| DELETE | `/notes/{id}` | Yes | Delete note |
| GET | `/projects/{projectId}/notes` | Yes | Project notes |

### 4.9 Testing History

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/test-runs` | Yes | Record test run (CLI/web) |
| GET | `/test-runs` | Yes | Global test history |
| GET | `/test-runs/{id}` | Yes | Test run detail |
| GET | `/projects/{projectId}/test-runs` | Yes | Project test history |

### 4.10 Insights

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/insights` | Yes | Global insights |
| POST | `/insights/generate` | Yes | Generate rule-based insights for scope/period |
| GET | `/insights/recurring-problems` | Yes | Recurring problem patterns |
| GET | `/insights/improvement-areas` | Yes | Improvement opportunities |
| GET | `/projects/{projectId}/insights` | Yes | Project insights |

### 4.11 Reports

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/reports/daily` | Yes | Daily report (query by date) |
| GET | `/reports/weekly` | Yes | Weekly report (query by week) |
| GET | `/reports/projects/{projectId}` | Yes | Project report |
| GET | `/reports/daily/export` | Yes | Export as `format=md|json` |
| GET | `/reports/weekly/export` | Yes | Export as `format=md|json` |
| GET | `/reports/projects/{projectId}/export` | Yes | Export project report |

### 4.12 CLI Sync

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/sync/devices` | Yes | Register CLI device |
| POST | `/sync/events` | Yes (device/user) | Batch ingest events; idempotent by `event_id` |
| GET | `/sync/pull` | Yes | Pull server-side changes/cursors (optional MVP) |
| GET | `/sync/status` | Yes | Last sync, pending conflicts/errors |

**Sync request shape (conceptual):**

```json
{
  "device_id": "uuid",
  "events": [
    {
      "event_id": "uuid",
      "event_type": "git.commit",
      "occurred_at": "2026-11-23T10:00:00Z",
      "project_local_id": "uuid",
      "session_local_id": "uuid",
      "payload": {},
      "schema_version": 1
    }
  ]
}
```

**Sync response:** per-event `accepted | duplicate | rejected` with reasons.

---

## 5. Proposed CLI Command Structure

Binary name: `devtrail`.

### 5.1 Global Behavior

- `devtrail --help`, `devtrail <command> --help`
- `--json` for machine-readable output where useful.
- `--project <name|id|path>` to override project detection.
- `--config <path>` optional config override.
- Exit codes: `0` success, non-zero for errors.

### 5.2 Commands

#### Auth / Config
| Command | Purpose |
|---|---|
| `devtrail init` | Initialize local config/SQLite DB in user config dir |
| `devtrail auth login` | Authenticate against backend; store token securely |
| `devtrail auth logout` | Remove local credentials / revoke session |
| `devtrail auth status` | Show auth and API endpoint status |
| `devtrail config get <key>` | Read config value |
| `devtrail config set <key> <value>` | Set config value (non-secret) |

#### Projects
| Command | Purpose |
|---|---|
| `devtrail project add [path]` | Register current or specified repo as project |
| `devtrail project list` | List registered local projects |
| `devtrail project info [project]` | Show project details |
| `devtrail project remove <project>` | Unregister local project (does not delete remote history by default) |

#### Sessions
| Command | Purpose |
|---|---|
| `devtrail session start [--project ...] [--title ...]` | Start dev session |
| `devtrail session pause` | Pause active session |
| `devtrail session resume` | Resume paused session |
| `devtrail session end [--summary ...]` | End active session |
| `devtrail session status` | Show active session |
| `devtrail session list [--project ...]` | List sessions |

#### Git / Activity
| Command | Purpose |
|---|---|
| `devtrail git sync [--project ...]` | Collect recent Git metadata and record events |
| `devtrail activity log [--project ...] [--limit ...]` | Show recent local activity |
| `devtrail activity add --type ... --note ...` | Manual activity entry |

#### Testing
| Command | Purpose |
|---|---|
| `devtrail test run -- <command...>` | Run test command, capture sanitized result metadata |
| `devtrail test list [--project ...]` | List recent local test runs |

#### Debugging Journal
| Command | Purpose |
|---|---|
| `devtrail debug new` | Create investigation (interactive prompts or flags) |
| `devtrail debug list [--project ...] [--status ...]` | List investigations |
| `devtrail debug show <id>` | Show investigation |
| `devtrail debug update <id>` | Update investigation fields |
| `devtrail debug resolve <id>` | Mark resolved with resolution/root cause |

#### Decisions / Notes
| Command | Purpose |
|---|---|
| `devtrail decision new` | Create decision record |
| `devtrail decision list [--project ...]` | List decisions |
| `devtrail note new` | Create technical note |
| `devtrail note list [--project ...]` | List notes |

#### Sync / Diagnostics
| Command | Purpose |
|---|---|
| `devtrail sync` | Push pending local records/events; pull if supported |
| `devtrail sync status` | Show queue size, last sync, failures |
| `devtrail doctor` | Check config, DB, git, auth, API connectivity |

### 5.3 Privacy/Safety Flags (MVP candidates)

- `--no-sync` for local-only commands where applicable.
- `--redact-message` / config to avoid storing commit messages.
- `--allow-remote-url` explicit opt-in for storing remote URL.

---

## 6. Proposed Frontend Page Structure

The web app should follow the navigation structure specified in the assignment brief.

### 6.1 Navigation Map

#### Workspace
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Overview | `pages/overview.html` | Global engineering dashboard | registered projects, recent sessions, recent git activity, recent investigations, testing activity, observations, timeline preview |
| Activity Timeline | `pages/activity.html` | Chronological global/project-filterable activity | event feed, filters by type/project/date |

#### Development
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Projects | `pages/projects.html` | Project list and project detail entry points | project cards/table, status, recent activity |
| Sessions | `pages/sessions.html` | Session history and details | session duration, status, associated activities |
| Git Activity | `pages/git-activity.html` | Commit/branch activity view | commits, branches, project filters |
| Testing History | `pages/testing.html` | Test runs and failures | command, result, duration, failure summary |

#### Engineering Knowledge
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Debugging Journal | `pages/debugging.html` | Investigation list/detail/editor | problem, steps, root cause, resolution, lessons |
| Decisions | `pages/decisions.html` | Engineering decision records | context, alternatives, reasoning, status |
| Technical Notes | `pages/notes.html` | Lightweight technical notes | title, body, tags, project/session links |

#### Intelligence
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Engineering Insights | `pages/insights.html` | Rule-based insight feed | observation, evidence, recommendation |
| Recurring Problems | `pages/recurring-problems.html` | Repeated investigation/failure patterns | counts, related investigations, evidence |
| Improvement Areas | `pages/improvement-areas.html` | Actionable improvement opportunities | testing gaps, regressions, hotspots |

#### Reports
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Daily Reports | `pages/reports-daily.html` | Daily summary and export | activities, sessions, problems, tests |
| Weekly Reports | `pages/reports-weekly.html` | Weekly summary and export | patterns, decisions, tests, opportunities |
| Project Reports | `pages/reports-project.html` | Project-level report | evolution, debugging, testing, decisions |

#### Settings
| Page | Route/File | Purpose | Key Data |
|---|---|---|---|
| Preferences | `pages/preferences.html` | UI and account preferences | display name, timezone, defaults |
| Integrations | `pages/integrations.html` | CLI/device/GitHub integration status | device tokens, sync status, future GitHub |
| Data and Privacy | `pages/data-privacy.html` | Collection/sync controls | retention, export/delete, redaction options |

### 6.2 Shared UI Components

- App shell: sidebar nav, page header, user menu.
- Stat/summary cards with clear labels and empty states.
- Timeline list with event type icons/badges.
- Filter bar: project, date range, status, type.
- Data table with pagination/cursor load-more.
- Detail drawer/panel or dedicated detail section.
- Forms with inline validation and accessible labels.
- Evidence block for insights: "Observed", "Evidence", "Possible explanation", "Recommendation".

### 6.3 Frontend Constraints

- No React/Vue/Angular/build-step dependency required for MVP.
- Vanilla JS modules (`type="module"`) for maintainability.
- Responsive layout for desktop-first, usable on smaller screens.
- Accessible semantics: landmarks, labels, focus states, keyboard navigation.

---

## 7. Implementation Roadmap

The roadmap below aligns the assignment's phases with incremental, testable delivery.

### Phase 1 — Repository Assessment and Requirements (current)
- [x] Inspect repository.
- [x] Document current state and gaps.
- [x] Propose architecture, schema, API, CLI, frontend, roadmap, risks.
- **Exit criteria:** assessment reviewed; key architectural decisions agreed.

### Phase 2 — Architecture and Foundation
- Initialize Git repository and `.gitignore`.
- Decide module strategy (`go.mod` location(s), module path).
- Create backend/CLI/frontend directory structure.
- Add config loading, logging, error conventions, Makefile/scripts.
- Add migration tooling and initial empty migration scaffold.
- Document ADRs for key decisions (auth, sync idempotency, event model, deployment).
- **Exit criteria:** project boots locally; conventions documented.

### Phase 3 — Backend Foundation
- DB connection, migrations, health endpoint.
- User registration/login/logout/profile.
- Project CRUD with ownership checks.
- Middleware: auth, request logging, validation, error handling.
- **Exit criteria:** authenticated user can create/list projects via API; tests pass.

### Phase 4 — CLI Foundation
- CLI command framework.
- Local config and SQLite store.
- `init`, `auth login/status`, `project add/list`.
- Git repo detection and safe metadata extraction.
- **Exit criteria:** CLI can register a repo locally and authenticate.

### Phase 5 — Activity Collection and Sync
- Event model and local event store.
- Session start/pause/resume/end.
- Git commit/branch metadata capture.
- Test run capture wrapper.
- Sync endpoint and CLI `sync` with idempotency.
- **Exit criteria:** local events sync to backend and appear in timeline APIs.

### Phase 6 — Engineering Knowledge
- Debugging investigations CRUD (API + CLI + web).
- Decisions CRUD.
- Technical notes CRUD.
- Search/filter and related-investigation linking.
- **Exit criteria:** user can document and retrieve investigations/decisions/notes.

### Phase 7 — Frontend
- App shell, auth pages, overview.
- Projects, sessions, git activity, testing pages.
- Debugging, decisions, notes pages.
- Settings pages.
- **Exit criteria:** core flows usable end-to-end against real APIs.

### Phase 8 — Engineering Intelligence
- Rule-based insight generators.
- Recurring problems and improvement areas views.
- Evidence/explanation rendering.
- **Exit criteria:** insights are reproducible and evidence-backed.

### Phase 9 — Reports
- Daily/weekly/project report generation.
- Markdown/JSON export.
- **Exit criteria:** reports reflect stored data accurately.

### Phase 10 — Testing, Hardening, Deployment
- Unit/integration/API/CLI/frontend smoke tests.
- Security review: authZ, input validation, secret redaction, rate limits.
- Render deployment config and runbooks.
- **Exit criteria:** MVP deployable, documented, and tested.

---

## 8. Technical Risks and Architectural Decisions

### 8.1 Key Decisions Needed Before Implementation

| Decision | Options | Recommendation | Rationale |
|---|---|---|---|
| Go module layout | Single module at repo root; separate `backend/go.mod` + `cli/go.mod`; workspace `go.work` | Start with **separate modules** for `backend` and `cli`, or a single root module with clear `cmd/backend`, `cmd/cli`. Prefer **single root module** for MVP simplicity unless release independence is needed. | Current `backend/go.mod` already exists; changing later is possible but noisy. Decide now. |
| Module path | `devtrail` vs `github.com/<org>/devtrail` | Use a fully qualified module path when repo location is known; otherwise keep local but plan migration. | Affects imports, tooling, and CI. |
| Auth mechanism | Opaque server sessions; JWT access+refresh | **Opaque server-side sessions/tokens** for MVP. | Simpler revocation, easier CLI device management. |
| API contract style | OpenAPI-first vs code-first | Code-first with generated/maintained OpenAPI doc early. | Keeps speed but avoids undocumented drift. |
| Event storage | Single `activity_events` JSONB payload vs typed tables per event | **Hybrid:** append-only `activity_events` + typed read tables for high-use domains (commits, test runs). | Preserves history while keeping queries simple. |
| Sync conflict strategy | Last-write-wins, CRDTs, field-level merge | **Idempotent append-only events + last-write-wins for mutable records with `updated_at` checks**. | MVP-simple; document limitations. |
| Commit message/privacy | Store full message, summary, hash only, opt-in | **Default to no/raw-minimal message storage**; allow opt-in summary. | Commit messages can contain sensitive references. |
| Frontend serving | Separate static site vs Go-served static files | **Go backend serves static frontend** for MVP; can split later. | Simpler deployment and CORS/auth cookie handling. |
| ID strategy | UUIDv4 vs ULID | **UUIDv4** or UUIDv7; prefer UUIDv7 if available/consistent for time ordering. | Stable IDs across CLI/backend. |
| Background jobs | In-process goroutines, cron, external queue | **In-process scheduled jobs** for MVP; abstract job runner. | Avoid infra complexity. |

### 8.2 Technical Risks

| Risk | Impact | Mitigation |
|---|---|---|
| **No Git history / no initialized repo** | Harder to enforce incremental commits and traceability | Initialize Git in Phase 2; establish commit conventions early. |
| **Schema over-design before real data** | Slows MVP, creates migration churn | Start with core tables; add read models only when query pain is real. |
| **Event payload inconsistency** | Ingestion bugs, unreliable insights | Define event schema versions and validators per event type. |
| **Sync duplication/conflicts** | Corrupt or duplicated timeline | Client-generated `event_id`, unique constraints, idempotent ingest transactions. |
| **Privacy leakage through metadata** | Sensitive paths/messages/commands stored | Redaction rules, allowlists, config controls, tests for sanitization. |
| **CLI command capture security** | `test run -- <cmd>` may expose secrets in args/env | Do not capture env; store command label carefully; document redaction/opt-out. |
| **Insights presented as certainty** | Misleading users | Enforce evidence fields and UI labels; no causal claims without evidence. |
| **Frontend without framework becomes inconsistent** | Spaghetti JS/CSS | Small component conventions, shared API/state modules, lint/format discipline. |
| **Single developer local paths as identity** | Privacy leak and unstable project identity | Use project IDs and repo fingerprints; avoid storing raw absolute paths centrally by default. |
| **Scope creep into AI/team features** | Delays MVP | Defer AI, team workspaces, deep GitHub integration until after validated data collection. |

### 8.3 Security Checklist for Phase 2+

- Password hashing: Argon2id or bcrypt with sane cost.
- Authorization middleware on every user-owned resource.
- Parameterized SQL only; no string-concatenated queries.
- Input validation at HTTP boundary and domain boundary where needed.
- Rate limiting on auth endpoints.
- Audit log/event for sensitive operations (token revoke, data export/delete).
- Secure defaults in config examples; never commit secrets.
- CORS restricted to known frontend origin in deployed environments.

### 8.4 Immediate Next Steps (Phase 2 Entry)

1. Initialize Git repo and add `.gitignore`.
2. Agree on module layout and module path.
3. Create `docs/architecture/` ADRs for auth, sync, event model, and deployment.
4. Scaffold backend `cmd/api`, config, database, migration runner, health endpoint.
5. Scaffold CLI `cmd/devtrail`, config, SQLite initialization.
6. Scaffold frontend shell with the required navigation structure.
7. Add Makefile/task scripts for `fmt`, `test`, `lint`, `run`.

---

## Appendix A — Proposed Initial Event Types

| Event Type | Source | Description |
|---|---|---|
| `project.registered` | CLI/Web | Project registered |
| `session.started` | CLI/Web | Session started |
| `session.paused` | CLI/Web | Session paused |
| `session.resumed` | CLI/Web | Session resumed |
| `session.ended` | CLI/Web | Session ended |
| `git.commit` | CLI | Commit metadata recorded |
| `git.branch_changed` | CLI | Branch checkout/change detected |
| `file.changed` | CLI | File change metadata recorded (path/counts only) |
| `test.run` | CLI/Web | Test execution recorded |
| `investigation.created` | CLI/Web | Investigation created |
| `investigation.updated` | CLI/Web | Investigation updated |
| `decision.created` | CLI/Web | Decision recorded |
| `note.created` | CLI/Web | Technical note created |

Event types should be versioned and documented in `docs/architecture/event-catalog.md` during Phase 2.
