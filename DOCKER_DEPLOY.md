# FlexFlag Docker Deployment Guide

Complete guide for deploying FlexFlag using Docker and Docker Compose.

## Table of Contents

- [Quick Start](#quick-start)
- [Docker Hub Deployment](#docker-hub-deployment)
- [Production Deployment](#production-deployment)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- Docker 20.10+ installed
- Docker Compose 2.0+ installed
- 2GB RAM minimum
- 10GB disk space

### 1. Pull and Run (Simplest Method)

```bash
# Pull the latest images from Docker Hub
docker pull YOUR_USERNAME/flexflag-api:latest
docker pull YOUR_USERNAME/flexflag-ui:latest
docker pull YOUR_USERNAME/flexflag-migrator:latest

# Run using production docker-compose
curl -o docker-compose.prod.yml https://raw.githubusercontent.com/YOUR_ORG/flexflag/main/docker-compose.prod.yml
curl -o .env.example https://raw.githubusercontent.com/YOUR_ORG/flexflag/main/.env.example

# Configure environment
cp .env.example .env
# Edit .env with your values

# Start FlexFlag
docker-compose -f docker-compose.prod.yml up -d
```

### 2. Access FlexFlag

- **Frontend UI**: http://localhost:3000
- **API Server**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

### 3. Default Credentials

```
Email: admin@flexflag.com
Password: admin123
```

**⚠️ IMPORTANT**: Change the admin password immediately after first login!

---

## Docker Hub Deployment

### For Maintainers: Publishing to Docker Hub

#### 1. Login to Docker Hub

```bash
docker login
```

#### 2. Build and Tag Images

```bash
# Set your Docker Hub username
export DOCKER_USERNAME=your-dockerhub-username
export VERSION=1.0.0

# Build all images
make docker-build-all VERSION=$VERSION DOCKER_USERNAME=$DOCKER_USERNAME
```

#### 3. Push to Docker Hub

```bash
# Push all images
make docker-push-all VERSION=$VERSION DOCKER_USERNAME=$DOCKER_USERNAME
```

#### 4. Verify Images

```bash
# Check published images
docker pull $DOCKER_USERNAME/flexflag-api:$VERSION
docker pull $DOCKER_USERNAME/flexflag-ui:$VERSION
docker pull $DOCKER_USERNAME/flexflag-migrator:$VERSION
```

---

## Production Deployment

### 1. Environment Configuration

Create a `.env` file with production values:

```bash
# Copy example and edit
cp .env.example .env
```

**Critical Production Settings:**

```env
# Generate secure secrets
JWT_SECRET=$(openssl rand -base64 32)
POSTGRES_PASSWORD=$(openssl rand -base64 24)

# Production URLs
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Database (use managed service)
POSTGRES_USER=flexflag_prod
POSTGRES_PASSWORD=<secure-password>
DATABASE_SSL_MODE=require

# Disable debug
DEBUG=false

# Optional: Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=https://api.yourdomain.com/api/v1/auth/google/callback
```

### 2. Deploy with Docker Compose

```bash
# Using Docker Hub images
DOCKER_REGISTRY=username/ VERSION=1.0.0 docker-compose -f docker-compose.prod.yml up -d

# Or build locally
docker-compose -f docker-compose.prod.yml up -d --build
```

### 3. Health Checks

```bash
# Check all services
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Check individual service
docker-compose -f docker-compose.prod.yml logs api
```

### 4. Backup Database

```bash
# Backup
docker exec flexflag-postgres pg_dump -U flexflag flexflag > backup.sql

# Restore
docker exec -i flexflag-postgres psql -U flexflag flexflag < backup.sql
```

---

## Configuration

### Environment Variables

#### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_PASSWORD` | Database password | `secure_password_123` |
| `JWT_SECRET` | JWT signing secret | `openssl rand -base64 32` |
| `NEXT_PUBLIC_API_URL` | API URL for frontend | `http://localhost:8080` |

#### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_PORT` | API server port | `8080` |
| `UI_PORT` | Frontend port | `3000` |
| `POSTGRES_PORT` | PostgreSQL port | `5433` |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | `http://localhost:3000` |
| `CACHE_ENABLED` | Enable Redis caching | `true` |
| `DEBUG` | Debug mode | `false` |

#### Google OAuth (Optional)

| Variable | Description |
|----------|-------------|
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL |

See [GOOGLE_OAUTH_SETUP.md](docs/GOOGLE_OAUTH_SETUP.md) for setup instructions.

### Port Mapping

| Service | Internal Port | External Port (Default) |
|---------|---------------|-------------------------|
| Frontend UI | 3000 | 3000 |
| API Server | 8080 | 8080 |
| PostgreSQL | 5432 | 5433 |
| Redis | 6379 | 6379 |

---

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Test connection
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT 1;"
```

### Migration Failures

```bash
# Check migrator logs
docker-compose logs migrator

# Run migrations manually
docker exec flexflag-api ./migrator -database-url="postgres://flexflag:flexflag@postgres:5432/flexflag?sslmode=disable" -direction=up
```

### API Health Check Failing

```bash
# Check API logs
docker-compose logs api

# Verify health endpoint
curl http://localhost:8080/health

# Check environment variables
docker exec flexflag-api env | grep FLEXFLAG
```

### Frontend Not Loading

```bash
# Check UI logs
docker-compose logs ui

# Verify API connection
docker exec flexflag-ui curl http://flexflag-api:8080/health

# Check environment
docker exec flexflag-ui env | grep NEXT_PUBLIC
```

### Reset Everything

```bash
# Stop and remove all containers
docker-compose -f docker-compose.prod.yml down -v

# Remove all FlexFlag data
docker volume rm flexflag_postgres_data flexflag_redis_data

# Start fresh
docker-compose -f docker-compose.prod.yml up -d
```

---

## Makefile Commands

Quick reference for Docker operations:

```bash
# Build all Docker images
make docker-build-all

# Push images to Docker Hub
make docker-push-all DOCKER_USERNAME=your-username

# Pull images from Docker Hub
make docker-pull-all DOCKER_USERNAME=your-username

# Tag images for release
make docker-tag VERSION=1.0.0

# Start production stack
make docker-prod-up

# Stop production stack
make docker-prod-down

# View logs
make docker-logs

# Clean up
make docker-clean
```

---

## Advanced Configuration

### Using External Database

```yaml
# docker-compose.prod.yml
services:
  api:
    environment:
      FLEXFLAG_DATABASE_HOST: your-db-host.com
      FLEXFLAG_DATABASE_PORT: 5432
      FLEXFLAG_DATABASE_SSL_MODE: require
```

### Using External Redis

```yaml
services:
  api:
    environment:
      FLEXFLAG_REDIS_HOST: your-redis-host.com
      FLEXFLAG_REDIS_PASSWORD: your-redis-password
```

### Scaling Services

```bash
# Scale API servers
docker-compose -f docker-compose.prod.yml up -d --scale api=3

# Use load balancer (nginx, traefik, etc.)
```

### SSL/TLS Configuration

Use nginx or traefik as reverse proxy:

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
```

---

## Monitoring

### Health Checks

```bash
# All services
curl http://localhost:8080/health

# Database
docker exec flexflag-postgres pg_isready

# Redis
docker exec flexflag-redis redis-cli ping
```

### Logs

```bash
# Follow all logs
docker-compose logs -f

# Specific service
docker-compose logs -f api

# Last 100 lines
docker-compose logs --tail=100
```

### Metrics

Access built-in metrics at:
- API Metrics: http://localhost:8080/metrics (if enabled)

---

## Security Best Practices

1. **Change default passwords** immediately
2. **Use strong JWT secrets** (32+ characters)
3. **Enable SSL/TLS** in production
4. **Restrict CORS origins** to your domains
5. **Use secrets management** (Docker secrets, Vault, etc.)
6. **Regular backups** of PostgreSQL data
7. **Keep images updated** regularly
8. **Use non-root users** (already configured)
9. **Scan images** for vulnerabilities

---

## Support

- **Documentation**: https://docs.flexflag.com
- **Issues**: https://github.com/YOUR_ORG/flexflag/issues
- **Community**: Join our Discord/Slack

---

## License

MIT License - See [LICENSE](LICENSE) file for details
