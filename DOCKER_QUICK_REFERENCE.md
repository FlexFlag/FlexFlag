# FlexFlag Docker Quick Reference

## For Publishers (Maintainers)

### Initial Setup
```bash
# 1. Login to Docker Hub
docker login

# 2. Set your username
export DOCKER_USERNAME=your-dockerhub-username
export VERSION=1.0.0
```

### Build and Publish
```bash
# Build all images
make docker-build-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION

# Test locally
DOCKER_REGISTRY=$DOCKER_USERNAME/ VERSION=$VERSION docker-compose -f docker-compose.prod.yml up -d

# Push to Docker Hub
make docker-push-all DOCKER_USERNAME=$DOCKER_USERNAME VERSION=$VERSION
```

---

## For Users (Developers)

### One-Line Install
```bash
curl -sSL https://raw.githubusercontent.com/YOUR_ORG/flexflag/main/quick-start.sh | bash
```

### Manual Install
```bash
# 1. Get docker-compose file
curl -o docker-compose.prod.yml https://raw.githubusercontent.com/YOUR_ORG/flexflag/main/docker-compose.prod.yml
curl -o .env https://raw.githubusercontent.com/YOUR_ORG/flexflag/main/.env.example

# 2. Configure (edit .env file)
nano .env

# 3. Pull and run
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

### Access
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Docs: http://localhost:8080/swagger/index.html

### Default Login
```
Email: admin@flexflag.com
Password: admin123
```
⚠️ Change password after first login!

---

## Common Commands

### For Publishers
```bash
# Build specific image
docker build -f Dockerfile.api -t $DOCKER_USERNAME/flexflag-api:$VERSION .

# Tag as latest
docker tag $DOCKER_USERNAME/flexflag-api:$VERSION $DOCKER_USERNAME/flexflag-api:latest

# Push single image
docker push $DOCKER_USERNAME/flexflag-api:$VERSION
docker push $DOCKER_USERNAME/flexflag-api:latest
```

### For Users
```bash
# Start
docker-compose -f docker-compose.prod.yml up -d

# Stop
docker-compose -f docker-compose.prod.yml down

# Logs
docker-compose -f docker-compose.prod.yml logs -f

# Update
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d

# Backup database
docker exec flexflag-postgres pg_dump -U flexflag flexflag > backup.sql
```

---

## Makefile Commands

### Publisher Commands
```bash
make docker-build-all      # Build all images
make docker-push-all       # Push to Docker Hub
make docker-tag            # Tag for release
```

### User Commands
```bash
make docker-prod-up        # Start production
make docker-prod-down      # Stop production
make docker-logs           # View logs
make docker-clean          # Clean up
```

---

## Environment Variables

### Required
- `POSTGRES_PASSWORD` - Database password
- `JWT_SECRET` - JWT signing key
- `NEXT_PUBLIC_API_URL` - API URL

### Optional
- `GOOGLE_CLIENT_ID` - OAuth client ID
- `GOOGLE_CLIENT_SECRET` - OAuth secret
- `CORS_ALLOWED_ORIGINS` - Allowed origins
- `DEBUG` - Debug mode (false in production)

---

## Troubleshooting

### Database Connection
```bash
docker exec flexflag-postgres psql -U flexflag -d flexflag -c "SELECT 1;"
```

### API Health
```bash
curl http://localhost:8080/health
```

### View Logs
```bash
docker-compose logs postgres
docker-compose logs api
docker-compose logs ui
```

### Reset Everything
```bash
docker-compose -f docker-compose.prod.yml down -v
docker-compose -f docker-compose.prod.yml up -d
```

---

## Links

- **Full Deploy Guide**: [DOCKER_DEPLOY.md](DOCKER_DEPLOY.md)
- **Publish Guide**: [PUBLISH_GUIDE.md](PUBLISH_GUIDE.md)
- **Main README**: [README.md](README.md)
