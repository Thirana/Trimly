# Local Development Setup

## Prerequisites

1. Go toolchain installed.
2. Version compatible with `go.mod` (`go 1.25.0`).

## Install dependencies

```bash
go mod download
```

## Run the API

```bash
go run ./cmd/api
```

Server defaults:

1. Listens on `:8080` when `PORT` is not set.
2. Uses in-memory storage (`MemoryStore`), so data is lost on restart.

## Environment variables

1. `PORT`
- Purpose: HTTP port binding for the API process.
- Default: `8080`
- Example: `PORT=9090`

2. `BASE_URL`
- Purpose: base URL used to build `short_url` in `POST /v1/links` response.
- Default: `http://localhost:8080`
- Example: `BASE_URL=http://localhost:9090`

3. `GIN_MODE`
- Note: current code calls `gin.SetMode(gin.ReleaseMode)` on startup, so release mode is forced by code.
- Practical impact: setting `GIN_MODE` alone will not switch runtime mode unless the code path changes.

## Quick health check

```bash
curl -i http://localhost:8080/health
```

Expected:

1. `200 OK`
2. JSON body: `{"status":"ok"}`
