#!/bin/bash

# FlexFlag Restart Development Services Script

echo "🔄 Restarting FlexFlag development services..."

# Stop existing services
./scripts/stop-dev.sh

# Wait a moment
sleep 2

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Create logs directory
mkdir -p logs

# Start API server
print_status "Starting API server..."
make run > logs/api.log 2>&1 &
API_PID=$!
echo $API_PID > .api.pid

# Start UI server
print_status "Starting UI server..."
cd ui
npm run dev > ../logs/ui.log 2>&1 &
UI_PID=$!
echo $UI_PID > ../.ui.pid
cd ..

print_success "FlexFlag services restarted!"
echo ""
echo "📱 Access FlexFlag:"
echo "• UI:         http://localhost:3000"
echo "• API:        http://localhost:8080"
echo ""
echo "📋 Useful commands:"
echo "• Stop services:    ./scripts/stop-dev.sh"
echo "• View API logs:    tail -f logs/api.log"
echo "• View UI logs:     tail -f logs/ui.log"