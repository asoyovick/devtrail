# ADR 0003: Event Model and CLI Sync

- **Status:** Accepted for MVP foundation
- **Date:** 2026-11-23

## Context

DevTrail captures development activity from the CLI and web UI. The CLI is local-first and must synchronize reliably with the backend. Events should preserve history, while some domain records (investigations, decisions, notes) are mutable.

## Decision

- Use an **append-only `activity_events`** table as the canonical activity history.
- Every event has a client-generated `event_id` (UUID), `event_type`, `occurred_at`, `source`, `schema_version`, and sanitized `payload`.
- Backend ingestion is **idempotent** by `event_id`; duplicates return an accepted/duplicate result, not an error.
- Use **typed read models** for high-query domains (e.g., git commits, test runs) when needed for efficient UI/API queries; they remain traceable to events.
- Mutable domain records use `updated_at` and last-write-wins conflict handling for MVP; conflicts are surfaced in sync responses when detected.
- Sync is batch-oriented and retryable; local SQLite tracks `pending/syncing/synced/failed`.

## Consequences

- History is auditable and replay-friendly.
- Idempotency prevents duplicate timeline entries on retries.
- Payload schemas must be versioned and validated per event type.
- Last-write-wins is simple but can lose concurrent edits; acceptable for personal-first MVP, with future improvement if collaboration is added.
