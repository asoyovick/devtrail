# ADR 0004: Privacy and Data Collection Defaults

- **Status:** Accepted for MVP foundation
- **Date:** 2026-11-23

## Context

DevTrail runs in developer environments and can observe sensitive information. The product principle is privacy by design: collect useful engineering metadata without collecting source code or secrets by default.

## Decision

- **Never collect by default:** source code contents, credentials, API keys, access tokens, environment variables, private config file contents, or raw sensitive test/log output.
- **Collect by default (when enabled):** repository/project identifiers, branch names, commit hashes, commit timestamps, file change metadata (paths/counts) as sanitized metadata, session timestamps, command labels/results for supported test runs, and manually entered investigations/decisions/notes.
- **Sensitive by configuration:** commit messages, remote URLs, author email/name, and raw command arguments. Default to minimal/redacted storage; require explicit user opt-in for richer capture.
- **CLI redaction:** sanitize before writing to local SQLite and before sync. Backend validates and rejects disallowed sensitive fields.
- **User controls:** provide settings for tracked projects, sync behavior, data export, and deletion/retention.

## Consequences

- Some insight quality is limited until users opt in to richer metadata; this is an acceptable tradeoff.
- Redaction rules must be tested and versioned.
- Documentation must clearly state what is collected, what is not, and how to configure it.
- Auditability is required for sensitive operations (export/delete/token revoke).
