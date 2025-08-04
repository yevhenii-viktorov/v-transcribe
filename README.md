# YouTube Transcription Service

A self-hosted YouTube transcription service built with Go, HTMX, and Docker Compose. Downloads YouTube videos, extracts audio, and generates transcripts using OpenAI Whisper.

## Features

- 🎥 Download audio from YouTube videos using yt-dlp
- ⚡ **Fast transcription** using optimized Whisper tiny model (5-10x faster)
- 🎙️ High-quality transcription using OpenAI Whisper
- 📱 Beautiful, responsive web interface with HTMX
- 📁 Download both audio and transcript files
- 📋 Copy transcripts to clipboard with feedback
- 🐳 Easy deployment with Docker Compose
- 📊 Real-time progress tracking with separate audio/transcription phases
- 🔄 **Job persistence** - survives restarts and continues processing
- 🛡️ **Reliable processing** with retry logic and timeout handling

## Prerequisites

- Docker ≥ 25 & docker-compose plugin
- **2GB RAM minimum** for `tiny` model, 4GB recommended
- Ubuntu/Linux environment

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone <your-repo> && cd v-transcribe
   ```

2. **Start the services**:
   ```bash
   docker compose up -d --build
   ```

3. **Access the application**:
   Open http://localhost in your browser

4. **Use the service**:
   - Paste a YouTube URL
   - Click "Transcribe"
   - Wait for processing to complete
   - Copy text or download the transcript file

## Architecture

```
┌──────────┐    POST /job        ┌──────────────┐
│  Frontend│ ───────────────────▶│   Go API      │
│  (HTMX)  │  job-id JSON        │  container    │
└──────────┘                     │ (yt-dlp +     │
▲    GET /job/:id           │  ffmpeg)      │
│  transcript / download    └──────┬────────┘
│                                  │ WAV
┌────┴─────────┐  HTTP (OpenAI-style)   ▼
│ go-whisper   │◀───────────────────────┘
│ model server │      text / srt
└──────────────┘
```

## Services

- **nginx**: Serves frontend and handles file downloads
- **api**: Go backend for job management and orchestration  
- **whisper**: OpenAI Whisper transcription service

## Configuration

### Environment Configuration

The application uses a structured configuration approach:

- **`config/default.env`**: Non-secret defaults (committed to repo)
- **`config/secrets.env.example`**: Template for secrets (committed to repo)
- **`config/secrets.env`**: Your actual secrets (ignored by git)

Environment files are loaded in order: defaults first, then secrets override defaults.

### Initial Setup

```bash
# Copy secrets template (done automatically by scripts)
cp config/secrets.env.example config/secrets.env

# Edit with your actual values if needed
nano config/secrets.env
```

### GPU Support (Optional)

For faster transcription with NVIDIA GPU, uncomment the GPU section in `docker-compose.yml`:

```yaml
deploy:
  resources:
    reservations:
      devices:
        - driver: nvidia
          capabilities: [gpu]
```

### Model Size

Edit the Whisper model in `docker-compose.yml`:

```yaml
environment:
  - WHISPER_MODEL=ggml-small-q5_0.bin  # For less RAM
  - WHISPER_MODEL=ggml-large-q5_0.bin  # For better accuracy
```

## Development

### Local Go Development

```bash
cd api
go mod tidy
go run main.go
```

### Frontend Development

The frontend is pure HTML/HTMX/Tailwind - just edit `frontend/index.html` and refresh.

## Troubleshooting

### Common Issues

1. **Out of memory**: Use smaller Whisper model (`ggml-small-q5_0.bin`)
2. **YouTube download fails**: Check if yt-dlp needs updating
3. **Transcription fails**: Ensure Whisper service is running and accessible

### Logs

```bash
# View all logs
docker compose logs -f

# View specific service logs
docker compose logs -f api
docker compose logs -f whisper
```

### Cleanup

```bash
# Stop services
docker compose down

# Remove volumes (transcripts and models)
docker compose down -v
```

## File Structure

```
v-transcribe/
├── api/
│   └── main.go           # Go backend API
├── frontend/
│   └── index.html        # HTMX frontend
├── config/
│   ├── default.env       # Non-secret defaults
│   ├── secrets.env.example # Secrets template
│   └── secrets.env       # Your secrets (gitignored)
├── docker/
│   └── Dockerfile.api    # API container build
├── scripts/
│   ├── test_locally.sh   # Local development setup
│   └── cleanup.sh        # Cleanup script
├── docs/
│   └── initial.md        # Project requirements  
├── docker-compose.yml    # Full production setup
├── docker-compose.local.yml # Local development setup
├── nginx.conf           # Nginx configuration
├── go.mod               # Go dependencies
└── README.md           # This file
```

## API Endpoints

- `POST /job` - Submit transcription job
- `GET /job/{id}` - Get job status and results
- `GET /files/{filename}` - Download transcript files

## License

MIT License