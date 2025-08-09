# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go Backend
- `make run` - Start the Go API server on port 8080
- `make dev` - Run server with hot reload using Air
- `make test` - Run unit tests with coverage report
- `make bench` - Run performance benchmarks
- `make lint` - Run golangci-lint
- `make build` - Build server, CLI, and migrator binaries

### Next.js Frontend (UI)
- `cd ui && npm run dev` - Start Next.js dev server on port 3000
- `cd ui && npm run build` - Build production frontend
- `cd ui && npm run lint` - Run ESLint

### Database & Infrastructure
- `docker-compose up -d` - Start PostgreSQL (port 5433) and Redis services
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback migrations
- `make migrate-create NAME=migration_name` - Create new migration

### Performance Testing
- `./benchmark_evaluation.sh` - Run evaluation performance benchmarks
- `./concurrent_benchmark.sh` - Run concurrent load tests

## Architecture Overview

FlexFlag is a high-performance feature flag management system with a Go backend and Next.js frontend, optimized for <10ms flag evaluation.

### Backend Architecture (Go)
- **Clean Architecture**: `cmd/`, `internal/`, `pkg/` structure
- **API Layer**: Gin HTTP handlers in `internal/api/handlers/`
- **Storage Layer**: Repository pattern with PostgreSQL (`internal/storage/`)
- **Core Types**: Shared models in `pkg/types/`

### Multi-Tier Evaluation System
1. **Standard Handler**: Basic flag evaluation (~2-6ms)
2. **Optimized Handler**: With connection pooling and in-memory caching
3. **Ultra-Fast Handler**: Pre-loading, response caching, unsafe optimizations (<1ms)

The Ultra-Fast handler preloads all flags into memory (`PrecomputedFlag` structs) and caches serialized responses with MD5 request hashing for sub-millisecond evaluation.

### Frontend Architecture (Next.js 14)
- **App Router**: Modern Next.js routing in `ui/app/`
- **Material-UI v6**: Component library with custom theming
- **TypeScript**: Full type safety with shared types
- **API Client**: Centralized HTTP client in `ui/lib/api.ts`

### Data Flow
1. Flags stored in PostgreSQL with JSONB support
2. Multiple evaluation endpoints: `/evaluate`, `/evaluate/fast`, `/evaluate/ultra`
3. Frontend communicates via REST API with environment-based flag management
4. Redis used for caching in optimized handlers

## Configuration

### Environment Variables
All config uses `FLEXFLAG_` prefix:
- `FLEXFLAG_DATABASE_PORT=5433` (Note: Uses 5433 to avoid conflicts)
- `FLEXFLAG_SERVER_PORT=8080`
- Configuration handled via Viper with YAML support

### Database Schema
- Flags table with JSONB columns for variations/targeting
- Migration system using golang-migrate
- Connection pooling configured for performance

## Performance Characteristics

### Current Benchmarks
- Ultra-fast evaluation: 95th percentile ~0.1ms (317x improvement over standard)
- Standard evaluation: ~2-6ms
- Batch evaluation: ~0.5-1ms per flag
- Target: <10ms (achieved)

### Optimization Techniques
- Pre-computed flag structures in memory
- Response-level caching with TTL
- Unsafe pointer operations for maximum speed
- Connection pooling with configurable min/max

## Key Implementation Notes

### Flag Types
- Boolean, String, Number, JSON supported
- Variations and targeting rules via JSONB
- Environment-based flag isolation (production, staging, development)

### API Endpoints Structure
- REST API at `/api/v1/`
- CRUD operations: `POST /flags`, `GET /flags`, `PUT /flags/:key`, `DELETE /flags/:key`
- Evaluation: `/evaluate`, `/evaluate/batch`, `/evaluate/ultra`
- Toggle endpoint: `POST /flags/:key/toggle`

### Frontend Features
- Real-time dashboard with performance metrics
- Flag CRUD with confirmation dialogs
- Environment switching (production/staging/development)
- Edit functionality requires both `key` and `type` fields in requests

### Common Patterns
- Repository pattern for data access
- Handler-based HTTP routing
- Environment-based configuration via query params
- JSON response format with nested `flags` array for list operations

### Performance Monitoring
- Built-in benchmarking scripts
- Cache hit/miss statistics
- Ultra-fast handler provides preload status and response cache metrics
- Development focused on <10ms evaluation times
- follow convention git commit types like fix, feature, tests,docs, refactor etc
- Do not include claude in commit messages
- do not include claude in PR creation
- commits should be atomic and be made in logical sequence