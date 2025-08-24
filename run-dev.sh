#!/bin/bash

# FlexFlag Single-Command Development Runner
# This script sets up and runs ALL FlexFlag components with one command

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                          🚀 FlexFlag Development                             ║"
    echo "║                        Single Command Setup & Run                           ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_section() {
    echo -e "\n${CYAN}▶ $1${NC}"
    echo "────────────────────────────────────────"
}

# Cleanup function for graceful shutdown
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down FlexFlag...${NC}"
    
    # Kill background processes
    if [ ! -z "$API_PID" ]; then
        print_status "Stopping API server (PID: $API_PID)"
        kill $API_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$UI_PID" ]; then
        print_status "Stopping UI server (PID: $UI_PID)"
        kill $UI_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$EDGE_PID" ]; then
        print_status "Stopping Edge server (PID: $EDGE_PID)"
        kill $EDGE_PID 2>/dev/null || true
    fi
    
    # Stop Docker services
    print_status "Stopping Docker services"
    docker-compose down > /dev/null 2>&1 || true
    
    echo -e "${GREEN}✅ FlexFlag stopped cleanly${NC}"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

print_banner

# Check for required tools
print_section "🔍 Checking Prerequisites"

MISSING_TOOLS=""

if ! command -v docker &> /dev/null; then
    MISSING_TOOLS+="docker "
fi

if ! command -v docker-compose &> /dev/null; then
    MISSING_TOOLS+="docker-compose "
fi

if ! command -v go &> /dev/null; then
    MISSING_TOOLS+="go "
fi

if ! command -v node &> /dev/null; then
    MISSING_TOOLS+="node "
fi

if ! command -v make &> /dev/null; then
    MISSING_TOOLS+="make "
fi

if [ ! -z "$MISSING_TOOLS" ]; then
    print_error "Missing required tools: $MISSING_TOOLS"
    echo "Please install the missing tools and try again."
    echo "See README.md for installation instructions."
    exit 1
fi

print_success "All required tools are installed"

# Setup environment
print_section "⚙️ Setting Up Environment"

if [ ! -f ".env" ]; then
    print_status "Creating .env from template..."
    cp .env.example .env
    
    # Generate secure JWT secret
    if command -v openssl &> /dev/null; then
        print_status "Generating secure JWT secret..."
        JWT_SECRET=$(openssl rand -base64 32)
        JWT_SECRET_ESCAPED=$(printf '%s\n' "$JWT_SECRET" | sed 's/[[\.*^$()+?{|]/\\&/g')
        sed -i.bak "s/your-super-secret-jwt-key-change-in-production/$JWT_SECRET_ESCAPED/" .env
        rm .env.bak
    fi
    
    print_success "Environment configured"
else
    print_warning ".env already exists"
fi

# Start infrastructure
print_section "🐳 Starting Infrastructure Services"

print_status "Starting PostgreSQL and Redis..."
docker-compose up -d postgres redis

# Wait for services to be ready
print_status "Waiting for services to initialize..."
sleep 8

# Run migrations and build
print_section "🔨 Building Application"

print_status "Running database migrations..."
make migrate-up > /dev/null 2>&1 || {
    print_error "Database migrations failed"
    cleanup
}

print_status "Building FlexFlag server..."
make build > /dev/null 2>&1 || {
    print_error "Build failed"
    cleanup
}

print_status "Installing UI dependencies..."
cd ui
npm install > /dev/null 2>&1 || {
    print_error "UI dependency installation failed"
    cleanup
}
cd ..

# Start all services
print_section "🚀 Starting FlexFlag Services"

# Start API server in background
print_status "Starting API server on port 8080..."
make run > /tmp/flexflag-api.log 2>&1 &
API_PID=$!
sleep 3

# Check if API server started successfully
if ! kill -0 $API_PID 2>/dev/null; then
    print_error "API server failed to start"
    print_error "Check logs: tail /tmp/flexflag-api.log"
    cleanup
fi

# Start UI server in background  
print_status "Starting UI server on port 3000..."
cd ui
npm run dev > /tmp/flexflag-ui.log 2>&1 &
UI_PID=$!
cd ..
sleep 5

# Check if UI server started successfully
if ! kill -0 $UI_PID 2>/dev/null; then
    print_error "UI server failed to start"
    print_error "Check logs: tail /tmp/flexflag-ui.log"
    cleanup
fi

# Start Edge server in background
print_status "Starting Edge server on port 8083 with SSE..."
FLEXFLAG_EDGE_SYNC_TYPE=sse FLEXFLAG_EDGE_PORT=8083 FLEXFLAG_HUB_URL=http://localhost:8080 ./bin/edge-server > /tmp/flexflag-edge.log 2>&1 &
EDGE_PID=$!
sleep 3

# Check if Edge server started successfully
if ! kill -0 $EDGE_PID 2>/dev/null; then
    print_warning "Edge server failed to start (this is optional)"
    EDGE_PID=""
fi

# Display success information
print_section "✅ FlexFlag is Running!"

echo -e "${GREEN}🎉 All services are running successfully!${NC}\n"

echo -e "${CYAN}📍 Service URLs:${NC}"
echo "┌─────────────────┬─────────────────────────────────────┐"
echo "│ Service         │ URL                                 │"
echo "├─────────────────┼─────────────────────────────────────┤"
echo "│ 🎨 UI Dashboard │ http://localhost:3000               │"
echo "│ 🔧 API Server   │ http://localhost:8080               │"
echo "│ ⚡ Edge Server  │ http://localhost:8083               │"
echo "│ 📊 API Health   │ http://localhost:8080/health        │"
echo "│ 🏠 Edge Servers │ http://localhost:3000/edge-servers  │"
echo "└─────────────────┴─────────────────────────────────────┘"

echo -e "\n${CYAN}🔐 Default Login Credentials:${NC}"
echo "┌─────────────┬─────────────────────┐"
echo "│ Email       │ admin@example.com   │"
echo "│ Password    │ secret              │"
echo "└─────────────┴─────────────────────┘"

echo -e "\n${CYAN}📁 Log Files:${NC}"
echo "• API Server: /tmp/flexflag-api.log"
echo "• UI Server:  /tmp/flexflag-ui.log"
echo "• Edge Server: /tmp/flexflag-edge.log"

echo -e "\n${CYAN}⌨️ Useful Commands:${NC}"
echo "• View API logs:   tail -f /tmp/flexflag-api.log"
echo "• View UI logs:    tail -f /tmp/flexflag-ui.log"
echo "• View Edge logs:  tail -f /tmp/flexflag-edge.log"
echo "• Stop all:        Ctrl+C (or kill this process)"

echo -e "\n${PURPLE}💡 Quick Start Tips:${NC}"
echo "1. Visit http://localhost:3000 and login with the credentials above"
echo "2. Create a new project and some feature flags"
echo "3. Check the Edge Servers page to see your running edge server"
echo "4. Test flag evaluation at http://localhost:8080/health"

echo -e "\n${YELLOW}📖 Documentation:${NC}"
echo "• Check CLAUDE.md for development commands"
echo "• See docs/ENVIRONMENT_VARIABLES.md for configuration"
echo "• Visit /swagger endpoint for API documentation"

echo -e "\n${GREEN}🎯 FlexFlag is ready for development!${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}\n"

# Keep script running and show live status
while true; do
    sleep 30
    
    # Check if processes are still running
    if [ ! -z "$API_PID" ] && ! kill -0 $API_PID 2>/dev/null; then
        print_error "API server stopped unexpectedly"
        cleanup
    fi
    
    if [ ! -z "$UI_PID" ] && ! kill -0 $UI_PID 2>/dev/null; then
        print_error "UI server stopped unexpectedly"
        cleanup
    fi
    
    if [ ! -z "$EDGE_PID" ] && ! kill -0 $EDGE_PID 2>/dev/null; then
        print_warning "Edge server stopped"
        EDGE_PID=""
    fi
    
    # Optional: Show brief status
    echo -e "${BLUE}[$(date '+%H:%M:%S')] Services running: API✅ UI✅$([ ! -z "$EDGE_PID" ] && echo ' Edge✅' || echo ' Edge❌')${NC}"
done