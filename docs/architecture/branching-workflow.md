# DevTrail Branching Workflow

## Permanent Branches

- **`main`**: Stable, deployable code only. No direct feature development.
- **`develop`**: Integration branch for completed, tested features.

## Working Branches

| Type | Pattern | Example |
|---|---|---|
| Feature | `feature/<module>-<description>` | `feature/auth-registration` |
| Bug fix | `fix/<module>-<description>` | `fix/sync-duplicate-events` |
| Documentation | `docs/<description>` | `docs/api-reference` |
| Refactor | `refactor/<module>-<description>` | `refactor/database-layer` |
| Hotfix | `hotfix/<description>` | `hotfix/auth-token-validation` |

## Standard Flow

1. Inspect repository state: `git status`, current branch, recent history.
2. Update `develop`: `git switch develop && git pull --ff-only origin develop`.
3. Create a focused branch: `git switch -c feature/<module>-<description>`.
4. Implement one logical change; run relevant tests/checks.
5. Review diff: `git diff`, `git diff --check`, `git status`.
6. Commit with Conventional Commits.
7. Push branch and open a focused PR into `develop`.
8. Merge to `main` only for stable, tested releases.

## Rules

- Do not commit directly to `main` during normal feature development.
- Do not force-push shared branches.
- Do not commit secrets or local environment artifacts.
- Do not fabricate Git history, contributors, or timestamps.
- Keep commits small, coherent, and accurately described.
