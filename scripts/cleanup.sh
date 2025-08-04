#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== YouTube Transcribe Cleanup Script ===${NC}"
echo ""

# Check which docker compose command is available
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    echo -e "${RED}Error: docker-compose is not installed.${NC}"
    exit 1
fi

# Check if services are running
if $DOCKER_COMPOSE -f docker-compose.local.yml ps -q 2>/dev/null | grep -q .; then
    echo -e "${YELLOW}Stopping running services...${NC}"
    $DOCKER_COMPOSE -f docker-compose.local.yml down
    echo -e "${GREEN}✓ Services stopped${NC}"
else
    echo -e "${GREEN}✓ No services running${NC}"
fi

# Ask about volumes
echo ""
read -p "Remove Docker volumes (whisper models and transcripts)? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Removing Docker volumes...${NC}"
    $DOCKER_COMPOSE -f docker-compose.local.yml down -v 2>/dev/null
    echo -e "${GREEN}✓ Volumes removed${NC}"
fi

# Ask about local directories
echo ""
read -p "Remove local data directories (transcripts, whisper-models)? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Removing local directories...${NC}"
    rm -rf transcripts whisper-models
    echo -e "${GREEN}✓ Directories removed${NC}"
fi

# Ask about config/secrets.env file
echo ""
read -p "Remove config/secrets.env file? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f config/secrets.env ]; then
        rm config/secrets.env
        echo -e "${GREEN}✓ config/secrets.env file removed${NC}"
    else
        echo -e "${YELLOW}config/secrets.env file not found${NC}"
    fi
fi

# Remove orphaned containers
echo ""
echo -e "${YELLOW}Checking for orphaned containers...${NC}"
$DOCKER_COMPOSE -f docker-compose.local.yml rm -f 2>/dev/null
echo -e "${GREEN}✓ Cleanup complete${NC}"

# Show remaining Docker images
echo ""
echo -e "${YELLOW}Docker images used by this project:${NC}"
docker images | grep -E "(v-transcribe|whisper|nginx)" | head -10

echo ""
echo -e "${YELLOW}To remove these images, run:${NC}"
echo "docker rmi \$(docker images | grep -E '(v-transcribe|whisper)' | awk '{print \$3}')"

echo ""
echo -e "${GREEN}Cleanup complete!${NC}"