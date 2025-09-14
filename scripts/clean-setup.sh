#!/bin/bash

# FlexFlag Clean Setup Script
# This script completely cleans up all containers and data for a fresh start

set -e

echo "🧹 FlexFlag Clean Setup"
echo "========================"

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

# Ask for confirmation
echo "⚠️  This will completely remove all FlexFlag containers, volumes, and cached data"
echo "⚠️  You will lose all existing flags, users, and data in your local development environment"
echo ""
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

print_status "Stopping all running containers..."
docker stop $(docker ps -q --filter "name=flexflag") 2>/dev/null || true

print_status "Removing FlexFlag containers..."
docker rm $(docker ps -aq --filter "name=flexflag") 2>/dev/null || true

print_status "Removing FlexFlag volumes..."
docker volume rm $(docker volume ls -q --filter "name=flexflag") 2>/dev/null || true

print_status "Removing FlexFlag networks..."
docker network rm $(docker network ls -q --filter "name=flexflag") 2>/dev/null || true

print_status "Cleaning Docker system..."
docker system prune -f

print_status "Removing cached dependencies..."
rm -rf ./ui/node_modules ./ui/.next 2>/dev/null || true
go clean -cache -modcache 2>/dev/null || true

print_status "Cleaning build artifacts..."
rm -rf ./bin 2>/dev/null || true

print_success "Clean setup complete!"
echo ""
echo "🚀 Ready for fresh setup!"
echo "Run: ./setup-dev.sh"