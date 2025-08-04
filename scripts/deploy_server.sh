#!/bin/bash

# YouTube Transcription Server Deployment Script
# This script helps deploy the transcription service to your existing server

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== YouTube Transcription Server Deployment ===${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.server.yml" ]; then
    echo -e "${RED}Error: docker-compose.server.yml not found. Run this script from the v-transcribe directory.${NC}"
    exit 1
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if main docker-compose.yml exists
if [ ! -f "../docker-compose.yml" ] && [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}Error: Main docker-compose.yml not found. Make sure you're in your server's docker directory.${NC}"
    exit 1
fi

echo -e "${YELLOW}Pre-deployment checklist:${NC}"
echo "1. ✓ Docker is running"
echo "2. ✓ docker-compose.server.yml found"

# Create config if it doesn't exist
if [ ! -f config/secrets.env ]; then
    echo -e "${YELLOW}Creating config/secrets.env from template...${NC}"
    cp config/secrets.env.example config/secrets.env
    echo -e "${GREEN}✓ Created config/secrets.env${NC}"
    echo -e "${YELLOW}Please edit config/secrets.env with your settings if needed${NC}"
else
    echo -e "${GREEN}✓ config/secrets.env exists${NC}"
fi

# Check for IP conflicts
echo ""
echo -e "${YELLOW}Checking for IP address conflicts...${NC}"
EXISTING_IPS=$(docker network inspect npm_proxy 2>/dev/null | grep -o "192\.168\.10\.[0-9]\+" | sort -u || echo "")
PROPOSED_IPS=("192.168.10.71" "192.168.10.72" "192.168.10.73")

for ip in "${PROPOSED_IPS[@]}"; do
    if echo "$EXISTING_IPS" | grep -q "$ip"; then
        echo -e "${RED}⚠ IP conflict detected: $ip is already in use${NC}"
        echo -e "${YELLOW}Please update the IP addresses in docker-compose.server.yml${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ IP $ip is available${NC}"
    fi
done

# Ask for confirmation
echo ""
echo -e "${YELLOW}Ready to deploy transcription services. This will:${NC}"
echo "- Add 3 new containers to your docker-compose"
echo "- Create Docker volumes for transcripts and models"
echo "- Use ~1.5-2GB RAM for Whisper processing"
echo "- Expose web interface on npm_proxy network"
echo ""
read -p "Continue with deployment? (y/N) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Deployment cancelled.${NC}"
    exit 1
fi

# Show instructions for manual integration
echo ""
echo -e "${BLUE}=== DEPLOYMENT INSTRUCTIONS ===${NC}"
echo ""
echo -e "${YELLOW}1. Copy the services from docker-compose.server.yml to your main docker-compose.yml:${NC}"
echo "   - Copy the 3 services under your existing 'services:' section"
echo "   - Copy the 2 volumes under your existing 'volumes:' section"
echo ""
echo -e "${YELLOW}2. Build and start the services:${NC}"
echo "   docker compose up -d --build v-transcribe-api v-transcribe-whisper v-transcribe-web"
echo ""
echo -e "${YELLOW}3. Setup Nginx Proxy Manager:${NC}"
echo "   - Domain: transcribe.yourdomain.com"
echo "   - Forward to: v-transcribe-web:80 or 192.168.10.73:80"
echo "   - Enable SSL with Let's Encrypt"
echo ""
echo -e "${YELLOW}4. Test the deployment:${NC}"
echo "   - Visit https://transcribe.yourdomain.com"
echo "   - Submit a test YouTube URL"
echo ""

# Optional: Check if we can automatically detect the main compose file
MAIN_COMPOSE=""
if [ -f "../docker-compose.yml" ]; then
    MAIN_COMPOSE="../docker-compose.yml"
elif [ -f "docker-compose.yml" ]; then
    MAIN_COMPOSE="docker-compose.yml"
fi

if [ -n "$MAIN_COMPOSE" ]; then
    echo ""
    echo -e "${BLUE}=== OPTIONAL: AUTOMATIC INTEGRATION ===${NC}"
    echo ""
    read -p "Automatically add services to $MAIN_COMPOSE? (y/N) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Create backup
        cp "$MAIN_COMPOSE" "${MAIN_COMPOSE}.backup.$(date +%Y%m%d_%H%M%S)"
        echo -e "${GREEN}✓ Created backup: ${MAIN_COMPOSE}.backup.$(date +%Y%m%d_%H%M%S)${NC}"
        
        # Add services (this is simplified - in reality you'd need more sophisticated merging)
        echo -e "${YELLOW}⚠ Manual integration recommended for production safety${NC}"
        echo -e "${YELLOW}Please manually copy the services from docker-compose.server.yml${NC}"
    fi
fi

echo ""
echo -e "${GREEN}=== DEPLOYMENT READY ===${NC}"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "1. Manually integrate the services from docker-compose.server.yml"
echo "2. Run: docker compose up -d --build v-transcribe-api v-transcribe-whisper v-transcribe-web"
echo "3. Setup your Nginx Proxy Manager proxy host"
echo "4. Test with a YouTube URL"
echo ""
echo -e "${BLUE}For detailed instructions, see SERVER_DEPLOYMENT.md${NC}"
echo ""