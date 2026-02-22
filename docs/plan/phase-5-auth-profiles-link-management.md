# Phase 5: Auth, Profiles, and Link Management APIs

## Goal

Introduce authenticated user workflows with secure session handling and strict ownership enforcement.

## Scope

1. User signup/login/logout.
2. Profile endpoint(s) for authenticated identity.
3. Link management endpoints (list, update, delete) scoped to owner.
4. Session or token strategy with secure defaults.

## Implementation approach

1. Define auth model:
- users table
- password hashing (`bcrypt`/`argon2`)
- session store or signed token approach
2. Add auth middleware:
- authenticate request
- attach user identity to context
3. Implement ownership checks in service/store paths:
- every mutating or private read operation must validate owner
4. Add API endpoints:
- `POST /auth/signup`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /me`
- `GET /links`
- `PATCH /links/:id` or `/:code`
- `DELETE /links/:id` or `/:code`
5. Keep consistent error semantics for auth and authorization failures.

## Security notes

1. Never log passwords or session/token secrets.
2. Use secure cookie settings if cookie sessions are used:
- `HttpOnly`
- `Secure` in production
- explicit `SameSite`
3. Define CSRF approach before production rollout when using cookies.
4. Add brute-force protections on login endpoints (expanded in Phase 6).

## Testing plan

1. Unit tests for auth service rules.
2. Handler tests for signup/login/logout and protected endpoints.
3. Authorization tests:
- owner can access
- non-owner is blocked with `403`
4. Session lifecycle tests (login, expiry, logout invalidation).

## Risks and mitigations

1. Broken object level authorization:
- centralize ownership checks and test both positive and negative paths.
2. Session fixation or weak secrets:
- rotate secrets and set expiration policies.
3. API complexity growth:
- keep auth concerns isolated in middleware/service layers.

## Exit criteria

1. Authenticated identity works end-to-end.
2. Profile and link management APIs are owner-safe.
3. Security defaults are in place for sessions/tokens.
4. Tests cover auth flows and ownership boundaries.
