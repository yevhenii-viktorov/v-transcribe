# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
1. **First think through the problem, read the codebase for relevant files, and write a plan to `tasks/todo.md`.**

2. **The plan should have a list of todo items that you can check off as you complete them.**

3. **Before you begin working, check in with me and I will verify the plan.**

4. **Then, begin working on the todo items, marking them as complete as you go.**

5. **Please, every step of the way just give me a high‑level explanation of what changes you made.**

6. **Make every task and code change you do as simple as possible.**  
   *We want to avoid making any massive or complex changes. Every change should impact as little code as possible. Everything is about simplicity.*

7. **Finally, add a review section to the `todo.md` file with a summary of the changes you made and any other relevant information.**

8. **DO NOT BE LAZY. NEVER BE LAZY.**  
   *If there is a bug, find the root cause and fix it. No temporary fixes. You are a senior developer—never be lazy.*

9. **MAKE ALL FIXES AND CODE CHANGES AS SIMPLE AS HUMANLY POSSIBLE.**  
   *They should only impact necessary code relevant to the task and nothing else. Your goal is to not introduce any bugs. It’s all about simplicity.*

## Development Commands

### Local Development
```bash
# Start local development stack
./scripts/test_locally.sh

# Manual local development
docker-compose -f docker-compose.local.yml up --build

# Run Go API locally for development
cd api
go mod tidy
go run main.go
```

### Docker Commands
```bash
# Build and start all services
docker compose up -d --build

# View logs
docker compose logs -f
docker compose logs -f api
docker compose logs -f whisper

# Stop services
docker compose down

# Clean up (remove volumes)
docker compose down -v
```

### Testing and Debugging
```bash
# Test API endpoint via nginx proxy (local development)
curl -X POST http://localhost:8080/job \
  -H "Content-Type: application/json" \
  -d '{"url":"https://youtube.com/watch?v=example"}'

# Check job status
curl http://localhost:8080/job/{job-id}

# Direct API access (production only, when port 8081 is exposed)
curl http://localhost:8081/job/{job-id}
```

## Architecture Overview

This is a YouTube transcription service with three main components:

1. **Go API Backend** (`api/main.go`)
   - HTTP server on port 8081
   - Job management with in-memory storage
   - Downloads YouTube audio using yt-dlp + ffmpeg
   - Calls Whisper service for transcription
   - Saves transcripts to `/data` volume

2. **Whisper Service** 
   - OpenAI Whisper transcription via `go-whisper` container
   - Runs on separate container accessible at `http://whisper:80`
   - Uses configurable models (default: `ggml-tiny.bin` for fast processing)

3. **Frontend** (`frontend/index.html`)
   - Pure HTML/HTMX/Tailwind CSS
   - Real-time job progress polling
   - Copy to clipboard and download functionality
   - Served via nginx proxy

### Data Flow
1. User submits YouTube URL via frontend
2. API creates job, downloads audio with yt-dlp → ffmpeg pipeline
3. API sends WAV file to Whisper service via HTTP
4. Whisper returns transcript text
5. API saves transcript to file and updates job status
6. Frontend polls for completion and displays results

## Configuration Structure

### Environment Configuration
- `config/default.env` - Non-secret defaults (committed to repo)
- `config/secrets.env.example` - Template for secrets (committed to repo)  
- `config/secrets.env` - Actual secrets (ignored by git)

Environment files are loaded in order: defaults first, then secrets override defaults.

## Key Files and Structure

- `api/main.go` - Single-file Go backend with job management
- `frontend/index.html` - Complete HTMX frontend
- `docker-compose.yml` - Full production setup with many services
- `docker-compose.local.yml` - Minimal local development setup
- `nginx.conf` - Nginx configuration for frontend and file serving
- `docker/Dockerfile.api` - Go API container build
- `config/` - Environment configuration directory

## Development Notes

### Go Backend
- Uses minimal dependencies (only `github.com/google/uuid`)
- In-memory job storage (jobs lost on restart)
- Synchronous processing with goroutines
- Direct subprocess calls to `yt-dlp`, `ffmpeg`, and `curl`

### Local vs Production Setup
- Local setup (`docker-compose.local.yml`) runs on port 8080
- Production setup includes many additional services (Plex, Sonarr, etc.)
- Both use same core transcription services

### Common Issues
- Memory requirements: Whisper models need 4-16GB RAM
- Port conflicts: Local development uses 8080, API uses 8081
- Audio processing: Requires yt-dlp and ffmpeg in API container
- Network: Services communicate via Docker internal networking

### No Testing Framework
This project has no automated tests or linting setup. Manual testing is done via the test script and direct API calls.
