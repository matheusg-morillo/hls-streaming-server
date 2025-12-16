# HLS Streaming Server

A simple HTTP server for streaming HLS (HTTP Live Streaming) video content.

## Requirements

- Go 1.25.5 or higher

## Development

### First Time Setup

1. Clone the repository
2. Install golangci-lint (project-local):
   ```bash
   make install-lint
   ```

### Running the Server

```bash
make run
```

The server will start on `http://localhost:8080/hls/`

### Building

```bash
make build
```

This creates a binary at `bin/hls-server`

### Linting

Run the linter:
```bash
make lint
```

Auto-fix issues where possible:
```bash
make lint-fix
```

## Project Structure

- `main.go` - HTTP server implementation
- `.golangci.yml` - Linter configuration
- `golangci-lint.mod` - Linter dependency management
- `.upload/` - HLS media files directory

## How It Works

The server serves HLS video files from the `.upload/` directory on the `/hls/` endpoint. Place your `.m3u8` playlist files and `.ts` video segments in the `.upload/` directory, then access them via `http://localhost:8080/hls/<filename>`.
