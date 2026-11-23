# Contributing to DevTrail

## Development Principles

- Build incrementally; do not implement the entire platform in one change.
- Keep changes focused, testable, and easy to review.
- Follow the agreed technology stack: Go backend/CLI, PostgreSQL/SQLite, HTML/CSS/Vanilla JS frontend.
- Do not commit secrets, credentials, environment files, or local databases.
- Do not fabricate Git history, contributors, timestamps, or test results.

## Branching Workflow

- `main` — stable, deployable branch. No direct feature development.
- `develop` — integration branch for completed, tested work.
- `feature/<module>-<description>` — new functionality.
- `fix/<module>-<description>` — bug fixes.
- `docs/<description>` — documentation changes.
- `refactor/<module>-<description>` — internal restructuring without intended behavior change.
- `hotfix/<description>` — critical production fix from `main`.

Create feature branches from the latest `develop`:

```bash
git switch develop
git pull --ff-only origin develop
git switch -c feature/<module>-<description>
```

## Commit Convention

Use Conventional Commits:

```text
<type>(<scope>): <description>
```

Allowed types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `build`, `ci`, `style`.

Examples:

- `feat(auth): implement user registration`
- `fix(sync): prevent duplicate event processing`
- `docs(api): document activity ingestion endpoints`

## Before Committing

- Inspect `git status` and `git diff`.
- Stage only relevant files; avoid blind `git add .` when unrelated changes exist.
- Run relevant tests and formatting/checks for the affected component.
- Ensure documentation is updated when behavior or setup changes.

## Pull Requests

PRs should include: summary, problem, implementation details, database/API changes, testing performed with actual results, security considerations, and limitations.

## Security

Never commit API keys, passwords, tokens, private credentials, `.env` files, local databases, or temporary debugging artifacts.
