# Security Hardening

## Current protections

1. URL input is constrained to valid `http/https` URLs with host checks.
2. Short code generation uses `crypto/rand`, reducing predictability.
3. Internal errors are not exposed to clients (`500 internal`).
4. Gin recovery middleware protects against panic crashes.

## Current code references

1. `internal/shortener/validate.go`
2. `internal/shortener/code.go`
3. `internal/httpapi/links_handlers.go`
4. `internal/httpapi/router.go`

## Current risk profile

1. No auth or authorization yet (acceptable for current phase).
2. No explicit rate limiting yet.
3. No request size limits yet.
4. No abuse controls for create endpoint yet.

## Planned hardening roadmap

1. Phase 5:
- secure auth sessions/tokens
- ownership enforcement for link management
2. Phase 6:
- endpoint rate limits
- input size limits
- abuse controls and monitoring
3. Phase 7:
- deployment hardening and secret management maturity

## Staff-level practice to keep

1. Treat redirect and creation endpoints as abuse surfaces early.
2. Avoid logging sensitive user or credential data.
3. Build security controls incrementally with explicit phase ownership.
