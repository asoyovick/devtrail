# ADR 0002: Authentication and Session Model

- **Status:** Accepted for MVP foundation
- **Date:** 2026-11-23

## Context

DevTrail requires secure authentication for web users and CLI synchronization. The system is personal-first in MVP but must enforce ownership and protect against cross-user access.

Options considered: opaque server-side sessions/tokens, JWT access+refresh tokens.

## Decision

- Use **opaque, server-side session tokens** for MVP.
- Store only a hash of session tokens in PostgreSQL; never store raw tokens.
- Support session expiration and explicit revocation (logout, device revoke).
- CLI devices receive their own revocable device/session tokens through an authenticated registration flow.
- Passwords are hashed with a modern adaptive hash (Argon2id preferred; bcrypt acceptable if dependency/ops constraints require it).

## Consequences

- Revocation and device management are straightforward.
- Every authenticated request requires a session lookup; this is acceptable at MVP scale and can be optimized later with caching.
- JWT can be reconsidered if stateless scaling becomes a real need, but not before MVP validation.
- Authorization checks must still be applied on every user-owned resource regardless of auth mechanism.
