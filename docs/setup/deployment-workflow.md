# Deployment Workflow

This file captures a simple, repeatable deployment path for this project.

## Provisioning order

1. Create Neon project and Postgres database.
2. Create Upstash Redis database.
3. Deploy backend service to Render.
4. Deploy Next.js frontend to Vercel.

## Backend deployment (Render)

1. Create a Render Web Service for the Go API repository.
2. Build/start command strategy:
- Build: `go build ./cmd/api`
- Start: binary or `go run ./cmd/api` (binary preferred for production-like setup)
3. Configure environment variables:
- `PORT` (Render usually injects this)
- `BASE_URL`
- `DATABASE_URL`
- `REDIS_URL` (when Phase 4 is active)
4. Configure health check path: `/health`.

## Database setup (Neon)

1. Capture `DATABASE_URL` from Neon.
2. Run migrations from this repository before switching production traffic:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
```

3. Validate unique constraints and indexes expected by Phase 3.

## Cache setup (Upstash Redis)

1. Capture `REDIS_URL`.
2. Prefer the TLS URL (`rediss://...`) from Upstash console.
3. Keep cache disabled until Phase 4 code path is ready.
4. Enable with fallback behavior validated first.
5. With `REDIS_ENABLED=true`, startup performs Redis `PING`; deployment should fail fast if URL/credentials are wrong.

## Frontend deployment (Vercel)

1. Connect the Next.js repository.
2. Set required env vars:
- `NEXT_PUBLIC_API_BASE_URL`
- `NEXT_PUBLIC_APP_BASE_URL`
3. Deploy and verify frontend requests hit the correct backend environment.

## Smoke test checklist

1. `GET /health` returns `200`.
2. `POST /v1/links` creates a link and returns expected JSON shape.
3. `GET /:code` redirects with `302`.
4. If expiry is used, expired links return `404 not_found`.

## Rollback strategy

1. Backend rollback: redeploy previous Render version.
2. Frontend rollback: promote previous Vercel deployment.
3. Data rollback: use migration down scripts only with explicit review.

## Change management

1. Any provider change must update:
- `docs/setup/deployment-targets.md`
- `docs/setup/environment-matrix.md`
- related phase plan docs
2. Major hosting/provider decision changes require a new ADR.
