# FlexFlag - High-Performance Feature Flag Management System

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?style=flat&logo=next.js)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/Live%20Demo-Vercel-black?style=flat&logo=vercel)](https://flex-flag-zyen-idpjh03va-reez-personals-projects.vercel.app/)

## 🌐 Live Demo

**[https://flex-flag-zyen-idpjh03va-reez-personals-projects.vercel.app/](https://flex-flag-zyen-idpjh03va-reez-personals-projects.vercel.app/)**

> Try it out with the demo credentials below — no setup required.

| Field    | Value                 |
|----------|-----------------------|
| Email    | `admin@example.com`   |
| Password | `admin123`            |

FlexFlag is a high-performance, developer-first feature flag management system with distributed edge servers, real-time synchronization, and sub-millisecond flag evaluation.

## ✨ Key Features

- 🚀 **Ultra-Fast Evaluation**: <1ms flag evaluation with edge servers
- 🌐 **Distributed Architecture**: Edge servers for global low-latency access
- ⚡ **Real-time Sync**: SSE/WS-based flag propagation to edge nodes
- 🎯 **Advanced Targeting**: User segments, rollouts, and A/B testing
- 🏢 **Multi-Project Support**: Project isolation with environment management
- 🎨 **Modern UI**: Next.js dashboard with dark mode and responsive design
- 🔐 **Enterprise Security**: JWT authentication, API key management, role-based access
- 📊 **Analytics Ready**: Built-in metrics and performance monitoring

## 🚀 Quick Start

### ⚡ Single Command Setup (Recommended)

**Get FlexFlag running in under 2 minutes:**

```bash
# Clone the repository
git clone https://github.com/FlexFlag/FlexFlag.git
cd FlexFlag

# Run the development setup (installs everything automatically)
./setup-dev.sh
```

🎉 **The setup script automatically:**
- ✅ Validates prerequisites (Docker, Go 1.21+, Node.js 18.17+)
- ✅ Starts PostgreSQL and Redis via Docker
- ✅ Runs database migrations with error recovery
- ✅ Builds the Go backend and installs UI dependencies
- ✅ **Starts both API server and UI automatically**
- ✅ Verifies setup with health checks

**That's it!** After setup completes, FlexFlag is ready to use!

**URLs:** API (http://localhost:8080) | UI (http://localhost:3000) | Swagger (http://localhost:8080/swagger)

**Default login:** `admin@example.com` / `admin123`

### 🎛️ Development Commands

```bash
# Stop all services
./scripts/stop-dev.sh

# Restart all services  
./scripts/restart-dev.sh

# View logs
tail -f logs/api.log    # API server logs
tail -f logs/ui.log     # UI server logs

# Run tests
make test
```

---

### 🧹 Clean Start (If Issues Occur)

If you encounter any setup issues, use the clean setup:

```bash
# Completely clean and restart (removes all data)
./scripts/clean-setup.sh
./setup-dev.sh
```

---

### 🛠️ Manual Setup (For Advanced Users)

<details>
<summary>Click to expand manual setup instructions</summary>

#### Prerequisites

Ensure you have the following installed:

- **Go 1.21+** - [Download Go](https://go.dev/dl/)
- **Node.js 18.17+** - [Download Node.js](https://nodejs.org/) *(Next.js requirement)*
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)

#### Step 1: Clone and Setup

```bash
# Clone the repository
git clone https://github.com/flexflag/flexflag.git
cd flexflag

# Automated setup (recommended)
./setup-dev.sh
```

#### Step 2: Start Services Manually

```bash
# Start infrastructure
docker-compose up -d

# Start API server
make run

# Start UI (in another terminal)
cd ui && npm run dev

FLEXFLAG_EDGE_SYNC_TYPE=sse ./bin/edge-server
```

**Service URLs:**
- 🎨 UI Dashboard: http://localhost:3000
- 🔧 API Server: http://localhost:8080  
- ⚡ Edge Server: http://localhost:8083

</details>

---

## 🏗️ Development Commands

```bash
# In a new terminal, navigate to the UI directory
cd ui

# Install Node.js dependencies
npm install

# Start the Next.js development server (runs on port 3000)
npm run dev
```

### Step 6: Access the Application

1. **Main Application**: http://localhost:3000
2. **API Documentation**: http://localhost:8080/swagger/index.html
3. **Health Check**: http://localhost:8080/health

### Step 7: Create Your First User

```bash
# Register a new admin user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "full_name": "Admin User",
    "role": "admin"
  }'
```

Then login at http://localhost:3000 with:
- Email: `admin@example.com`
- Password: `admin123`

## 🏗️ Project Structure

```
flexflag/
├── cmd/                      # Application entrypoints
│   ├── server/              # Main API server
│   ├── edge-server/         # Distributed edge server
│   ├── cli/                 # CLI tool (planned)
│   └── migrator/            # Database migration tool
├── internal/                # Private application code
│   ├── api/                 # HTTP handlers and middleware
│   │   └── handlers/        # Request handlers
│   ├── auth/                # Authentication & authorization
│   ├── config/              # Configuration management
│   ├── edge/                # Edge server components
│   ├── services/            # Business logic services
│   └── storage/             # Data access layer
│       └── postgres/        # PostgreSQL repositories
├── pkg/                     # Public library code
│   ├── types/               # Shared data types
│   └── utils/               # Utility functions
├── ui/                      # Next.js frontend application
│   ├── app/                 # App router pages
│   ├── components/          # Reusable UI components
│   ├── contexts/            # React contexts
│   ├── lib/                 # Client-side utilities
│   └── types/               # TypeScript type definitions
├── migrations/              # Database migration files
├── deployments/             # Deployment configurations
├── docs/                    # Architecture documentation
├── api/                     # Generated Swagger documentation
└── benchmark_scripts/       # Performance testing tools
```

## 🌐 Edge Server Setup (Optional)

FlexFlag supports distributed edge servers for ultra-low latency flag evaluation worldwide.

### Step 1: Generate API Key

1. Login to the FlexFlag UI at http://localhost:3000
2. Navigate to a project → **API Keys**
3. Create a new API key with `evaluation` permissions
4. Copy the generated API key (starts with `ff_production_...`)

### Step 2: Build and Start Edge Server

```bash
# Build the edge server binary
make build-edge

# Start edge server (replace with your actual API key)
FLEXFLAG_EDGE_PORT=8083 \
FLEXFLAG_EDGE_HUB_URL=http://localhost:8080 \
FLEXFLAG_EDGE_API_KEY=your_api_key_here \
./bin/edge-server
```

### Step 3: Verify Edge Server Connection

```bash
# Check edge server health
curl http://localhost:8083/health

# Test flag evaluation via edge server
curl -X POST http://localhost:8083/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key_here" \
  -d '{
    "flag_key": "test-flag",
    "user_context": {
      "user_id": "test-user",
      "attributes": {}
    }
  }'
```

### Step 4: Monitor Edge Servers

- Navigate to **Edge Servers** in the UI to monitor connected edge servers
- View real-time connection status, uptime, and regional distribution
- Auto-refresh every 30 seconds shows live server health

### Multi-Region Edge Deployment

```bash
# Deploy edge servers to multiple regions using Docker Compose
docker-compose -f docker-compose.edge.yml up -d

# This starts edge servers in:
# - US East (port 8081)
# - US West (port 8082) 
# - Europe (port 8083)
# - Asia (port 8084)
```

## 🛠️ Development Commands

### Backend Development

```bash
# Start with hot reload using Air
make dev

# Run tests
make test                    # Unit tests
make test-integration        # Integration tests (requires running services)
make bench                   # Performance benchmarks

# Database operations
make migrate-up              # Apply migrations
make migrate-down           # Rollback migrations
make migrate-create NAME=add_feature  # Create new migration

# Code quality
make lint                    # Run golangci-lint
make fmt                     # Format Go code
make vet                     # Run go vet

# Build
make build                   # Build all binaries
make build-server           # Build server only
make build-edge             # Build edge server only
```

### Frontend Development

```bash
cd ui

# Development server
npm run dev                  # Start Next.js dev server (port 3000)

# Build and deploy
npm run build               # Production build
npm run start               # Start production server
npm run lint                # Run ESLint
npm run type-check          # TypeScript type checking
```

### Performance Testing

```bash
# Test main server performance
./benchmark_evaluation.sh

# Test edge server performance
./benchmark-edge.sh

# Concurrent load testing
./concurrent_benchmark.sh
```

## 🔧 Configuration

### Environment Variables

FlexFlag uses environment variables with the `FLEXFLAG_` prefix:

#### Server Configuration
```bash
# Server settings
FLEXFLAG_SERVER_HOST=0.0.0.0
FLEXFLAG_SERVER_PORT=8080

# Database settings (uses port 5433 to avoid conflicts)
FLEXFLAG_DATABASE_HOST=localhost
FLEXFLAG_DATABASE_PORT=5433
FLEXFLAG_DATABASE_USERNAME=flexflag
FLEXFLAG_DATABASE_PASSWORD=flexflag
FLEXFLAG_DATABASE_NAME=flexflag
FLEXFLAG_DATABASE_SSL_MODE=disable
FLEXFLAG_DATABASE_MAX_CONNS=10
FLEXFLAG_DATABASE_MIN_CONNS=2

# Redis settings
FLEXFLAG_REDIS_HOST=localhost
FLEXFLAG_REDIS_PORT=6379
FLEXFLAG_REDIS_DATABASE=0
FLEXFLAG_REDIS_PASSWORD=""

# Authentication
FLEXFLAG_JWT_SECRET=your-secret-key-here
FLEXFLAG_JWT_EXPIRY=24h
```

#### Edge Server Configuration
```bash
# Edge server settings
FLEXFLAG_EDGE_PORT=8081
FLEXFLAG_EDGE_ENVIRONMENT=production
FLEXFLAG_EDGE_HUB_URL=http://localhost:8080
FLEXFLAG_EDGE_API_KEY=your_api_key_here

# Edge server cache settings
FLEXFLAG_EDGE_CACHE_TTL=1h
FLEXFLAG_EDGE_MAX_FLAGS=10000
FLEXFLAG_EDGE_MAX_API_KEYS=1000
FLEXFLAG_EDGE_CLEANUP_INTERVAL=5m

# Edge server sync settings
FLEXFLAG_EDGE_RECONNECT_INTERVAL=30s
FLEXFLAG_EDGE_HEARTBEAT_INTERVAL=30s
FLEXFLAG_EDGE_MAX_RETRIES=5
FLEXFLAG_EDGE_ENABLE_SSL=false
```

### Configuration File (Optional)

Create a `config.yaml` file in the project root:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  environment: development

database:
  host: localhost
  port: 5433
  username: flexflag
  password: flexflag
  database: flexflag
  ssl_mode: disable
  max_conns: 10
  min_conns: 2

redis:
  host: localhost
  port: 6379
  database: 0
  password: ""

auth:
  jwt_secret: your-secret-key-here
  jwt_expiry: 24h

logging:
  level: info
  format: json
```

## 🚀 Docker Deployment

### Using Pre-built Images from Docker Hub (Recommended)

FlexFlag provides pre-built Docker images on Docker Hub for quick deployment:

```bash
# Pull the latest images
docker pull <username>/flexflag-api:latest
docker pull <username>/flexflag-ui:latest
docker pull <username>/flexflag-migrator:latest

# Or pull specific version
docker pull <username>/flexflag-api:v1.0.0
docker pull <username>/flexflag-ui:v1.0.0
docker pull <username>/flexflag-migrator:v1.0.0
```

**Available images:**
- `<username>/flexflag-api` - Go API server
- `<username>/flexflag-ui` - Next.js frontend
- `<username>/flexflag-migrator` - Database migration tool

**Quick start with Docker Hub images:**

```bash
# 1. Copy the environment file
cp .env.example .env

# 2. Edit .env and set your Docker Hub username (optional)
# DOCKER_REGISTRY=yourusername/

# 3. Start the stack
docker-compose -f docker-compose.prod.yml up -d

# 4. Check status
docker-compose -f docker-compose.prod.yml ps

# 5. View logs
docker-compose -f docker-compose.prod.yml logs -f
```

**Access the application:**
- 🌐 Frontend: http://localhost:3000
- 🚀 API: http://localhost:8080
- 📖 API Docs: http://localhost:8080/swagger/index.html

### Single Node Deployment (Build from Source)

```bash
# Build Docker images locally
make docker-build

# Run with Docker Compose
docker-compose up -d

# Check service status
docker-compose ps
```

### Multi-Region Edge Deployment

```bash
# Deploy edge servers across multiple regions
docker-compose -f docker-compose.edge.yml up -d

# Scale edge servers
docker-compose -f docker-compose.edge.yml up -d --scale edge-us-east=3
```

### Production Deployment

```bash
# Build production images
docker build -t flexflag:latest .
docker build -f cmd/edge-server/Dockerfile -t flexflag-edge:latest .

# Deploy with production compose file
docker-compose -f docker-compose.prod.yml up -d
```

## 📊 Performance Benchmarks

FlexFlag is optimized for high-performance flag evaluation:

### Current Performance (Local Testing)

- **Standard Evaluation**: ~2-6ms per request
- **Optimized Evaluation**: ~0.5-1ms per request  
- **Ultra-Fast Evaluation**: ~0.1ms per request (317x improvement)
- **Edge Server Evaluation**: <1ms per request
- **Throughput**: >10,000 evaluations/second per edge server

### Performance Testing

```bash
# Run comprehensive performance tests
./benchmark_evaluation.sh

# Test edge server performance
./benchmark-edge.sh

# Concurrent load testing
./concurrent_benchmark.sh

# Results will show:
# - Average response time
# - 95th percentile latency
# - Requests per second
# - Memory usage
```

## 🎯 Feature Flag Usage

### Via UI Dashboard

1. **Create Project**: Navigate to Projects → Create New Project
2. **Manage Environments**: Production, Staging, Development (customizable)
3. **Create Flags**: Boolean, String, Number, JSON, or Variant (A/B testing)
4. **Set Targeting**: User segments, percentage rollouts, custom rules
5. **Monitor Performance**: Real-time metrics and evaluation analytics

### Via REST API

```bash
# Create a feature flag
curl -X POST http://localhost:8080/api/v1/flags \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new-checkout-flow",
    "name": "New Checkout Flow",
    "type": "boolean",
    "enabled": true,
    "default": false,
    "environment": "production",
    "project_id": "your_project_id"
  }'

# Evaluate a flag
curl -X POST http://localhost:8080/api/v1/evaluate \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_key": "new-checkout-flow",
    "user_context": {
      "user_id": "user_123",
      "attributes": {
        "plan": "premium",
        "region": "us-east"
      }
    }
  }'

# Batch evaluate multiple flags
curl -X POST http://localhost:8080/api/v1/evaluate/batch \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_keys": ["new-checkout-flow", "dark-mode", "beta-features"],
    "user_context": {
      "user_id": "user_123",
      "attributes": {"plan": "premium"}
    }
  }'
```

### Via Edge Server (Ultra-Low Latency)

```bash
# Evaluate via nearest edge server
curl -X POST http://your-edge-server:8081/api/v1/evaluate \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_key": "new-checkout-flow",
    "user_context": {
      "user_id": "user_123",
      "attributes": {"region": "us-west"}
    }
  }'
```

## 🧪 Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/api/handlers/...
```

### Integration Tests

```bash
# Start test dependencies
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration

# Test specific functionality
go test -tags=integration ./internal/storage/...
```

### End-to-End Tests

```bash
# Start full application stack
make run &
cd ui && npm run dev &

# Run E2E tests (when implemented)
# npm run test:e2e
```

## 🐛 Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database connectivity
make migrate-status

# Reset database if needed
make migrate-down
make migrate-up
```

#### Edge Server Not Connecting
```bash
# Verify API key is valid
curl -X POST http://localhost:8080/api/v1/edge/auth \
  -H "X-API-Key: your_api_key"

# Check edge server logs
docker-compose logs edge-server

# Verify WebSocket connectivity
curl -H "Upgrade: websocket" \
     -H "Connection: Upgrade" \
     -H "X-API-Key: your_api_key" \
     http://localhost:8080/api/v1/edge/sync/ws
```

#### UI Build Issues
```bash
# Clear Node.js cache
cd ui
rm -rf .next node_modules package-lock.json
npm install
npm run build
```

#### Performance Issues
```bash
# Check resource usage
docker stats

# Monitor database connections
docker-compose exec postgres psql -U flexflag -c "SELECT * FROM pg_stat_activity;"

# Check Redis memory usage
docker-compose exec redis redis-cli info memory
```

### Debug Mode

```bash
# Start server with debug logging
FLEXFLAG_LOG_LEVEL=debug make run

# Enable Go race detector
go run -race cmd/server/main.go

# Profile memory usage
go tool pprof http://localhost:8080/debug/pprof/heap
```

## 📚 API Documentation

### Interactive Documentation
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/profile` - Get user profile

#### Projects & Environments
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{slug}/environments` - List environments

#### Feature Flags
- `GET /api/v1/flags` - List flags
- `POST /api/v1/flags` - Create flag
- `PUT /api/v1/flags/{key}` - Update flag
- `POST /api/v1/flags/{key}/toggle` - Toggle flag

#### Flag Evaluation
- `POST /api/v1/evaluate` - Evaluate single flag
- `POST /api/v1/evaluate/batch` - Batch evaluate flags
- `POST /api/v1/evaluate/ultra` - Ultra-fast evaluation

#### Edge Servers
- `GET /api/v1/edge/servers` - List connected edge servers
- `GET /api/v1/edge/sync` - Bulk sync endpoint
- `GET /api/v1/edge/sync/ws` - WebSocket sync endpoint

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Run the test suite**: `make test`
6. **Commit your changes**: `git commit -m 'feat: add amazing feature'`
7. **Push to the branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Development Guidelines

- Follow [Conventional Commits](https://conventionalcommits.org/) for commit messages
- Add tests for new features and bug fixes
- Update documentation for API changes
- Run `make lint` before submitting
- Ensure all tests pass with `make test`

## 📄 License

FlexFlag is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## 🔗 Links

- **GitHub**: [github.com/flexflag/flexflag](https://github.com/flexflag/flexflag)
- **Documentation**: [flexflag.github.io](https://flexflag.github.io/FlexFlag/)
- **Issues**: [github.com/flexflag/flexflag/issues](https://github.com/flexflag/flexflag/issues)
- **Discord**: [Join our community](https://discord.gg/fpewTJyx9S)

## 💡 Architecture Overview

FlexFlag uses a modern, distributed architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Next.js UI    │    │  Edge Server    │    │  Edge Server    │
│  (Port 3000)    │    │  (Port 8081)    │    │  (Port 8082)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │          SSE/WebSocket│          SSE/WebSocket│
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                     Main FlexFlag Server                        │
│                        (Port 8080)                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   REST API  │ │    SSE/WS   │ │   Auth      │ │  Sync       ││
│  │   Handlers  │ │   Handler   │ │  Service    │ │  Service    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
         ┌─────────────────────────────────────────────────────────┐
         │                Infrastructure                           │
         │  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐│
         │  │ PostgreSQL  │     │    Redis    │     │   Docker    ││
         │  │ (Port 5433) │     │ (Port 6379) │     │  Compose    ││
         │  └─────────────┘     └─────────────┘     └─────────────┘│
         └─────────────────────────────────────────────────────────┘
```

### Performance Tiers

1. **Standard Evaluation** (~2-6ms): Direct database queries
2. **Optimized Evaluation** (~0.5-1ms): With Redis caching  
3. **Ultra-Fast Evaluation** (~0.1ms): Pre-computed in-memory flags
4. **Edge Evaluation** (<1ms): Distributed edge servers with local cache

Built with ❤️ by the FlexFlag team
