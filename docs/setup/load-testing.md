# Load Testing (Phase 4 Baseline)

This guide provides a minimal repeatable load-test baseline for redirect hot path.

## Tool

Use `k6` with script:

1. `scripts/load/redirect_baseline.js`

## Prerequisites

1. API is running locally.
2. DB contains at least one valid short code.
3. For Redis-on measurements:
- `REDIS_ENABLED=true`
- valid `REDIS_URL`
4. For Redis-off baseline:
- `REDIS_ENABLED=false`

## Install k6

macOS (Homebrew):

```bash
brew install k6
```

## Prepare a test code

Create one short link and save the returned code:

```bash
curl -s -X POST http://localhost:8080/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com/load-test"}'
```

## Run baseline test

```bash
BASE_URL=http://localhost:8080 \
REDIRECT_CODE=<code> \
VUS=20 \
DURATION=30s \
k6 run scripts/load/redirect_baseline.js
```

## Compare Redis off vs Redis on

1. Run once with `REDIS_ENABLED=false`.
2. Run once with `REDIS_ENABLED=true`.
3. Compare:
- `http_req_duration` (`p(50)`, `p(95)`, `p(99)`)
- `http_req_failed`
- request rate
- app `cache_metrics ...` logs (`short_hit`, `miss_hit`, `db_fallback`, `cache_error`)

## Notes

1. Script expects redirect status `302`.
2. Keep the same `REDIRECT_CODE`, `VUS`, and `DURATION` between runs for fair comparison.
3. Upstash free tier has command/month quotas, so keep repeated runs moderate.
