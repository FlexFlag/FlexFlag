#!/bin/bash

# FlexFlag Stop Development Services Script

echo "🛑 Stopping FlexFlag development services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Stop API server
if [ -f .api.pid ]; then
    API_PID=$(cat .api.pid)
    if kill -0 $API_PID 2>/dev/null; then
        print_status "Stopping API server (PID: $API_PID)..."
        kill $API_PID
        rm .api.pid
        print_success "API server stopped"
    else
        print_status "API server not running"
        rm -f .api.pid
    fi
else
    print_status "No API PID file found"
fi

# Stop UI server
if [ -f .ui.pid ]; then
    UI_PID=$(cat .ui.pid)
    if kill -0 $UI_PID 2>/dev/null; then
        print_status "Stopping UI server (PID: $UI_PID)..."
        kill $UI_PID
        rm .ui.pid
        print_success "UI server stopped"
    else
        print_status "UI server not running"
        rm -f .ui.pid
    fi
else
    print_status "No UI PID file found"
fi

# Kill any remaining processes on the ports
print_status "Checking for any remaining processes..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true

print_success "All FlexFlag development services stopped!"
echo ""
echo "📋 To restart: ./scripts/restart-dev.sh or ./setup-dev.sh"