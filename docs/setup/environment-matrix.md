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
- Value source: Upstash Redis endpoint (TLS `rediss://` URL from Upstash console).

6. `REDIS_ENABLED`
- Controls whether Redis cache path is active.
- Default: `false`.
- Set to `true` only when `REDIS_URL` is configured and reachable.
- Runtime behavior: app performs startup Redis `PING` fail-fast check when enabled.

7. `REDIS_POSITIVE_TTL`
- TTL for positive cache keys (`v1:url:short:{code}`).
- Default: `10m`.

8. `REDIS_MISS_TTL`
- TTL for negative cache keys (`v1:url:miss:{code}`).
- Default: `45s`.

9. `REDIS_CONNECT_TIMEOUT`
- Startup connectivity timeout for Redis `PING`.
- Default: `3s`.

10. `REDIS_OP_TIMEOUT`
- Runtime operation timeout for Redis calls.
- Default: `150ms`.
- Backward compatibility: if unset, app falls back to `REDIS_TIMEOUT` when present.

11. `CACHE_METRICS_LOG_INTERVAL`
- Interval for periodic cache metrics log snapshots (`cache_metrics ...`).
- Default: `30s`.
- Set `0` to disable periodic cache metrics logs.

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
REDIS_URL=rediss://default:password@host:6379
REDIS_ENABLED=true
REDIS_POSITIVE_TTL=10m
REDIS_MISS_TTL=45s
REDIS_CONNECT_TIMEOUT=3s
REDIS_OP_TIMEOUT=150ms
CACHE_METRICS_LOG_INTERVAL=30s
```

## Example frontend `.env.local`

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_APP_BASE_URL=http://localhost:3000
```
