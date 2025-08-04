#!/bin/bash

# One-Command Server Deployment for YouTube Transcription Service
# Usage: ./deploy.sh

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 YouTube Transcription Service - One Command Deploy${NC}"
echo ""

# Check if docker-compose.yml exists in parent or current directory
COMPOSE_FILE=""
if [ -f "../docker-compose.yml" ]; then
    COMPOSE_FILE="../docker-compose.yml"
    WORK_DIR=".."
elif [ -f "docker-compose.yml" ]; then
    COMPOSE_FILE="docker-compose.yml"
    WORK_DIR="."
else
    echo -e "${RED}❌ docker-compose.yml not found in current or parent directory${NC}"
    echo "Please run this from your server's docker directory"
    exit 1
fi

echo -e "${GREEN}✅ Found docker-compose.yml at: $COMPOSE_FILE${NC}"

# Setup config
if [ ! -f "config/secrets.env" ]; then
    echo -e "${YELLOW}📝 Creating config/secrets.env...${NC}"
    cp config/secrets.env.example config/secrets.env
fi

# Check if services already exist
if grep -q "v-transcribe-api" "$COMPOSE_FILE"; then
    echo -e "${YELLOW}⚠️  Transcription services already exist in docker-compose.yml${NC}"
    echo "Skipping service addition..."
else
    echo -e "${YELLOW}📋 Adding transcription services to docker-compose.yml...${NC}"
    
    # Add the services
    echo "" >> "$COMPOSE_FILE"
    echo "  # YouTube Transcription Services" >> "$COMPOSE_FILE"
    cat docker-compose.server.yml | grep -A 1000 "services:" | grep -v "^services:" >> "$COMPOSE_FILE"
fi

# Deploy
echo -e "${GREEN}🚀 Deploying with: docker compose up -d${NC}"
cd "$WORK_DIR"
docker compose up -d

echo ""
echo -e "${GREEN}✅ Deployment Complete!${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. 🌐 Setup Nginx Proxy Manager:"
echo "   - Domain: transcribe.yourdomain.com"
echo "   - Forward to: 192.168.10.73:80"
echo "   - Enable SSL"
echo ""
echo "2. 🧪 Test: curl http://192.168.10.73"
echo "3. 🎉 Visit: https://transcribe.yourdomain.com"
echo ""
echo -e "${GREEN}🎬 Ready to transcribe YouTube videos!${NC}"