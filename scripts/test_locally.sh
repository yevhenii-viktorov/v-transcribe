#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== YouTube Transcribe Local Testing Script ===${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check which docker compose command is available
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    echo -e "${RED}Error: docker-compose is not installed.${NC}"
    echo -e "${YELLOW}Install Docker Desktop or docker-compose package.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Using: $DOCKER_COMPOSE${NC}"

# Create config/secrets.env file if it doesn't exist
if [ ! -f config/secrets.env ]; then
    echo -e "${YELLOW}Creating config/secrets.env file from template...${NC}"
    cp config/secrets.env.example config/secrets.env
    echo -e "${GREEN}✓ Created config/secrets.env file${NC}"
    echo -e "${YELLOW}Please edit config/secrets.env with your actual credentials${NC}"
else
    echo -e "${GREEN}✓ config/secrets.env file exists${NC}"
fi

# Create necessary directories
echo -e "${YELLOW}Creating necessary directories...${NC}"
mkdir -p transcripts whisper-models
echo -e "${GREEN}✓ Directories created${NC}"

# Check if port 8080 is available
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${RED}Warning: Port 8080 is already in use!${NC}"
    echo -e "${YELLOW}You may need to change the port in docker-compose.local.yml${NC}"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build and start services
echo ""
echo -e "${YELLOW}Building and starting services...${NC}"
echo -e "${YELLOW}This may take a few minutes on first run...${NC}"
echo ""

$DOCKER_COMPOSE -f docker-compose.local.yml up --build -d

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Services started successfully!${NC}"
    echo ""
    echo -e "${GREEN}Access the application at: ${NC}http://localhost:8080"
    echo ""
    echo -e "${YELLOW}Useful commands:${NC}"
    echo "  View logs:        $DOCKER_COMPOSE -f docker-compose.local.yml logs -f"
    echo "  View API logs:    $DOCKER_COMPOSE -f docker-compose.local.yml logs -f api"
    echo "  Stop services:    $DOCKER_COMPOSE -f docker-compose.local.yml down"
    echo "  Clean up:         ./scripts/cleanup.sh"
else
    echo ""
    echo -e "${RED}Failed to start services. Check the logs:${NC}"
    echo "$DOCKER_COMPOSE -f docker-compose.local.yml logs"
    exit 1
fi