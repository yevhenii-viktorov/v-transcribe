# Server Deployment Guide

## 🚀 Quick Deployment

### 1. **Copy Code to Server**
```bash
# On your server, in your docker-compose directory
git clone https://github.com/yevhenii-viktorov/v-transcribe.git
cd v-transcribe

# Copy configuration
cp config/secrets.env.example config/secrets.env
# Edit config/secrets.env if you need custom values
```

### 2. **Add Services to Your Docker Compose**
```bash
# Copy the services from docker-compose.server.yml to your main docker-compose.yml
# Add the services under your existing services: section
# Add the volumes under your existing volumes: section
```

### 3. **Deploy**
```bash
# Build and start the new services
docker compose up -d --build v-transcribe-api v-transcribe-whisper v-transcribe-web

# Check status
docker compose ps | grep v-transcribe
```

### 4. **Setup Nginx Proxy Manager**
- **Domain**: `transcribe.yourdomain.com`
- **Forward to**: `v-transcribe-web:80` or `192.168.10.73:80`
- **SSL**: Enable with Let's Encrypt
- **Force SSL**: ON

---

## 📋 Detailed Integration Steps

### **Step 1: Prepare Your Server**

Make sure your existing docker-compose.yml has these networks (it should already):
```yaml
networks:
  npm_proxy:
    name: npm_proxy
    driver: bridge
    ipam:
      config:
        - subnet: 192.168.10.0/24
```

### **Step 2: Check IP Availability**

Check which IPs are already in use in your 192.168.10.0/24 subnet:
```bash
docker compose ps --format "table {{.Name}}\t{{.Networks}}"
grep -r "ipv4_address.*192.168.10" docker-compose.yml
```

**IPs used by transcription services:**
- `192.168.10.71` - v-transcribe-api
- `192.168.10.72` - v-transcribe-whisper  
- `192.168.10.73` - v-transcribe-web

If any conflict, update them in the server config before deploying.

### **Step 3: Environment Configuration**

Your existing `config/default.env` and `config/secrets.env` will be used automatically.

**Optional**: Add transcription-specific variables to `config/secrets.env`:
```bash
# Transcription Configuration (optional)
WHISPER_MODEL=ggml-tiny.bin
MAX_UPLOAD_SIZE=500M
TRANSCRIPTION_TIMEOUT=7200
```

### **Step 4: Volume Setup**

The services will create these Docker volumes:
- `youtube-whisper-models` - Whisper AI models (downloaded automatically)
- `youtube-transcripts` - Generated transcripts and audio files

**Optional**: Use bind mounts instead of volumes:
```yaml
volumes:
  - $DATADIR/transcribe/models:/data
  - $DATADIR/transcribe/transcripts:/data
```

### **Step 5: Deploy Services**

```bash
# Build and start transcription services
docker compose up -d --build v-transcribe-api v-transcribe-whisper v-transcribe-web

# Check logs
docker compose logs -f v-transcribe-api
docker compose logs -f v-transcribe-whisper

# Verify health
curl http://192.168.10.73  # Should return the web interface
curl http://192.168.10.72/api/v1/models  # Should return whisper models
```

---

## 🌐 Nginx Proxy Manager Setup

### **Create Proxy Host**

1. **Login to Nginx Proxy Manager** (usually at `yourdomain.com:81`)

2. **Add New Proxy Host**:
   - **Domain Names**: `transcribe.yourdomain.com`
   - **Scheme**: `http`
   - **Forward Hostname/IP**: `v-transcribe-web` (or `192.168.10.73`)
   - **Forward Port**: `80`
   - **Cache Assets**: ON
   - **Block Common Exploits**: ON
   - **Websockets Support**: OFF

3. **SSL Certificate**:
   - **SSL Certificate**: Request a new SSL Certificate with Let's Encrypt
   - **Force SSL**: ON
   - **HSTS Enabled**: ON
   - **HTTP/2 Support**: ON

4. **Advanced Configuration** (optional):
   ```nginx
   # Handle large file uploads
   client_max_body_size 500M;
   client_body_timeout 300s;
   proxy_read_timeout 300s;
   proxy_connect_timeout 300s;
   proxy_send_timeout 300s;
   
   # Security headers
   add_header X-Frame-Options "SAMEORIGIN" always;
   add_header X-Content-Type-Options "nosniff" always;
   add_header Referrer-Policy "no-referrer-when-downgrade" always;
   ```

### **Test Your Setup**

1. **Visit**: `https://transcribe.yourdomain.com`
2. **Submit a test YouTube URL**
3. **Monitor progress** in real-time
4. **Download transcript** when complete

---

## 🔧 Configuration Options

### **Memory Optimization**

For servers with limited RAM, use the tiny Whisper model:
```yaml
environment:
  - WHISPER_MODEL=ggml-tiny.bin  # ~400MB RAM usage
```

For better quality on servers with more RAM:
```yaml
environment:
  - WHISPER_MODEL=ggml-base.bin  # ~600MB RAM usage
  - WHISPER_MODEL=ggml-small.bin # ~800MB RAM usage
```

### **GPU Acceleration** (Optional)

If you have an NVIDIA GPU:
```yaml
v-transcribe-whisper:
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            capabilities: [gpu]
```

### **Storage Locations**

Default volume locations (managed by Docker):
```bash
# Find actual paths
docker volume inspect youtube-transcripts
docker volume inspect youtube-whisper-models
```

Custom storage paths:
```yaml
volumes:
  - /path/to/your/transcripts:/data
  - /path/to/your/models:/whisper-models
```

---

## 📊 Monitoring & Maintenance

### **Health Checks**

```bash
# Check service status
docker compose ps | grep v-transcribe

# Check service health
curl http://192.168.10.72/api/v1/models  # Whisper models
curl http://192.168.10.73/health || echo "Add health endpoint"

# Check logs
docker compose logs --tail 50 v-transcribe-api
docker compose logs --tail 50 v-transcribe-whisper
```

### **Resource Usage**

```bash
# Monitor memory/CPU usage
docker stats v-transcribe-api v-transcribe-whisper v-transcribe-web

# Check disk usage
docker system df
du -sh $(docker volume inspect youtube-transcripts --format '{{.Mountpoint}}')
```

### **Backup Important Data**

```bash
# Backup transcripts
docker run --rm -v youtube-transcripts:/data -v $(pwd):/backup alpine tar czf /backup/transcripts-backup.tar.gz /data

# Restore transcripts
docker run --rm -v youtube-transcripts:/data -v $(pwd):/backup alpine tar xzf /backup/transcripts-backup.tar.gz -C /
```

---

## 🚨 Troubleshooting

### **Common Issues**

1. **Service won't start**
   ```bash
   # Check IP conflicts
   docker compose ps | grep 192.168.10
   
   # Check logs
   docker compose logs v-transcribe-api
   ```

2. **Out of memory errors**
   ```bash
   # Use smaller Whisper model
   # Increase server RAM
   # Add swap space
   ```

3. **Proxy connection refused**
   ```bash
   # Check if web service is running
   docker compose ps v-transcribe-web
   
   # Test internal connectivity
   docker exec nginx-proxy-manager curl http://v-transcribe-web
   ```

4. **Transcription fails**
   ```bash
   # Check Whisper service health
   curl http://192.168.10.72/api/v1/models
   
   # Restart Whisper service if needed
   docker compose restart v-transcribe-whisper
   ```

### **Performance Tuning**

- **Memory**: Increase limits if you see OOM kills
- **Timeouts**: Increase for very long videos
- **Model**: Use `tiny` for speed, `base` for quality
- **Concurrency**: Only run one transcription at a time for now

---

## ✅ Post-Deployment Checklist

- [ ] Services are running: `docker compose ps | grep v-transcribe`
- [ ] Web interface accessible: `https://transcribe.yourdomain.com`
- [ ] Whisper models downloaded: Check logs for model download
- [ ] Test transcription: Submit a short YouTube video
- [ ] SSL certificate active: Check browser lock icon
- [ ] Monitoring setup: Add to your existing monitoring
- [ ] Backup strategy: Plan for transcript backups

Your YouTube transcription service is now deployed and ready for production use! 🎉