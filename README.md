# HLS Streaming Server

An HTTP server for HLS (HTTP Live Streaming) video content built in Go as part of a practical study plan to master streaming technologies.

## 📚 Project Context

This project is part of **Phase 1** of the "Streaming & Video Technologies" study plan, aimed at technical preparation for working at streaming companies like Netflix.

### Learning Objectives

- ✅ Implement basic HTTP server for HLS streaming
- ✅ Serve video segments (.ts) and playlists (.m3u8)
- 🔄 Understand streaming protocols (HLS)
- 🔄 Practice video conversion with FFmpeg
- 🔄 Work with HTML5 players (Video.js)
- 🔄 Measure latency and streaming performance

---

## 🚀 Getting Started

### Prerequisites

- Go 1.25.5 or higher
- FFmpeg (for video conversion)
- Docker and docker-compose (optional)

### Installation

1. Clone the repository
```bash
git clone <repo-url>
cd hls-streaming-server
```

2. Install linter (first time)
```bash
make install-lint
```

3. Create upload directory
```bash
mkdir -p .upload
```

---

## 🏃 Running the Server

### Development Mode

```bash
make run
```

Server will start at `http://localhost:8080`

### With Docker

```bash
docker-compose up --build
```

### Manual Build

```bash
make build
./bin/hls-server
```

---

## 🎮 Testing the Stream

### Check Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "Healthy",
  "time": "2025-01-15T10:30:00Z"
}
```

### Access HLS Playlist

```bash
curl http://localhost:8080/hls/stream.m3u8
```

---

## 🛠️ Available Commands

```bash
# Format code
make fmt

# Check formatting
make fmt-check

# Run linter
make lint

# Auto-fix linter issues
make lint-fix

# Build
make build

# Clean binaries
make clean

# See all commands
make help
```

---

## 📁 Project Structure

```
.
├── src/
│   ├── adapter/           # Data conversion (domain → JSON)
│   ├── application/       # Application entry point
│   ├── controller/        # Business logic
│   ├── domain/            # Domain entities
│   ├── middleware/        # HTTP middlewares
│   ├── port/              # HTTP adapters (in/out)
│   └── wire/              # DTOs (Data Transfer Objects)
├── .upload/               # HLS files (m3u8 + ts)
├── Dockerfile             # Docker container
├── docker-compose.yaml    # Orchestration
├── Makefile               # Development commands
├── .golangci.yml          # Linter configuration
└── README.md              # This file
```

### Architecture

The project follows **Clean Architecture** with clear separation of concerns:

- **Domain**: Business entities (pure models)
- **Controller**: Use cases and application logic
- **Adapter**: Conversion between layers
- **Port**: Input/output interfaces (HTTP handlers)
- **Middleware**: Cross-cutting concerns (CORS, logging, etc)

---

## 💡 Technologies

- **Language**: Go 1.25.5
- **Protocol**: HLS (HTTP Live Streaming)
- **Container**: Docker + docker-compose
- **Code Quality**: golangci-lint

---

## 📄 License

This project is for educational purposes.

---

**Remember**: This is a practical learning project focused on building streaming infrastructure skills. 🚀🎬
