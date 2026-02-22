# Authorization

## Current implementation status

1. Not implemented yet in this repository.
2. There are no owner-only link management endpoints yet.

## Planned implementation (Phase 5)

1. Ownership model for links and profile resources.
2. Middleware and service checks that enforce ownership.
3. Clear separation between:
- authentication (who the user is)
- authorization (what the user can access)

## Core rules for upcoming implementation

1. Deny by default for private resources.
2. Enforce ownership in service/store paths, not only handlers.
3. Return stable authorization status codes:
- `401` unauthenticated
- `403` authenticated but not allowed
4. Prevent resource enumeration when appropriate.

## Expected API surface after Phase 5

1. `GET /me`
2. `GET /links` (owner scoped)
3. `PATCH /links/:id` or `/:code` (owner scoped)
4. `DELETE /links/:id` or `/:code` (owner scoped)

## Related plan document

1. `docs/plan/phase-5-auth-profiles-link-management.md`
