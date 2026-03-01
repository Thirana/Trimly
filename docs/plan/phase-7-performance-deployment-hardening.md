# Phase 7: Performance Engineering and Deployment Hardening

## Goal

Move from feature-complete service to production-hardened system with measurable performance and reliable deployment.

## Scope

1. Performance profiling and tuning.
2. Runtime hardening and failure handling.
3. CI/CD quality gates.
4. Infrastructure and deployment baseline.

## Implementation approach

1. Performance engineering:
- profile CPU, memory, and allocation hotspots
- optimize redirect path first
- validate changes with repeatable benchmarks/load tests
2. Runtime hardening:
- strict server timeouts
- graceful shutdown and drain behavior
- safe defaults for connection pools
3. CI/CD:
- format and test checks
- race detector where applicable
- migration checks for DB changes
- build artifacts with immutable version tags
4. Infra baseline:
- environment-specific config management
- secret handling via runtime secret store
- health/readiness endpoints aligned with orchestrator

## Deployment model notes

1. Containerize API with minimal base image.
2. Use rolling/blue-green strategy where possible.
3. Add rollback path tied to previous stable image and schema compatibility rules.

## Observability and operations

1. Alerting on:
- latency regressions
- error-rate spikes
- DB/Redis saturation
2. Basic runbooks:
- incident triage
- rollback steps
- partial outage handling (DB/Redis degradation)

## Testing and validation

1. Pre-release load test against production-like environment.
2. Failure injection tests:
- Redis unavailable
- DB slow/unavailable
- high error traffic
3. Validate SLO targets and capture release notes.
4. Validate cache hot-path command efficiency:
- no `EXISTS` in redirect read path
- command sequence remains `GET short` -> `GET miss` on short-key miss
- `MGET` restricted to non-request batch tooling
5. Validate Redis reliability safeguards:
- Redis path can be disabled via config without API behavior drift
- stampede mitigation behavior under concurrent same-code misses
- skewed hot-key workload does not collapse latency budgets

## Risks and mitigations

1. Tuning without measurement:
- require benchmark/load-test evidence for performance claims.
2. Deployment regressions:
- canary rollout and automated rollback triggers.
3. Secret leakage:
- no secrets in code or logs; rotate keys regularly.

## Exit criteria

1. Performance baseline and regression budget are documented.
2. CI/CD pipeline enforces core quality gates.
3. Deployment and rollback process is repeatable.
4. Service can be operated safely with defined alerts and runbooks.
