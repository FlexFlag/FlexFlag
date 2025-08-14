# FlexFlag Edge Server Deployment Guide

This guide covers deploying FlexFlag edge servers for ultra-low latency flag evaluation in distributed environments.

## Quick Start

### 1. Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- API key from your FlexFlag central hub

### 2. Configuration

Set the required environment variable:

```bash
export FLEXFLAG_EDGE_API_KEY="your-api-key-here"
```

### 3. Deploy Edge Infrastructure

```bash
# Deploy complete edge infrastructure
make edge-deploy

# Or use the deployment script directly
./deployments/deploy-edge.sh deploy
```

### 4. Verify Deployment

```bash
# Check status
make edge-status

# Test health endpoints
curl http://localhost:8081/health
curl http://localhost:8082/health

# View edge cache statistics
curl http://localhost:8081/api/v1/cache/stats
```

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client Apps   │    │  Load Balancer   │    │  Edge Servers   │
│                 │───▶│    (Nginx)       │───▶│  (Ultra-fast    │
│  Web/Mobile/API │    │   Port: 8082     │    │   Evaluation)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                ▲                        │
                                │                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │   Central Hub    │◀───│  Real-time Sync │
                       │ (Main FlexFlag)  │    │   (WebSocket)   │
                       │   Port: 8080     │    └─────────────────┘
                       └──────────────────┘
```

## Service Endpoints

| Service | Port | Endpoint | Purpose |
|---------|------|----------|---------|
| Central Hub | 8080 | http://localhost:8080 | Main FlexFlag API |
| Edge Server | 8081 | http://localhost:8081 | Direct edge access |
| Load Balancer | 8082 | http://localhost:8082 | Production endpoint |

## API Usage

### Flag Evaluation

```bash
# Single flag evaluation (ultra-fast)
curl -X POST http://localhost:8082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "flag_key": "feature-toggle",
    "user_id": "user123",
    "user_key": "user123",
    "attributes": {
      "email": "user@example.com",
      "plan": "premium"
    }
  }'
```

### Batch Evaluation

```bash
# Multiple flags in one request
curl -X POST http://localhost:8082/api/v1/evaluate/batch \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "flag_keys": ["feature-a", "feature-b", "feature-c"],
    "user_id": "user123",
    "user_key": "user123",
    "attributes": {
      "email": "user@example.com"
    }
  }'
```

### Cache Statistics

```bash
# View edge server performance metrics
curl http://localhost:8081/api/v1/cache/stats
```

Expected response:
```json
{
  "flag_count": 150,
  "api_key_count": 25,
  "hit_rate": 0.98,
  "total_requests": 10000,
  "cache_hits": 9800,
  "cache_misses": 200,
  "last_updated": "2024-01-15T10:30:00Z",
  "uptime_seconds": 3600
}
```

## Performance Targets

| Metric | Target | Typical |
|--------|--------|---------|
| P50 Latency | < 0.5ms | ~0.2ms |
| P95 Latency | < 1ms | ~0.4ms |
| P99 Latency | < 2ms | ~0.8ms |
| Cache Hit Rate | > 95% | ~98% |
| Throughput | > 10k RPS | ~15k RPS |

## Configuration

### Environment Variables

#### Edge Server Configuration
```bash
# Server settings
FLEXFLAG_EDGE_PORT=8081
FLEXFLAG_EDGE_ENVIRONMENT=production
FLEXFLAG_HUB_URL=http://localhost:8080

# Authentication
FLEXFLAG_EDGE_API_KEY=your-api-key

# Cache settings
FLEXFLAG_EDGE_MAX_FLAGS=10000
FLEXFLAG_EDGE_MAX_API_KEYS=1000
FLEXFLAG_EDGE_CACHE_TTL=1h
FLEXFLAG_EDGE_CLEANUP_INTERVAL=5m
FLEXFLAG_EDGE_ENABLE_METRICS=true

# Sync settings
FLEXFLAG_EDGE_RECONNECT_INTERVAL=30s
FLEXFLAG_EDGE_HEARTBEAT_INTERVAL=30s
FLEXFLAG_EDGE_MAX_RETRIES=5
FLEXFLAG_EDGE_BUFFER_SIZE=1000
```

## Deployment Commands

### Make Commands

```bash
# Build edge server binary
make build-edge

# Run locally for development
make run-edge

# Build Docker image
make docker-build-edge

# Deploy complete infrastructure
make edge-deploy

# Scale to multiple replicas
EDGE_REPLICAS=3 make edge-scale

# Check deployment status
make edge-status

# Run performance tests
make edge-test

# View logs
make edge-logs

# Stop all edge services
make edge-stop
```

### Deployment Script

```bash
# Direct script usage
./deployments/deploy-edge.sh build     # Build only
./deployments/deploy-edge.sh deploy    # Full deployment
./deployments/deploy-edge.sh scale     # Scale servers
./deployments/deploy-edge.sh status    # Show status
./deployments/deploy-edge.sh test      # Performance test
./deployments/deploy-edge.sh logs      # View logs
./deployments/deploy-edge.sh stop      # Stop services
```

## Scaling

### Horizontal Scaling

```bash
# Scale to 3 edge server replicas
EDGE_REPLICAS=3 make edge-scale
```

### Geographic Distribution

For multiple regions, deploy edge servers in different locations:

```bash
# Region 1 (US-East)
FLEXFLAG_HUB_URL=https://us-east.flexflag.com \
EDGE_REPLICAS=2 make edge-deploy

# Region 2 (EU-West) 
FLEXFLAG_HUB_URL=https://eu-west.flexflag.com \
EDGE_REPLICAS=2 make edge-deploy
```

## Monitoring

### Health Checks

- **Liveness**: `GET /health` - Basic server health
- **Readiness**: `GET /ready` - Cache populated and sync active
- **Metrics**: `GET /api/v1/cache/stats` - Performance statistics

### Nginx Load Balancer Monitoring

```bash
# Nginx status (only from localhost/internal networks)
curl http://localhost:8082:8080/nginx_status
```

### Docker Health Checks

All services include Docker health checks:

```bash
# Check container health
docker-compose -f docker-compose.edge.yml ps

# View health check logs
docker inspect flexflag_flexflag-edge_1 | grep -A 10 "Health"
```

## Troubleshooting

### Common Issues

#### 1. Edge Server Not Starting

Check logs:
```bash
make edge-logs
```

Common causes:
- Invalid API key
- Cannot connect to central hub
- Port conflicts

#### 2. High Cache Miss Rate

Check sync status:
```bash
curl http://localhost:8081/ready
```

Solutions:
- Verify WebSocket connection to hub
- Check API key permissions
- Force cache refresh: `POST /api/v1/cache/refresh`

#### 3. Performance Issues

Monitor metrics:
```bash
curl http://localhost:8081/api/v1/cache/stats
```

Optimization:
- Increase `FLEXFLAG_EDGE_MAX_FLAGS`
- Tune cache TTL settings
- Scale horizontally with more replicas

#### 4. Load Balancer Issues

Check Nginx status:
```bash
docker-compose -f docker-compose.edge.yml logs nginx
```

### Debug Mode

Run edge server locally for debugging:

```bash
# Enable debug logging
FLEXFLAG_EDGE_ENVIRONMENT=development make run-edge
```

## Security

### Network Security

- Edge servers communicate with central hub over TLS
- API keys are hashed and stored securely
- CORS configured for browser requests

### API Key Management

- Use environment-specific API keys
- Rotate keys regularly
- Monitor key usage through metrics

### Firewall Rules

Recommended firewall configuration:

```bash
# Allow inbound to edge server
ufw allow 8081/tcp comment "FlexFlag Edge Server"

# Allow inbound to load balancer  
ufw allow 8082/tcp comment "FlexFlag Load Balancer"

# Allow outbound to central hub
ufw allow out 8080/tcp comment "FlexFlag Hub Connection"
ufw allow out 443/tcp comment "HTTPS Hub Connection"
```

## Production Deployment

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flexflag-edge
spec:
  replicas: 3
  selector:
    matchLabels:
      app: flexflag-edge
  template:
    metadata:
      labels:
        app: flexflag-edge
    spec:
      containers:
      - name: flexflag-edge
        image: flexflag-edge:latest
        env:
        - name: FLEXFLAG_EDGE_API_KEY
          valueFrom:
            secretKeyRef:
              name: flexflag-secrets
              key: edge-api-key
        ports:
        - containerPort: 8081
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
```

### Cloud Provider Specific

#### AWS
- Deploy using ECS Fargate or EKS
- Use Application Load Balancer
- Store secrets in AWS Secrets Manager

#### GCP  
- Deploy using Cloud Run or GKE
- Use Cloud Load Balancer
- Store secrets in Secret Manager

#### Azure
- Deploy using Container Instances or AKS  
- Use Application Gateway
- Store secrets in Key Vault

## Performance Optimization

### Tuning Parameters

```bash
# High-performance configuration
FLEXFLAG_EDGE_MAX_FLAGS=50000
FLEXFLAG_EDGE_MAX_API_KEYS=5000
FLEXFLAG_EDGE_CACHE_TTL=2h
FLEXFLAG_EDGE_BUFFER_SIZE=5000
FLEXFLAG_EDGE_CLEANUP_INTERVAL=1m
```

### Resource Limits

```yaml
# Docker Compose resource limits
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '1.0'
    reservations:
      memory: 256M
      cpus: '0.5'
```

### Benchmarking

```bash
# Run comprehensive performance tests
make edge-test

# Custom load test with Apache Bench
ab -n 10000 -c 100 \
   -H "X-API-Key: your-api-key" \
   -T "application/json" \
   -p test-payload.json \
   "http://localhost:8082/api/v1/evaluate"
```

## Support

For issues with edge server deployment:

1. Check the [troubleshooting section](#troubleshooting)
2. Review logs: `make edge-logs`
3. Verify configuration against this guide
4. Open an issue with deployment details and logs