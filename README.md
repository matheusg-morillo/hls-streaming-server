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

## Creating the m3u8 and ts files (simple)
```bash
ffmpeg -i <video_path> \
  -codec copy \
  -start_number 0 \
  -hls_time 10 \
  -hls_list_size 0 \
  -f hls output/output.m3u8
```

## Creating the m3u8 and ts files (for adaptive streaming)

**Common ffmpeg options:**
- `-b:v 2500k` - video bitrate
- `-maxrate/-bufsize` - prevent bitrate spikes (VBV buffer)
- `-s 1280x720` - output resolution
- `-c:a aac` - audio codec
- `-b:a 128k` - audio bitrate
- `-hls_time 10` - segment duration (10 seconds)
- `-hls_list_size 0` - keep all segments (0 = unlimited)
- `-hls_segment_filename` - naming pattern for .ts segments

### Individual Commands

#### 1080p
```bash
ffmpeg -i input.mp4 \
  -c:v libx264 \
  -b:v 5000k \
  -maxrate 5000k \
  -bufsize 10000k \
  -s 1920x1080 \
  -c:a aac \
  -b:a 128k \
  -hls_time 10 \
  -hls_list_size 0 \
  -hls_segment_filename "upload/1080p_%03d.ts" \
  "upload/1080p.m3u8"
```

#### 720p
```bash
ffmpeg -i input.mp4 \
  -c:v libx264 \
  -b:v 2500k \
  -maxrate 2500k \
  -bufsize 5000k \
  -s 1280x720 \
  -c:a aac \
  -b:a 128k \
  -hls_time 10 \
  -hls_list_size 0 \
  -hls_segment_filename "upload/720p_%03d.ts" \
  "upload/720p.m3u8"
```

#### 480p
```bash
ffmpeg -i input.mp4 \
  -c:v libx264 \
  -b:v 1200k \
  -maxrate 1200k \
  -bufsize 2400k \
  -s 854x480 \
  -c:a aac \
  -b:a 96k \
  -hls_time 10 \
  -hls_list_size 0 \
  -hls_segment_filename "upload/480p_%03d.ts" \
  "upload/480p.m3u8"
```

#### 360p
```bash
ffmpeg -i input.mp4 \
  -c:v libx264 \
  -b:v 600k \
  -maxrate 600k \
  -bufsize 1200k \
  -s 640x360 \
  -c:a aac \
  -b:a 64k \
  -hls_time 10 \
  -hls_list_size 0 \
  -hls_segment_filename "upload/360p_%03d.ts" \
  "upload/360p.m3u8"
```

### Unified Command (All Variants at Once)

Run this single command to generate all quality variants:

```bash
ffmpeg -i input.mp4 \
  -c:v libx264 -b:v 5000k -maxrate 5000k -bufsize 10000k -s 1920x1080 -c:a aac -b:a 128k \
    -hls_time 10 -hls_list_size 0 -hls_segment_filename "upload/1080p_%03d.ts" "upload/1080p.m3u8" \
  -c:v libx264 -b:v 2500k -maxrate 2500k -bufsize 5000k -s 1280x720 -c:a aac -b:a 128k \
    -hls_time 10 -hls_list_size 0 -hls_segment_filename "upload/720p_%03d.ts" "upload/720p.m3u8" \
  -c:v libx264 -b:v 1200k -maxrate 1200k -bufsize 2400k -s 854x480 -c:a aac -b:a 96k \
    -hls_time 10 -hls_list_size 0 -hls_segment_filename "upload/480p_%03d.ts" "upload/480p.m3u8" \
  -c:v libx264 -b:v 600k -maxrate 600k -bufsize 1200k -s 640x360 -c:a aac -b:a 64k \
    -hls_time 10 -hls_list_size 0 -hls_segment_filename "upload/360p_%03d.ts" "upload/360p.m3u8"
```

### Master Playlist (master.m3u8)

Create this file to reference all variants:

```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=5128000,RESOLUTION=1920x1080
1080p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2628000,RESOLUTION=1280x720
720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1328000,RESOLUTION=854x480
480p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=664000,RESOLUTION=640x360
360p.m3u8
```

Then stream using: `http://localhost:8080/hls/master.m3u8`
