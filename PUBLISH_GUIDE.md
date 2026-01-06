# FlexFlag - Docker Hub Publishing Guide

Complete step-by-step guide to publish FlexFlag to Docker Hub so developers can use `docker pull`.

## Prerequisites

1. **Docker Hub Account**
   - Create account at https://hub.docker.com
   - Note your username (e.g., `johndoe`)

2. **Local Setup**
   - Docker installed and running
   - Repository cloned locally
   - Git tags ready (optional but recommended)

---

## Step-by-Step Publishing

### 1. Login to Docker Hub

```bash
docker login
```

Enter your Docker Hub credentials when prompted.

### 2. Set Your Configuration

```bash
# Set your Docker Hub username
export DOCKER_USERNAME=your-dockerhub-username

# Set version (use git tag or semantic version)
export VERSION=1.0.0
# Or use git tag
export VERSION=$(git describe --tags --always)
```

### 3. Build All Images

```bash
# Build API, UI, and Migrator images
make docker-build-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
```

This will create:
- `your-username/flexflag-api:1.0.0` and `your-username/flexflag-api:latest`
- `your-username/flexflag-ui:1.0.0` and `your-username/flexflag-ui:latest`
- `your-username/flexflag-migrator:1.0.0` and `your-username/flexflag-migrator:latest`

### 4. Test Images Locally

```bash
# Update docker-compose.prod.yml with your username
sed -i '' "s/DOCKER_REGISTRY:-/DOCKER_REGISTRY:-$DOCKER_USERNAME\//g" docker-compose.prod.yml

# Test run
DOCKER_REGISTRY=$DOCKER_USERNAME/ VERSION=$VERSION docker-compose -f docker-compose.prod.yml up -d

# Check health
curl http://localhost:8080/health
curl http://localhost:3000

# View logs
docker-compose -f docker-compose.prod.yml logs

# Stop
docker-compose -f docker-compose.prod.yml down
```

### 5. Push to Docker Hub

```bash
# Push all images
make docker-push-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
```

You should see output like:
```
✅ All images pushed successfully!
📦 Published images:
   - your-username/flexflag-api:1.0.0
   - your-username/flexflag-ui:1.0.0
   - your-username/flexflag-migrator:1.0.0
```

### 6. Verify on Docker Hub

1. Visit https://hub.docker.com/u/your-username
2. Check that all three images are listed
3. Verify tags (both `1.0.0` and `latest`)

### 7. Test Pull

```bash
# On a different machine or clean environment
docker pull your-username/flexflag-api:latest
docker pull your-username/flexflag-ui:latest
docker pull your-username/flexflag-migrator:latest
```

---

## Creating Docker Hub Repositories

### Option 1: Auto-Create (Recommended)

Images are automatically created when you first push. No manual setup needed!

### Option 2: Manual Setup (for better descriptions)

1. Go to https://hub.docker.com
2. Click "Create Repository"
3. Create three repositories:
   - `flexflag-api`
   - `flexflag-ui`
   - `flexflag-migrator`
4. Add descriptions:

**flexflag-api:**
```
FlexFlag API Server - High-performance feature flag management system

Fast, scalable feature flag service with <10ms evaluation times.
Includes user management, OAuth support, and real-time updates.

Features:
- REST API with Swagger docs
- PostgreSQL + Redis backend
- JWT authentication
- Google OAuth support
- Multi-project support
- Audit logging

Quick Start:
docker pull your-username/flexflag-api:latest

Documentation: https://github.com/your-org/flexflag
```

**flexflag-ui:**
```
FlexFlag Frontend UI - Modern feature flag management dashboard

Beautiful Material-UI dashboard for managing feature flags.

Features:
- Real-time flag management
- User administration
- Project management
- OAuth login
- Dark mode support

Quick Start:
docker pull your-username/flexflag-ui:latest

Documentation: https://github.com/your-org/flexflag
```

**flexflag-migrator:**
```
FlexFlag Database Migrator - Automated schema management

Handles PostgreSQL database migrations for FlexFlag.
Runs automatically in docker-compose setup.

Quick Start:
docker pull your-username/flexflag-migrator:latest

Documentation: https://github.com/your-org/flexflag
```

---

## Update README on Docker Hub

Add this content to each repository's README on Docker Hub:

### For flexflag-api:

```markdown
# FlexFlag API Server

High-performance feature flag management system with <10ms evaluation times.

## Quick Start

```bash
docker pull your-username/flexflag-api:latest
docker run -d \
  -p 8080:8080 \
  -e FLEXFLAG_DATABASE_HOST=postgres \
  -e FLEXFLAG_DATABASE_PORT=5432 \
  -e FLEXFLAG_DATABASE_USERNAME=flexflag \
  -e FLEXFLAG_DATABASE_PASSWORD=flexflag \
  -e FLEXFLAG_JWT_SECRET=your-secret-key \
  your-username/flexflag-api:latest
```

## Full Stack Setup

Use docker-compose for complete setup:

```bash
curl -o docker-compose.yml https://raw.githubusercontent.com/your-org/flexflag/main/docker-compose.prod.yml
curl -o .env https://raw.githubusercontent.com/your-org/flexflag/main/.env.example
docker-compose up -d
```

## Environment Variables

- `FLEXFLAG_DATABASE_HOST` - PostgreSQL host
- `FLEXFLAG_DATABASE_PORT` - PostgreSQL port (default: 5432)
- `FLEXFLAG_JWT_SECRET` - JWT signing secret (required)
- `FLEXFLAG_REDIS_HOST` - Redis host (optional)

See full documentation: https://github.com/your-org/flexflag

## Tags

- `latest` - Latest stable release
- `1.0.0` - Specific version
- `edge` - Development builds (if applicable)

## Links

- GitHub: https://github.com/your-org/flexflag
- Documentation: https://docs.flexflag.com
- Issues: https://github.com/your-org/flexflag/issues
```

---

## Automated Publishing with GitHub Actions

Create `.github/workflows/docker-publish.yml`:

```yaml
name: Publish Docker Images

on:
  push:
    tags:
      - 'v*'
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Extract version
        id: meta
        run: echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Build and push API
        uses: docker/build-push-action@v4
        with:
          context: .
          file: Dockerfile.api
          push: true
          tags: |
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-api:${{ steps.meta.outputs.version }}
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-api:latest

      - name: Build and push UI
        uses: docker/build-push-action@v4
        with:
          context: ./ui
          file: ./ui/Dockerfile
          push: true
          tags: |
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-ui:${{ steps.meta.outputs.version }}
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-ui:latest

      - name: Build and push Migrator
        uses: docker/build-push-action@v4
        with:
          context: .
          file: Dockerfile.migrator
          push: true
          tags: |
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-migrator:${{ steps.meta.outputs.version }}
            ${{ secrets.DOCKERHUB_USERNAME }}/flexflag-migrator:latest
```

**Setup GitHub Secrets:**
1. Go to Repository Settings → Secrets and variables → Actions
2. Add secrets:
   - `DOCKERHUB_USERNAME` - Your Docker Hub username
   - `DOCKERHUB_TOKEN` - Docker Hub access token (create at https://hub.docker.com/settings/security)

---

## Version Management

### Semantic Versioning

Follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Tagging Releases

```bash
# Create git tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Build and push with version
export VERSION=1.0.0
make docker-build-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
make docker-push-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
```

### Multiple Tags

Always tag with both version and `latest`:

```bash
# Specific version
docker tag flexflag-api:1.0.0 your-username/flexflag-api:1.0.0

# Latest
docker tag flexflag-api:1.0.0 your-username/flexflag-api:latest

# Push both
docker push your-username/flexflag-api:1.0.0
docker push your-username/flexflag-api:latest
```

---

## User Documentation

Update your main README.md with Docker Hub instructions:

```markdown
## Quick Start with Docker

### One-Line Install

```bash
curl -sSL https://raw.githubusercontent.com/your-org/flexflag/main/quick-start.sh | bash
```

### Manual Docker Setup

```bash
# Pull images
docker pull your-username/flexflag-api:latest
docker pull your-username/flexflag-ui:latest
docker pull your-username/flexflag-migrator:latest

# Get docker-compose file
curl -o docker-compose.yml https://raw.githubusercontent.com/your-org/flexflag/main/docker-compose.prod.yml

# Start FlexFlag
docker-compose up -d
```

Access at:
- Frontend: http://localhost:3000
- API: http://localhost:8080
- API Docs: http://localhost:8080/swagger/index.html
```

---

## Maintenance

### Update Images

```bash
# Rebuild and push latest
export DOCKER_USERNAME=your-username
export VERSION=1.0.1

make docker-build-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
make docker-push-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
```

### Delete Old Images

On Docker Hub:
1. Go to repository
2. Click "Tags"
3. Select old versions
4. Click "Delete"

Or use API:
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  https://hub.docker.com/v2/repositories/your-username/flexflag-api/tags/old-version/
```

---

## Troubleshooting

### Push Failed

```bash
# Re-login
docker logout
docker login

# Verify credentials
docker info | grep Username
```

### Image Too Large

```bash
# Check image size
docker images | grep flexflag

# Optimize Dockerfile
- Use multi-stage builds (already done)
- Minimize layers
- Use .dockerignore
```

### Rate Limiting

Docker Hub has pull rate limits:
- Anonymous: 100 pulls / 6 hours
- Free account: 200 pulls / 6 hours
- Pro account: Unlimited

Consider:
- Caching in CI/CD
- Using Docker Hub Pro
- Self-hosting registry

---

## Best Practices

1. **Always tag with version** - Don't rely only on `latest`
2. **Test before pushing** - Run full integration tests
3. **Use multi-stage builds** - Keep images small
4. **Security scanning** - Use `docker scan` or Snyk
5. **Automate with CI/CD** - GitHub Actions for consistency
6. **Document changes** - Update CHANGELOG.md
7. **Sign images** - Use Docker Content Trust (optional)

---

## Security

### Scan Images

```bash
# Docker scan (requires login)
docker scan your-username/flexflag-api:latest

# Trivy
trivy image your-username/flexflag-api:latest

# Snyk
snyk container test your-username/flexflag-api:latest
```

### Sign Images (Optional)

```bash
# Enable Docker Content Trust
export DOCKER_CONTENT_TRUST=1

# Sign and push
docker push your-username/flexflag-api:1.0.0
```

---

## Support

- **Issues**: https://github.com/your-org/flexflag/issues
- **Discussions**: https://github.com/your-org/flexflag/discussions
- **Docker Hub**: https://hub.docker.com/u/your-username

---

## Summary Checklist

- [ ] Create Docker Hub account
- [ ] Login locally: `docker login`
- [ ] Set DOCKER_USERNAME environment variable
- [ ] Build images: `make docker-build-all`
- [ ] Test images locally
- [ ] Push to Docker Hub: `make docker-push-all`
- [ ] Verify on hub.docker.com
- [ ] Update repository READMEs on Docker Hub
- [ ] Test pull from different machine
- [ ] Set up GitHub Actions (optional)
- [ ] Update main README with Docker instructions
- [ ] Create git release tag
- [ ] Announce release!

**You're ready to publish! 🚀**
