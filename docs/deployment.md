---
title: Deployment
nav_order: 3
---

# Deployment
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Vercel + Railway (Recommended)

The simplest production setup: Next.js frontend on Vercel, Go backend on Railway.

### 1. Deploy the backend (Railway)

1. Go to [railway.app](https://railway.app) → **New Project** → **Deploy from GitHub**
2. Add a **PostgreSQL** plugin and a **Redis** plugin to the project
3. In your backend service → **Variables**, add:

```
FLEXFLAG_DATABASE_HOST=${{Postgres.PGHOST}}
FLEXFLAG_DATABASE_PORT=${{Postgres.PGPORT}}
FLEXFLAG_DATABASE_USERNAME=${{Postgres.PGUSER}}
FLEXFLAG_DATABASE_PASSWORD=${{Postgres.PGPASSWORD}}
FLEXFLAG_DATABASE_DATABASE=${{Postgres.PGDATABASE}}
FLEXFLAG_DATABASE_SSL_MODE=require

FLEXFLAG_REDIS_HOST=${{Redis.REDISHOST}}
FLEXFLAG_REDIS_PORT=${{Redis.REDISPORT}}
FLEXFLAG_REDIS_PASSWORD=${{Redis.REDISPASSWORD}}

FLEXFLAG_JWT_SECRET=<generate with: openssl rand -base64 32>
FLEXFLAG_SERVER_ENVIRONMENT=production
FLEXFLAG_CORS_ALLOWED_ORIGINS=https://your-app.vercel.app
```

4. Set **Build Command:** `go build -o out ./cmd/server`
5. Set **Start Command:** `/app/out`
6. In **Settings → Networking**, click **Generate Domain** — note this URL

### 2. Run migrations (Railway)

Add a second Railway service from the same repo:

- **Build Command:** `go build -o out ./cmd/migrator`
- **Start Command:** `/app/out up`
- Same DB env vars as the backend

It runs once and exits — that's expected.

### 3. Deploy the frontend (Vercel)

1. Go to [vercel.com](https://vercel.com) → **New Project** → import your repo
2. Set **Root Directory** to `FlexFlag/ui`
3. Add environment variable:
   - `INTERNAL_API_URL` = your Railway backend URL
4. Deploy

---

## Docker Compose (Self-Hosted)

For running everything on a single server:

```bash
git clone https://github.com/FlexFlag/FlexFlag.git
cd FlexFlag
cp .env.example .env
# Edit .env with your values
docker-compose -f docker-compose.prod.yml up -d
```

Services started:
- PostgreSQL on port 5433
- Redis on port 6379
- API server on port 8080
- Next.js UI on port 3000
- Nginx reverse proxy on port 80

---

## Environment Variables

### Backend (Go API)

| Variable | Default | Description |
|---|---|---|
| `FLEXFLAG_SERVER_PORT` | `8080` | API server port |
| `FLEXFLAG_SERVER_ENVIRONMENT` | `development` | `production` / `staging` / `development` |
| `FLEXFLAG_DATABASE_HOST` | `localhost` | PostgreSQL host |
| `FLEXFLAG_DATABASE_PORT` | `5433` | PostgreSQL port |
| `FLEXFLAG_DATABASE_USERNAME` | `flexflag` | PostgreSQL user |
| `FLEXFLAG_DATABASE_PASSWORD` | — | PostgreSQL password |
| `FLEXFLAG_DATABASE_DATABASE` | `flexflag` | PostgreSQL database name |
| `FLEXFLAG_DATABASE_SSL_MODE` | `disable` | `disable` / `require` |
| `FLEXFLAG_REDIS_HOST` | `localhost` | Redis host |
| `FLEXFLAG_REDIS_PORT` | `6379` | Redis port |
| `FLEXFLAG_REDIS_PASSWORD` | — | Redis password |
| `FLEXFLAG_JWT_SECRET` | — | JWT signing secret (required) |
| `FLEXFLAG_JWT_EXPIRY` | `24h` | Token expiry duration |
| `FLEXFLAG_CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated allowed origins |

### Frontend (Next.js)

| Variable | Description |
|---|---|
| `INTERNAL_API_URL` | Backend URL used by Next.js rewrites (e.g. `https://api.railway.app`) |

### Edge Server

| Variable | Default | Description |
|---|---|---|
| `FLEXFLAG_EDGE_PORT` | `8081` | Edge server port |
| `FLEXFLAG_HUB_URL` | `http://localhost:8080` | Main API server URL |
| `FLEXFLAG_EDGE_API_KEY` | — | API key for edge authentication |
| `FLEXFLAG_EDGE_SYNC_TYPE` | `sse` | `sse` or `websocket` |
| `FLEXFLAG_EDGE_MAX_FLAGS` | `10000` | Max flags to cache |
| `FLEXFLAG_EDGE_CACHE_TTL` | `1h` | Cache TTL |
