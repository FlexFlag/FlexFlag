---
title: Home
layout: home
nav_order: 1
---

# FlexFlag
{: .fs-9 }

High-performance, developer-first feature flag management with sub-millisecond evaluation and distributed edge servers.
{: .fs-6 .fw-300 }

[Live Demo](https://flex-flag-zyen-idpjh03va-reez-personals-projects.vercel.app/){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[Get Started](#quick-start){: .btn .fs-5 .mb-4 .mb-md-0 .mr-2 }
[GitHub](https://github.com/FlexFlag/FlexFlag){: .btn .fs-5 .mb-4 .mb-md-0 }

---

> **Demo credentials** — Email: `admin@example.com` · Password: `admin123`

---

## What is FlexFlag?

FlexFlag is a self-hosted feature flag system built for teams that need speed and control. It ships with:

- A **Go API server** with three evaluation tiers (standard, optimized, ultra-fast)
- A **Next.js dashboard** for managing flags, rollouts, segments, and A/B experiments
- **Distributed edge servers** that sync in real-time via SSE/WebSocket for <1ms evaluation
- **Multi-project** support with per-environment flag isolation

---

## Key Features

| Feature | Description |
|---|---|
| **Ultra-Fast Evaluation** | <1ms flag evaluation via pre-computed in-memory structures |
| **Edge Servers** | Distributed evaluation nodes with real-time flag sync |
| **Advanced Targeting** | User segments, percentage rollouts, custom attribute rules |
| **A/B Experiments** | Variant flags with built-in experimentation support |
| **Approval Workflows** | Flag change approvals with reviewer assignment |
| **Remote Config** | String/number/JSON config values alongside boolean flags |
| **Multi-Project** | Project isolation with production/staging/development environments |
| **API Keys** | Scoped API keys for SDK and edge server authentication |
| **Audit Logs** | Full history of flag changes with actor tracking |
| **Analytics** | Real-time evaluation counters and performance metrics |

---

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
git clone https://github.com/FlexFlag/FlexFlag.git
cd FlexFlag
cp .env.example .env
./setup-dev.sh
```

After setup: [http://localhost:3000](http://localhost:3000) · Login: `admin@example.com` / `admin123`

### Option 2: Manual

**Prerequisites:** Go 1.24+, Node.js 18+, Docker

```bash
# Start PostgreSQL + Redis
docker-compose up -d

# Run migrations
go run cmd/migrator/main.go up

# Start API server
go run cmd/server/main.go

# Start UI (new terminal)
cd ui && npm install && npm run dev
```

---

## Evaluation Performance

| Tier | Endpoint | Latency | How it works |
|---|---|---|---|
| Standard | `/evaluate` | ~2–6ms | Direct PostgreSQL query |
| Optimized | `/evaluate/fast` | ~0.5–1ms | Redis-cached results |
| Ultra-Fast | `/evaluate/ultra` | ~0.1ms | Pre-computed in-memory + response cache |
| Edge | `edge-server/evaluate` | <1ms | Local in-memory, synced via SSE |

---

## Architecture

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   Next.js UI     │     │  Edge Server A   │     │  Edge Server B   │
│  (Vercel/3000)   │     │   (Port 8081)    │     │   (Port 8082)    │
└────────┬─────────┘     └────────┬─────────┘     └────────┬─────────┘
         │                        │   SSE / WebSocket       │
         └────────────────────────┼─────────────────────────┘
                                  ▼
                    ┌─────────────────────────┐
                    │   FlexFlag API Server   │
                    │      (Port 8080)        │
                    │  REST · SSE · WS · JWT  │
                    └────────────┬────────────┘
                                 │
               ┌─────────────────┴──────────────────┐
               ▼                                    ▼
      ┌─────────────────┐                ┌─────────────────┐
      │   PostgreSQL    │                │     Redis       │
      │  (flags, users, │                │  (cache, fast   │
      │   audit logs)   │                │   eval layer)   │
      └─────────────────┘                └─────────────────┘
```

---

## API Overview

### Authentication
```bash
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/auth/profile
```

### Feature Flags
```bash
GET    /api/v1/flags
POST   /api/v1/flags
PUT    /api/v1/flags/:key
DELETE /api/v1/flags/:key
POST   /api/v1/flags/:key/toggle
```

### Evaluation
```bash
POST /api/v1/evaluate              # Standard
POST /api/v1/evaluate/batch        # Batch
POST /api/v1/evaluate/fast         # Optimized
POST /api/v1/evaluate/fast/batch   # Optimized batch
POST /api/v1/evaluate/ultra        # Ultra-fast
```

### Projects & Environments
```bash
GET  /api/v1/projects
POST /api/v1/projects
GET  /api/v1/projects/:slug/environments
POST /api/v1/projects/:slug/environments
```

### Edge Servers
```bash
GET /api/v1/edge/sync      # Bulk sync (REST)
GET /api/v1/edge/sync/ws   # WebSocket sync
GET /api/v1/edge/sync/sse  # SSE sync
GET /api/v1/edge/servers   # Connected servers
```

Full interactive API docs available at `/swagger` on any running instance.

---

## Deployment

FlexFlag is deployed in two parts:

| Part | Platform | Notes |
|---|---|---|
| **Frontend** (Next.js) | Vercel | Set `INTERNAL_API_URL` env var |
| **Backend** (Go API) | Railway / Render / Fly.io | Set DB + Redis env vars |
| **Database** | Railway Postgres plugin | Auto-linked via `${{Postgres.PGHOST}}` etc. |
| **Cache** | Railway Redis plugin | Auto-linked via `${{Redis.REDISHOST}}` etc. |

See the [Deployment Guide](./deployment) for full instructions.

---

## SDKs

- [JavaScript / TypeScript SDK](./sdks/javascript/)
- [Python SDK](./sdks/python/)

---

## Community

- [Discord](https://discord.gg/fpewTJyx9S)
- [GitHub Issues](https://github.com/FlexFlag/FlexFlag/issues)
- [npm: flexflag-client](https://www.npmjs.com/package/flexflag-client)
