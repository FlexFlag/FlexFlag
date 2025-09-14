# FlexFlag Troubleshooting Guide

This guide helps resolve common setup and development issues with FlexFlag.

## 🚨 Common Setup Issues

### 1. Migration Failures

**Symptom:** Setup fails with "migration failed" or "dirty database" errors

**Solutions:**
```bash
# Option 1: Clean setup (recommended)
./scripts/clean-setup.sh
./setup-dev.sh

# Option 2: Force migration reset
go run cmd/migrator/main.go -database-url "postgres://flexflag:flexflag@localhost:5433/flexflag?sslmode=disable" -force-version 0
make migrate-up

# Option 3: Manual migration troubleshooting
make migrate-down
make migrate-up
```

### 2. Docker Issues

**Symptom:** "Failed to start Docker services" or container errors

**Solutions:**
```bash
# Check if Docker is running
docker ps

# Clean up containers and volumes
docker-compose -f docker-compose.dev.yml down
docker system prune -f

# Restart Docker Desktop/daemon and try again
./setup-dev.sh
```

### 3. Port Conflicts

**Symptom:** "Port already in use" errors

**Solutions:**
```bash
# Check what's using the ports
lsof -i :5433  # PostgreSQL
lsof -i :6379  # Redis
lsof -i :8080  # API Server
lsof -i :3000  # UI

# Kill processes or change ports in .env file
# Update FLEXFLAG_DATABASE_PORT, FLEXFLAG_REDIS_PORT, etc.
```

### 4. Node.js/npm Issues

**Symptom:** UI installation fails or npm errors

**Solutions:**
```bash
# Clean npm cache and reinstall
cd ui
rm -rf node_modules package-lock.json
npm cache clean --force
npm install

# Update Node.js if version is too old (requires 18.17+)
# Use nvm, fnm, or download from nodejs.org
```

### 5. Go Build Issues

**Symptom:** "go build" fails or missing dependencies

**Solutions:**
```bash
# Clean Go module cache
go clean -modcache
go mod download
go mod tidy

# Rebuild
make clean
make build
```

## 🔍 Debugging Commands

### Check Service Health
```bash
# Database connection
docker exec flexflag-postgres pg_isready -U flexflag -d flexflag

# Redis connection  
docker exec flexflag-redis redis-cli ping

# API server health
curl http://localhost:8080/health

# Check logs
docker-compose logs postgres
docker-compose logs redis
```

### Migration Status
```bash
# Check current migration version
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# List all tables
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "\dt"
```

### Performance Issues
```bash
# Check resource usage
docker stats

# Database performance
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT * FROM pg_stat_activity;"

# API benchmarks
./benchmark_evaluation.sh
```

## 🛠️ Environment Issues

### Environment Variables Not Loading
```bash
# Check .env file exists and has correct values
cat .env | grep FLEXFLAG_

# Verify environment loading
make run | head -20  # Should show config values
```

### JWT Secret Issues
```bash
# Regenerate JWT secret
openssl rand -base64 32

# Update in .env
FLEXFLAG_JWT_SECRET="your-new-secret-here"
```

## 🚀 Performance Troubleshooting

### Slow Flag Evaluation
```bash
# Check evaluation endpoint performance
curl -X POST http://localhost:8080/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{"flag_key":"test","user_id":"user1","environment":"production"}'

# Use ultra-fast endpoint
curl -X POST http://localhost:8080/api/v1/evaluate/ultra \
  -H "Content-Type: application/json" \
  -d '{"flag_key":"test","user_id":"user1","environment":"production"}'

# Check cache stats
curl http://localhost:8080/api/v1/evaluate/cache/stats
```

### Database Performance
```bash
# Check connection pool
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT count(*) FROM pg_stat_activity;"

# Analyze slow queries
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT query, calls, total_time, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

## 📊 Monitoring & Logs

### Application Logs
```bash
# API server logs (when running via make run)
tail -f logs/flexflag.log

# Docker container logs
docker logs flexflag-postgres
docker logs flexflag-redis
docker-compose logs -f
```

### Metrics & Health Checks
```bash
# Server metrics
curl http://localhost:8080/api/v1/evaluate/ultra/stats

# System health
curl http://localhost:8080/health

# Database stats
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT * FROM pg_database_size('flexflag');"
```

## 🆘 Getting Help

### Gather Debug Information
```bash
# System info
echo "OS: $(uname -a)"
echo "Go: $(go version)"
echo "Node: $(node --version)"
echo "Docker: $(docker --version)"

# FlexFlag setup info
docker ps --filter "name=flexflag"
docker images --filter "reference=*flexflag*"
```

### Reset Everything (Nuclear Option)
```bash
# ⚠️ This destroys all data!
./scripts/clean-setup.sh
rm .env
./setup-dev.sh
```

### Create GitHub Issue
When creating an issue, please include:
1. Operating system and version
2. Go, Node.js, and Docker versions
3. Error messages and logs
4. Steps to reproduce
5. Output of debug commands above

**GitHub Issues:** https://github.com/FlexFlag/FlexFlag/issues