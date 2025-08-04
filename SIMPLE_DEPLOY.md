# Simple Server Deployment

## 🚀 One Command Deploy

### Step 1: Copy Code to Server
```bash
# In your docker directory where your main docker-compose.yml is
git clone <your-repo> v-transcribe
```

### Step 2: Add to Your docker-compose.yml
```bash
# Append the content from docker-compose.server.yml to your docker-compose.yml
cat v-transcribe/docker-compose.server.yml >> docker-compose.yml
```

### Step 3: Deploy
```bash
# That's it! One command deployment
docker compose up -d
```

---

## ✅ Verification

```bash
# Check services are running
docker compose ps | grep transcribe

# Test web interface
curl http://192.168.10.73
```

## 🌐 Access via Nginx Proxy Manager

**Add Proxy Host:**
- **Domain**: `transcribe.yourdomain.com`
- **Forward to**: `192.168.10.73` port `80`
- **SSL**: Enable Let's Encrypt

**Done!** Visit `https://transcribe.yourdomain.com`

---

## 📂 What Gets Added

**3 Services:**
- `v-transcribe-api` (192.168.10.71)
- `v-transcribe-whisper` (192.168.10.72) 
- `v-transcribe-web` (192.168.10.73)

**2 Volumes:**
- `youtube-whisper-models`
- `youtube-transcripts`

**Memory Usage:**
- ~1.5GB total (whisper needs most RAM)

That's it! Simple as `docker compose up -d` 🎉