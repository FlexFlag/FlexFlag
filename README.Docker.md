# FlexFlag Docker Setup

This document provides comprehensive instructions for running FlexFlag using Docker and Docker Compose.

## Quick Start

### 1. Production Setup (Recommended)

```bash
# Start the entire stack
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the stack
docker-compose down
```

**Services:**
- **UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Database**: localhost:5433
- **Redis**: localhost:6379
- **Nginx**: http://localhost (load balancer)

### 2. Development Setup (with hot reload)

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop development environment
docker-compose -f docker-compose.dev.yml down
```

**Additional Development Services:**
- **pgAdmin**: http://localhost:5050 (admin@flexflag.com / admin)
- **Redis Commander**: http://localhost:8081
- **API Debugger**: Port 2345 (Delve)

## Configuration

### Environment Variables

The following environment variables can be customized in `docker-compose.yml`:

#### Database
- `POSTGRES_USER=flexflag`
- `POSTGRES_PASSWORD=flexflag`
- `POSTGRES_DB=flexflag`

#### API Server
- `FLEXFLAG_SERVER_HOST=0.0.0.0`
- `FLEXFLAG_SERVER_PORT=8080`
- `FLEXFLAG_SERVER_DEBUG=true`
- `FLEXFLAG_JWT_SECRET=your-super-secret-jwt-key`
- `FLEXFLAG_CACHE_ENABLED=true`
- `FLEXFLAG_CACHE_TTL=300`

#### Frontend
- `NEXT_PUBLIC_API_URL=http://localhost:8080`
- `NODE_ENV=development`

## Docker Images

### API Server (`Dockerfile.api`)
- **Base**: `golang:1.21-alpine`
- **Features**: Multi-stage build, non-root user, health checks
- **Size**: ~20MB (optimized)

### Frontend (`ui/Dockerfile`)
- **Base**: `node:18-alpine`
- **Features**: Multi-stage build, Next.js optimization, static assets
- **Size**: ~150MB

### Database Migrator (`Dockerfile.migrator`)
- **Base**: `golang:1.21-alpine`
- **Purpose**: Runs database migrations once on startup
- **Usage**: One-time execution container

## Development Features

### Hot Reload
- **API**: Uses [Air](https://github.com/cosmtrek/air) for Go hot reload
- **UI**: Next.js development server with file watching
- **Configuration**: Automatic restart on code changes

### Debugging
- **API**: Delve debugger on port 2345
- **UI**: Chrome DevTools integration
- **Logs**: Real-time log streaming

### Database Management
- **pgAdmin**: Web-based PostgreSQL administration
- **Redis Commander**: Redis key-value store management
- **Migrations**: Automatic database schema updates

## Production Considerations

### Security
```yaml
# Change these in production:
environment:
  FLEXFLAG_JWT_SECRET: "your-production-secret-key"
  POSTGRES_PASSWORD: "secure-password"
```

### SSL/TLS
Uncomment SSL configuration in `nginx.conf`:
```bash
# Generate SSL certificates
mkdir ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem -out ssl/cert.pem

# Update nginx.conf SSL section
```

### Performance Tuning
```yaml
# Adjust based on your needs:
environment:
  FLEXFLAG_CACHE_TTL: 600
  FLEXFLAG_REDIS_MAX_CONNECTIONS: 100
```

## Commands

### Basic Operations
```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up -d postgres redis

# Restart a service
docker-compose restart api

# View service logs
docker-compose logs -f api

# Execute command in container
docker-compose exec api sh
docker-compose exec postgres psql -U flexflag

# Remove everything (including volumes)
docker-compose down -v
```

### Development Commands
```bash
# Development environment
docker-compose -f docker-compose.dev.yml up -d

# Rebuild after code changes
docker-compose -f docker-compose.dev.yml up --build

# Run tests
docker-compose exec api-dev go test ./...

# Run database migrations manually
docker-compose exec api-dev ./migrator

# Access development database
docker-compose exec postgres psql -U flexflag -d flexflag
```

### Maintenance
```bash
# Update images
docker-compose pull

# Clean up unused resources
docker system prune -f

# Backup database
docker-compose exec postgres pg_dump -U flexflag flexflag > backup.sql

# Restore database
docker-compose exec -T postgres psql -U flexflag flexflag < backup.sql
```

## Troubleshooting

### Common Issues

1. **Port conflicts**
   ```bash
   # Check if ports are in use
   lsof -i :3000 -i :8080 -i :5433 -i :6379
   
   # Change ports in docker-compose.yml
   ports:
     - "3001:3000"  # UI on port 3001 instead
   ```

2. **Permission errors**
   ```bash
   # Fix file permissions
   sudo chown -R $USER:$USER .
   
   # Reset Docker volumes
   docker-compose down -v
   docker-compose up -d
   ```

3. **Build failures**
   ```bash
   # Clean build cache
   docker-compose build --no-cache
   
   # Remove old images
   docker image prune -f
   ```

4. **Database connection issues**
   ```bash
   # Check database health
   docker-compose exec postgres pg_isready -U flexflag
   
   # View database logs
   docker-compose logs postgres
   ```

### Health Checks

All services include health checks:
```bash
# Check service health
docker-compose ps

# View health status
docker inspect flexflag-api | grep Health -A 20
```

## Performance Monitoring

### Resource Usage
```bash
# Monitor resource usage
docker stats

# Service-specific stats
docker stats flexflag-api flexflag-ui flexflag-postgres
```

### Log Analysis
```bash
# Follow all logs
docker-compose logs -f

# API performance logs
docker-compose logs api | grep "evaluation_time"

# Error logs only
docker-compose logs api | grep ERROR
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Nginx (80)    │────│  Frontend (3000)│
│  Load Balancer  │    │    Next.js      │
└─────────────────┘    └─────────────────┘
         │                       │
         │              ┌─────────────────┐
         └──────────────│   API (8080)    │
                        │      Go         │
                        └─────────────────┘
                                 │
                   ┌─────────────┼─────────────┐
                   │             │             │
         ┌─────────────────┐ ┌─────────────────┐
         │ PostgreSQL      │ │     Redis       │
         │   (5433)        │ │    (6379)       │
         └─────────────────┘ └─────────────────┘
```

## Next Steps

1. **Customize configuration** in `docker-compose.yml`
2. **Set up SSL certificates** for production
3. **Configure monitoring** (Prometheus, Grafana)
4. **Set up CI/CD pipeline** with Docker builds
5. **Scale services** using Docker Swarm or Kubernetes

For more information, see the main [README.md](README.md) file.