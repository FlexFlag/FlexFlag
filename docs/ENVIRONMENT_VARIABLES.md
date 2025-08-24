# FlexFlag Environment Variables Reference

This document provides a comprehensive reference for all FlexFlag environment variables.

## Quick Start

1. Copy `.env.example` to `.env`
2. Run `./setup-dev.sh` for automated development setup
3. Or manually configure the variables below

## Variable Categories

### 🚀 Core Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_SERVER_HOST` | `0.0.0.0` | Server bind address |
| `FLEXFLAG_SERVER_PORT` | `8080` | Server port |
| `FLEXFLAG_SERVER_DEBUG` | `false` | Enable debug mode |
| `FLEXFLAG_SERVER_ENVIRONMENT` | `development` | Runtime environment |
| `FLEXFLAG_CORS_ALLOWED_ORIGINS` | `"http://localhost:3000"` | Comma-separated CORS origins |

### 🗄️ Database Configuration (PostgreSQL)

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_DATABASE_HOST` | `localhost` | Database host |
| `FLEXFLAG_DATABASE_PORT` | `5432` | Database port |
| `FLEXFLAG_DATABASE_USERNAME` | `flexflag` | Database username |
| `FLEXFLAG_DATABASE_PASSWORD` | `flexflag` | Database password |
| `FLEXFLAG_DATABASE_DATABASE` | `flexflag` | Database name |
| `FLEXFLAG_DATABASE_SSL_MODE` | `disable` | SSL mode (disable/require/verify-ca/verify-full) |
| `FLEXFLAG_DATABASE_MAX_CONNS` | `10` | Maximum connections in pool |
| `FLEXFLAG_DATABASE_MIN_CONNS` | `2` | Minimum connections in pool |

### 🔴 Redis Configuration (Caching)

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_REDIS_HOST` | `localhost` | Redis host |
| `FLEXFLAG_REDIS_PORT` | `6379` | Redis port |
| `FLEXFLAG_REDIS_PASSWORD` | `""` | Redis password |
| `FLEXFLAG_REDIS_DATABASE` | `0` | Redis database number |

### 🔐 Authentication & Security

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_JWT_SECRET` | `"change-me"` | JWT signing secret (use 256-bit key) |
| `FLEXFLAG_JWT_EXPIRY` | `24h` | JWT expiration time |

### ⚡ Performance & Caching

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_CACHE_ENABLED` | `true` | Enable in-memory caching |
| `FLEXFLAG_CACHE_TTL` | `300` | Cache TTL in seconds |
| `FLEXFLAG_METRICS_ENABLED` | `true` | Enable metrics collection |

### 🌐 Edge Server Configuration

#### Core Settings
| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_EDGE_PORT` | `8081` | Edge server port |
| `FLEXFLAG_EDGE_ENVIRONMENT` | `production` | Edge server environment |
| `FLEXFLAG_HUB_URL` | `http://localhost:8080` | FlexFlag hub URL |
| `FLEXFLAG_EDGE_API_KEY` | `""` | API key for hub authentication |

#### Cache Settings
| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_EDGE_MAX_FLAGS` | `10000` | Maximum flags to cache |
| `FLEXFLAG_EDGE_MAX_API_KEYS` | `1000` | Maximum API keys to cache |
| `FLEXFLAG_EDGE_CACHE_TTL` | `1h` | Cache entry TTL |
| `FLEXFLAG_EDGE_CLEANUP_INTERVAL` | `5m` | Cache cleanup interval |
| `FLEXFLAG_EDGE_ENABLE_METRICS` | `true` | Enable metrics collection |

#### Sync Settings
| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_EDGE_SYNC_TYPE` | `websocket` | Sync protocol (`websocket` or `sse`) |
| `FLEXFLAG_EDGE_RECONNECT_INTERVAL` | `30s` | Reconnection interval |
| `FLEXFLAG_EDGE_HEARTBEAT_INTERVAL` | `30s` | Heartbeat interval |
| `FLEXFLAG_EDGE_MAX_RETRIES` | `5` | Maximum connection retries |
| `FLEXFLAG_EDGE_BUFFER_SIZE` | `1000` | Message buffer size |
| `FLEXFLAG_EDGE_ENABLE_SSL` | `false` | Enable SSL for connections |

### 🎨 UI Configuration (Next.js)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | API server URL for frontend |

### 📊 Logging & Debugging

| Variable | Default | Description |
|----------|---------|-------------|
| `FLEXFLAG_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `FLEXFLAG_LOG_FORMAT` | `json` | Log format (json/text) |
| `FLEXFLAG_LOG_OUTPUT` | `stdout` | Log output destination |

## Configuration Scenarios

### Local Development
```bash
# Copy the example file
cp .env.example .env

# Run the setup script
./setup-dev.sh

# Or start manually
docker-compose up -d postgres redis
make run
```

### Docker Compose Development
```bash
# Use docker-compose.dev.yml
docker-compose -f docker-compose.dev.yml up
```

### Production Deployment

#### 1. Security Configuration
```bash
# Generate secure JWT secret
FLEXFLAG_JWT_SECRET=$(openssl rand -base64 32)

# Set production CORS
FLEXFLAG_CORS_ALLOWED_ORIGINS="https://yourdomain.com,https://app.yourdomain.com"

# Use production environment
FLEXFLAG_SERVER_ENVIRONMENT=production
```

#### 2. Database Configuration
```bash
FLEXFLAG_DATABASE_HOST=your-db-host.com
FLEXFLAG_DATABASE_PORT=5432
FLEXFLAG_DATABASE_USERNAME=flexflag_prod
FLEXFLAG_DATABASE_PASSWORD=secure_password_here
FLEXFLAG_DATABASE_SSL_MODE=require
FLEXFLAG_DATABASE_MAX_CONNS=50
```

#### 3. Redis Configuration
```bash
FLEXFLAG_REDIS_HOST=your-redis-host.com
FLEXFLAG_REDIS_PORT=6379
FLEXFLAG_REDIS_PASSWORD=redis_password_here
```

#### 4. Performance Tuning
```bash
FLEXFLAG_CACHE_TTL=600
FLEXFLAG_EDGE_MAX_FLAGS=50000
FLEXFLAG_EDGE_CACHE_TTL=2h
FLEXFLAG_EDGE_BUFFER_SIZE=5000
```

### Edge Server Deployment

#### 1. Generate API Key
1. Login to FlexFlag dashboard
2. Navigate to project → API Keys
3. Create new API key with appropriate permissions
4. Copy the generated key

#### 2. Configure Edge Server
```bash
FLEXFLAG_EDGE_API_KEY=your_generated_api_key
FLEXFLAG_HUB_URL=https://your-flexflag-server.com
FLEXFLAG_EDGE_SYNC_TYPE=sse
FLEXFLAG_EDGE_PORT=8081
```

#### 3. High-Performance Settings
```bash
FLEXFLAG_EDGE_MAX_FLAGS=50000
FLEXFLAG_EDGE_MAX_API_KEYS=5000
FLEXFLAG_EDGE_CACHE_TTL=2h
FLEXFLAG_EDGE_BUFFER_SIZE=5000
FLEXFLAG_EDGE_CLEANUP_INTERVAL=1m
```

## SDK Integration

### Node.js Applications
```javascript
// In your application's .env
FLEXFLAG_API_KEY=your_client_api_key
FLEXFLAG_API_URL=http://localhost:8080
FLEXFLAG_ENVIRONMENT=production
```

### Python Applications
```python
# Django settings.py or .env
FLEXFLAG_CONFIG = {
    'api_key': 'your_client_api_key',
    'api_url': 'http://localhost:8080',
    'environment': 'production',
    'timeout': 5000,
    'cache_enabled': True
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check `FLEXFLAG_DATABASE_*` variables
   - Ensure PostgreSQL is running
   - Verify network connectivity

2. **JWT Token Invalid**
   - Ensure `FLEXFLAG_JWT_SECRET` is set and consistent
   - Check token expiry with `FLEXFLAG_JWT_EXPIRY`

3. **Edge Server Not Connecting**
   - Verify `FLEXFLAG_EDGE_API_KEY` is valid
   - Check `FLEXFLAG_HUB_URL` is correct
   - Ensure edge server can reach the hub

4. **CORS Errors**
   - Update `FLEXFLAG_CORS_ALLOWED_ORIGINS`
   - Include all domains that will access the API

### Validation Commands

```bash
# Test database connection
make test-db

# Verify API server is running
curl http://localhost:8080/health

# Check edge server status
curl http://localhost:8080/api/v1/edge/servers

# Test flag evaluation
curl -X POST http://localhost:8080/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"flag_key": "test-flag", "user_id": "user123"}'
```

## Best Practices

1. **Never commit secrets to version control**
2. **Use different JWT secrets per environment**
3. **Enable SSL in production (`FLEXFLAG_DATABASE_SSL_MODE=require`)**
4. **Set appropriate connection pool sizes for your load**
5. **Monitor cache hit rates and adjust TTL values**
6. **Use Redis for distributed caching in production**
7. **Configure log levels appropriately for each environment**

## Migration from Older Versions

If upgrading from older FlexFlag versions, check for:
- Changed environment variable names
- New required configuration options
- Database schema migrations
- Updated default values