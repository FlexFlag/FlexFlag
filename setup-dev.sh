#!/bin/bash

# FlexFlag Development Setup Script
# This script helps developers quickly set up FlexFlag for local development

set -e

echo "🚀 FlexFlag Development Setup"
echo "==============================="

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
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready
print_status "Waiting for PostgreSQL to be ready..."
sleep 5

# Run database migrations
print_status "Running database migrations..."
if command -v make &> /dev/null; then
    make migrate-up
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

# Install UI dependencies
print_status "Installing UI dependencies..."
cd ui
npm install
cd ..

print_success "Development setup complete!"
echo ""
echo "🎉 FlexFlag is ready for development!"
echo ""
echo "Next steps:"
echo "1. Start the API server:    make run"
echo "2. Start the UI:            cd ui && npm run dev"  
echo "3. Start an edge server:    FLEXFLAG_EDGE_SYNC_TYPE=sse FLEXFLAG_EDGE_PORT=8083 ./bin/edge-server"
echo ""
echo "URLs:"
echo "• API Server: http://localhost:8080"
echo "• UI:         http://localhost:3000"
echo "• Edge:       http://localhost:8083"
echo ""
echo "Default login:"
echo "• Email:    admin@example.com"
echo "• Password: secret"
echo ""
echo "📚 Check CLAUDE.md for development commands and architecture overview"