# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run          # Start server (requires .env with HLS_DIR)
make build        # Compile to bin/hls-server
make lint         # Run golangci-lint via go tool
make lint-fix     # Auto-fix lint issues
make fmt          # Format with gofmt + goimports
make fmt-check    # Check formatting without writing
make clean        # Remove bin/
```

No test suite exists yet (`golangci.yml` has `tests: false`). The linter is managed as a Go tool dependency via `tools.mod` / `tools.sum`, not installed globally.

**Environment:** The server requires a `.env` file with `HLS_DIR` pointing to the directory containing HLS media files (default: `./.upload`).

**HLS media generation:**
```bash
# Multi-resolution script (recommended)
./scripts/generate_hls.sh <input.mp4> [output_dir] [segment_seconds]

# Outputs: <output_dir>/master.m3u8 + <output_dir>/{360p,480p,720p,1080p}/
```

## Architecture

The project follows Clean Architecture with a strict inward dependency rule. Import direction:

```
port → controller → domain
port → adapter → domain
port → wire/out
adapter → wire/out
```

- **`domain/`** — Pure Go structs. No JSON tags, no dependencies on other internal packages.
- **`controller/`** — Business logic (use cases). Returns domain types.
- **`adapter/`** — Converts domain types to wire DTOs (`domain.X → wire/out.X`).
- **`wire/out/`** — JSON-serializable DTOs (structs with `json` tags). These exist so JSON tags never pollute domain models.
- **`port/`** — HTTP entry points. `http_in.go` defines the `Routes` map; `server.go` wires the mux, registers routes, and applies middleware.
- **`application/`** — Thin entry point: calls `port.SetupServer()` and starts `http.Server`.

**Adding a new endpoint** follows this flow:
1. Define domain type in `domain/`
2. Implement use case in `controller/`
3. Create DTO in `wire/out/` and conversion in `adapter/`
4. Add handler function + register in `Routes` map in `port/http_in.go`

## Middleware System

Middlewares use the idiomatic Go pattern: `type Middleware func(http.Handler) http.Handler`.

Key functions in `middleware/middleware.go`:
- `Chain(mws ...Middleware) Middleware` — composes a pipeline (executes left-to-right at call time, wraps in reverse so first listed runs outermost)
- `Then(mw, handler)` / `ThenFunc(mw, handlerFunc)` — apply a middleware to a specific handler

Current middleware stack (applied in `server.go`):
1. `Logger()` wraps the entire mux (global)
2. `CORS()` + static file serving applied to `/hls/` route via `UseStaticFiles()`

HLS files are served from `HLS_DIR` at `/hls/` with `Accept-Ranges: bytes` and `Cache-Control: no-cache` headers required for ABR (Adaptive Bitrate) switching.

## HLS / ABR Context

The `.upload/` directory holds pre-transcoded HLS content:
- `master.m3u8` — top-level playlist referencing all variants
- `{360p,480p,720p,1080p}/manifest.m3u8` — per-resolution playlists
- `*.ts` — MPEG-TS video segments (~10s each)

**360p is listed first in `master.m3u8`** — intentional for cold-start behavior. Video.js/VHS defaults to the lowest variant when no bandwidth history exists, preventing initial rebuffering.

Cache headers are intentionally differentiated: `.m3u8` playlists use `no-cache` (they can change); `.ts` segments are immutable and should use aggressive caching in future CDN phases.
