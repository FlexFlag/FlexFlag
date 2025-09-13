#!/bin/bash

set -e

echo "🚀 FlexFlag Demo Deployment Script"
echo "================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
    echo "✅ Docker installed. Please log out and back in, then run this script again."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose not found. Installing..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo "✅ Docker Compose installed"
fi

# Generate secure secrets
echo "🔐 Generating secure secrets..."
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)

# Create environment file
cat > .env.demo << EOF
# FlexFlag Demo Environment Configuration
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USERNAME=flexflag
DATABASE_PASSWORD=$DB_PASSWORD
DATABASE_NAME=flexflag_demo
REDIS_HOST=redis
REDIS_PORT=6379
JWT_SECRET=$JWT_SECRET

# Demo-specific settings
FLEXFLAG_DEMO_MODE=true
FLEXFLAG_DEMO_RESET_INTERVAL=1h
FLEXFLAG_DEMO_MAX_FLAGS=50
FLEXFLAG_DEMO_MAX_PROJECTS=5
FLEXFLAG_DEMO_TITLE=FlexFlag Demo
FLEXFLAG_DEMO_SUBTITLE=Experience high-performance feature flags

# Performance settings
FLEXFLAG_SERVER_PORT=8080
FLEXFLAG_LOG_LEVEL=info
GIN_MODE=release
EOF

echo "📝 Environment configuration created (.env.demo)"

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.demo.yml down -v 2>/dev/null || true

# Build and start the demo
echo "🏗️ Building and starting FlexFlag demo..."
docker-compose -f docker-compose.demo.yml --env-file .env.demo up -d --build

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 30

# Check if services are running
if docker-compose -f docker-compose.demo.yml ps | grep -q "Up"; then
    echo "✅ FlexFlag demo deployed successfully!"
    echo ""
    echo "🌐 Access your demo at:"
    echo "   http://localhost:8080"
    echo ""
    echo "🔐 Demo credentials:"
    echo "   Email:    demo@flexflag.io"
    echo "   Password: demo123"
    echo ""
    echo "📊 Admin access:"
    echo "   Email:    admin@flexflag.io" 
    echo "   Password: admin123"
    echo ""
    echo "📚 API Documentation:"
    echo "   http://localhost:8080/swagger/index.html"
    echo ""
    echo "🔧 Management commands:"
    echo "   Stop:     docker-compose -f docker-compose.demo.yml down"
    echo "   Logs:     docker-compose -f docker-compose.demo.yml logs -f"
    echo "   Restart:  docker-compose -f docker-compose.demo.yml restart"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Visit http://localhost:8080 and login"
    echo "   2. Explore the demo project and feature flags"
    echo "   3. Test the JavaScript SDK with the demo API"
    echo "   4. Configure your domain and SSL for public access"
else
    echo "❌ Deployment failed. Check logs:"
    docker-compose -f docker-compose.demo.yml logs
    exit 1
fi