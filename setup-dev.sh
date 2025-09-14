#!/bin/bash

# FlexFlag Development Setup Script
# This script helps developers quickly set up FlexFlag for local development

set -e

echo "🚀 FlexFlag Development Setup"
echo "==============================="

# Add cleanup function for failed setups
cleanup() {
    print_error "Setup failed! Cleaning up..."
    docker-compose down 2>/dev/null || true
    exit 1
}
trap cleanup ERR

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if .env exists
if [ ! -f ".env" ]; then
    print_status "Creating .env file from .env.example..."
    cp .env.example .env
    print_success ".env file created!"
else
    print_warning ".env file already exists, skipping..."
fi

# Check for required tools
print_status "Checking for required tools..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18.17+ first."
    exit 1
else
    # Check Node.js version - Next.js requires >= v18.17.0
    NODE_VERSION=$(node -v | sed 's/v//')
    NODE_MAJOR=$(echo $NODE_VERSION | cut -d. -f1)
    NODE_MINOR=$(echo $NODE_VERSION | cut -d. -f2)
    
    if [ "$NODE_MAJOR" -lt 18 ] || ([ "$NODE_MAJOR" -eq 18 ] && [ "$NODE_MINOR" -lt 17 ]); then
        print_error "Node.js version $NODE_VERSION is installed, but Next.js requires >= v18.17.0"
        print_error "Please upgrade Node.js: https://nodejs.org/"
        exit 1
    fi
fi

print_success "All required tools are installed!"

# Generate JWT secret if needed
print_status "Checking JWT secret..."
if grep -q "your-super-secret-jwt-key-change-in-production" .env; then
    print_status "Generating secure JWT secret..."
    if command -v openssl &> /dev/null; then
        JWT_SECRET=$(openssl rand -base64 32)
        # Escape special characters for sed
        JWT_SECRET_ESCAPED=$(printf '%s\n' "$JWT_SECRET" | sed 's/[[\.*^$()+?{|]/\\&/g')
        sed -i.bak "s/your-super-secret-jwt-key-change-in-production/$JWT_SECRET_ESCAPED/" .env
        rm .env.bak
        print_success "JWT secret generated and updated in .env"
    else
        print_warning "OpenSSL not found, please manually update JWT secret in .env"
    fi
fi

# Start infrastructure services
print_status "Starting PostgreSQL and Redis with Docker Compose..."
if ! docker-compose -f docker-compose.dev.yml up -d; then
    print_error "Failed to start Docker services"
    print_error "Make sure Docker is running and try: docker-compose -f docker-compose.dev.yml down && docker-compose -f docker-compose.dev.yml up -d"
    exit 1
fi

# Wait for PostgreSQL to be ready with better health check
print_status "Waiting for PostgreSQL to be ready..."
max_attempts=30
attempts=0
while ! docker exec flexflag-postgres pg_isready -U flexflag -d flexflag >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ $attempts -ge $max_attempts ]; then
        print_error "PostgreSQL failed to start after 30 attempts"
        print_error "Check logs with: docker-compose logs postgres"
        exit 1
    fi
    print_status "Waiting for PostgreSQL... (attempt $attempts/$max_attempts)"
    sleep 2
done
print_success "PostgreSQL is ready!"

# Run database migrations with better error handling
print_status "Running database migrations..."
if command -v make &> /dev/null; then
    # Check for dirty migration state and fix it
    if ! make migrate-up 2>/dev/null; then
        print_warning "Migration failed, attempting to fix dirty state..."
        
        # Try to force reset and re-run
        print_status "Checking migration state..."
        DIRTY_VERSION=$(docker exec flexflag-postgres psql -U flexflag -d flexflag -tAc "SELECT version FROM schema_migrations WHERE dirty = true LIMIT 1;" 2>/dev/null || echo "")
        
        if [ ! -z "$DIRTY_VERSION" ]; then
            print_status "Found dirty migration at version $DIRTY_VERSION, forcing reset..."
            go run cmd/migrator/main.go -database-url "postgres://flexflag:flexflag@localhost:5433/flexflag?sslmode=disable" -force-version $((DIRTY_VERSION - 1)) || true
        fi
        
        # Try migration again
        if ! make migrate-up; then
            print_error "Migrations failed even after cleanup. Please check database manually."
            print_error "Try running: make migrate-down && make migrate-up"
            exit 1
        fi
    fi
else
    print_warning "Make not found, please run 'make migrate-up' manually after setup"
fi

# Build the application
print_status "Building FlexFlag server..."
if command -v make &> /dev/null; then
    make build
else
    print_warning "Make not found, please run 'make build' manually"
fi

# Install UI dependencies with error handling
print_status "Installing UI dependencies..."
if [ -d "ui" ]; then
    cd ui
    if ! npm install; then
        print_error "Failed to install UI dependencies"
        print_error "Try running: cd ui && rm -rf node_modules package-lock.json && npm install"
        exit 1
    fi
    cd ..
else
    print_warning "UI directory not found, skipping npm install"
fi

# Create logs directory
mkdir -p logs

# Start both API server and UI automatically
print_status "Starting FlexFlag services..."

# Start API server in background
print_status "Starting API server on port 8080..."
make run > logs/api.log 2>&1 &
API_PID=$!
echo $API_PID > .api.pid

# Wait a moment for server to start
sleep 3

# Test health endpoint
if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
    print_success "API server is running!"
else
    print_warning "API server may still be starting..."
fi

# Start UI in background
print_status "Starting UI on port 3000..."
cd ui
npm run dev > ../logs/ui.log 2>&1 &
UI_PID=$!
echo $UI_PID > ../.ui.pid
cd ..

# Wait for UI to start
sleep 5

# Check if UI is running
if curl -f -s http://localhost:3000 > /dev/null 2>&1; then
    print_success "UI is running!"
else
    print_status "UI is starting up... (may take a few moments)"
fi

print_success "🎉 FlexFlag is running!"
echo ""
echo "📱 Access FlexFlag:"
echo "• UI:         http://localhost:3000"
echo "• API:        http://localhost:8080"  
echo "• Swagger:    http://localhost:8080/swagger"
echo "• Health:     http://localhost:8080/health"
echo ""
echo "🔐 Default login:"
echo "• Email:      admin@example.com"
echo "• Password:   secret"
echo ""
echo "📋 Useful commands:"
echo "• Stop services:           ./scripts/stop-dev.sh"
echo "• Restart services:        ./scripts/restart-dev.sh"
echo "• View API logs:           tail -f logs/api.log"
echo "• View UI logs:            tail -f logs/ui.log"
echo "• Run tests:               make test"
echo ""
echo "🆘 Having issues? Check: TROUBLESHOOTING.md"