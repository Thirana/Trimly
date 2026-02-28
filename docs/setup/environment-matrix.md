# Environment Variable Matrix

This matrix defines runtime configuration for local and cloud environments.

## Backend (Go API)

1. `PORT`
- Local: optional (`8080` default).
- Cloud: required by host platform runtime contract.

2. `BASE_URL`
- Local: `http://localhost:8080`.
- Cloud: public API base URL (for `short_url` in create response).

3. `GIN_MODE`
- Current code sets release mode directly in `cmd/api/main.go`.
- Keep as informational until mode becomes config-driven.

4. `DATABASE_URL`
- Required when Postgres store is enabled (Phase 3).
- Value source: Neon connection string.

5. `REDIS_URL`
- Required when Redis cache path is enabled (Phase 4).
- Value source: Redis Cloud endpoint.

## Frontend (Next.js)

1. `NEXT_PUBLIC_API_BASE_URL`
- Local frontend to local API: `http://localhost:8080`.
- Cloud frontend to cloud API: public backend URL.

2. `NEXT_PUBLIC_APP_BASE_URL`
- Frontend public URL (useful for sharing, metadata, and links).

## Suggested environment files

1. Backend local: `.env` (not committed).
2. Frontend local: `.env.local` in Next.js repo (not committed).
3. Cloud: provider-managed environment variables only.

## Secrets handling rules

1. Never commit real secrets.
2. Keep per-environment values isolated.
3. Rotate credentials when sharing or exposing accidentally.

## Example backend `.env`

```bash
PORT=8080
BASE_URL=http://localhost:8080
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
REDIS_URL=redis://default:password@host:6379
```

## Example frontend `.env.local`

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_APP_BASE_URL=http://localhost:3000
```
