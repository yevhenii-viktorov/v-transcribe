# Local Testing Guide

## Prerequisites
- Docker and Docker Compose installed
- At least 4GB RAM available for Whisper model
- Port 8080 available

## Quick Start

1. **Setup configuration:**
   ```bash
   # Configuration files are created automatically by the test script
   # Or manually copy:
   cp config/secrets.env.example config/secrets.env
   ```

2. **Create necessary directories:**
   ```bash
   mkdir -p transcripts whisper-models
   ```

3. **Start the services:**
   ```bash
   docker-compose -f docker-compose.local.yml up --build
   ```

4. **Access the application:**
   - Open http://localhost:8080 in your browser
   - Paste a YouTube URL and click "Transcribe"

## What's Different in Local Setup

The `docker-compose.local.yml` file is a minimal setup that includes only:
- **api**: The Go backend service
- **whisper**: The transcription service
- **web**: Nginx serving the frontend

This avoids conflicts with your server setup which includes many other services.

## Troubleshooting

### Port 8080 already in use
Change the port mapping in `docker-compose.local.yml`:
```yaml
web:
  ports:
    - "3000:80"  # Change to any available port
```

### Out of memory errors
The Whisper model requires significant RAM. Try using a smaller model:
```yaml
whisper:
  environment:
    - WHISPER_MODEL=ggml-tiny-q5_0.bin  # Smaller model
```

### Build errors
Make sure you have the latest Docker version:
```bash
docker --version  # Should be 20.10 or higher
docker-compose --version  # Should be 2.0 or higher
```

## Logs
View logs for debugging:
```bash
# All services
docker-compose -f docker-compose.local.yml logs

# Specific service
docker-compose -f docker-compose.local.yml logs api
docker-compose -f docker-compose.local.yml logs whisper
```

## Stop Services
```bash
docker-compose -f docker-compose.local.yml down
```

## Clean Up
Remove containers and volumes:
```bash
docker-compose -f docker-compose.local.yml down -v
rm -rf transcripts whisper-models
```