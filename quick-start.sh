#!/bin/bash

# FlexFlag Quick Start Script
# This script sets up FlexFlag with minimal configuration

set -e

echo "🚀 FlexFlag Quick Start"
echo "======================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose from https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✅ Docker is installed${NC}"
echo -e "${GREEN}✅ Docker Compose is installed${NC}"
echo ""

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${YELLOW}📝 Creating .env file from template...${NC}"
    cp .env.example .env

    # Generate secure JWT secret
    JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "change-this-secure-secret-in-production")

    # Update .env with generated secret (macOS and Linux compatible)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|FLEXFLAG_JWT_SECRET=.*|FLEXFLAG_JWT_SECRET=\"$JWT_SECRET\"|" .env
    else
        sed -i "s|FLEXFLAG_JWT_SECRET=.*|FLEXFLAG_JWT_SECRET=\"$JWT_SECRET\"|" .env
    fi

    echo -e "${GREEN}✅ .env file created${NC}"
else
    echo -e "${GREEN}✅ .env file already exists${NC}"
fi

echo ""
echo -e "${YELLOW}🏗️  Building and starting FlexFlag...${NC}"
echo "This may take a few minutes on first run..."
echo ""

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo ""
echo -e "${YELLOW}⏳ Waiting for services to start...${NC}"
sleep 10

# Check health
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API is healthy${NC}"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT+1))
    echo -n "."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ API failed to start${NC}"
    echo "View logs: docker-compose -f docker-compose.prod.yml logs"
    exit 1
fi

# Check UI
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:3000 > /dev/null 2>&1; then
        echo -e "${GREEN}✅ UI is healthy${NC}"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT+1))
    echo -n "."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ UI failed to start${NC}"
    echo "View logs: docker-compose -f docker-compose.prod.yml logs ui"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 FlexFlag is running!${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Access FlexFlag:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "  🌐 Frontend:      ${GREEN}http://localhost:3000${NC}"
echo -e "  🚀 API:           ${GREEN}http://localhost:8080${NC}"
echo -e "  📖 API Docs:      ${GREEN}http://localhost:8080/swagger/index.html${NC}"
echo -e "  ❤️  Health Check:  ${GREEN}http://localhost:8080/health${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}Default Credentials:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Email:    admin@flexflag.com"
echo "  Password: admin123"
echo ""
echo -e "${RED}⚠️  IMPORTANT: Change the password after first login!${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}Useful Commands:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  View logs:    docker-compose -f docker-compose.prod.yml logs -f"
echo "  Stop:         docker-compose -f docker-compose.prod.yml down"
echo "  Restart:      docker-compose -f docker-compose.prod.yml restart"
echo "  Status:       docker-compose -f docker-compose.prod.yml ps"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${GREEN}Happy flagging! 🚩${NC}"
echo ""
