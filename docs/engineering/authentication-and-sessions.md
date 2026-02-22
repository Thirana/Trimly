# Authentication and Sessions

## Current implementation status

1. Not implemented yet in this repository.
2. Current API is public for link creation and redirect.

## Planned implementation (Phase 5)

1. Signup/login/logout flows.
2. Secure session or token strategy.
3. Auth middleware that attaches user identity to request context.
4. Session expiration and revocation behavior.

## Key design decisions to make before implementation

1. Session model choice:
- cookie sessions
- access + refresh tokens
2. Token/session storage strategy.
3. Secret rotation policy.
4. Endpoint contract for auth failures (`401`) and forbidden actions (`403`).

## Security baseline for Phase 5

1. Never store plaintext passwords.
2. Hash passwords with a modern algorithm.
3. If cookies are used, enforce `HttpOnly`, `Secure` in prod, and explicit `SameSite`.
4. Add CSRF strategy when cookie auth is introduced.

## Related plan document

1. `docs/plan/phase-5-auth-profiles-link-management.md`
