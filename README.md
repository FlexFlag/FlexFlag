# FlexFlag - Open Source Feature Flag Management System

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

FlexFlag is a developer-first, GitOps-native feature flag management system that bridges the gap between expensive enterprise solutions and limited open source alternatives.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Installation

1. Clone the repository:
```bash
git clone https://github.com/flexflag/flexflag.git
cd flexflag
```

2. Install dependencies:
```bash
make deps
make install-tools
```

3. Start services with Docker Compose:
```bash
docker-compose up -d
```

4. Run database migrations:
```bash
make migrate-up
```

5. Start the server:
```bash
make run
```

The server will be available at `http://localhost:8080`

## 🏗️ Project Structure

```
flexflag/
├── cmd/                  # Application entrypoints
│   ├── server/          # API server
│   ├── cli/             # CLI tool
│   └── migrator/        # Database migrations
├── internal/            # Private application code
│   ├── api/            # HTTP handlers and routing
│   ├── core/           # Core business logic
│   ├── storage/        # Data access layer
│   ├── config/         # Configuration management
│   └── auth/           # Authentication/authorization
├── pkg/                # Public library code
│   ├── sdk/            # Go SDK for clients
│   ├── types/          # Shared types/models
│   └── utils/          # Utility functions
├── api/                # API definitions
├── deployments/        # Deployment configurations
├── migrations/         # Database migrations
├── docs/              # Documentation
└── test/              # Integration tests
```

## 🛠️ Development

### Running with hot reload:
```bash
make dev
```

### Running tests:
```bash
make test                 # Unit tests
make test-integration     # Integration tests
make bench               # Benchmarks
```

### Linting and formatting:
```bash
make lint                # Run linter
make fmt                 # Format code
```

### Building:
```bash
make build               # Build all binaries
```

## 🔧 Configuration

FlexFlag uses environment variables or a `config.yaml` file for configuration:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  environment: development

database:
  host: localhost
  port: 5432
  username: flexflag
  password: flexflag
  database: flexflag

redis:
  host: localhost
  port: 6379
  database: 0
  ttl: 300

auth:
  jwt_secret: your-secret-key
  token_expiry: 3600

logging:
  level: info
  format: json
```

Environment variables use the prefix `FLEXFLAG_` (e.g., `FLEXFLAG_SERVER_PORT=8080`)

## 📚 API Documentation

Once the server is running, API documentation is available at:
- OpenAPI Spec: `http://localhost:8080/api/docs`
- Health Check: `http://localhost:8080/health`

## 🐳 Docker Deployment

Build and run with Docker:
```bash
make docker-build
make docker-run
```

## 📦 CLI Usage

```bash
# Create a new flag
flexflag create my-feature-flag

# List all flags
flexflag list

# Toggle a flag
flexflag toggle my-feature-flag
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

FlexFlag is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## 🗺️ Roadmap

### Phase 1: Core Engine (Months 1-4)
- [x] Basic project structure
- [ ] Flag evaluation engine
- [ ] REST API
- [ ] PostgreSQL integration
- [ ] First SDK

### Phase 2: Developer Experience (Months 5-8)
- [ ] Web dashboard
- [ ] CLI tool enhancements
- [ ] Additional SDKs
- [ ] Documentation platform

### Phase 3: GitOps & Scale (Months 9-12)
- [ ] GitOps integration
- [ ] Performance optimization
- [ ] Advanced targeting
- [ ] Analytics & insights

### Phase 4: Community & Ecosystem (Months 13-18)
- [ ] Additional SDKs
- [ ] Integration ecosystem
- [ ] Enterprise features
- [ ] Community platform

## 💡 Key Features (Planned)

- **GitOps Native**: Feature flags as code with Git-based workflows
- **Performance First**: <10ms evaluation, edge computing ready
- **Developer Experience**: Intuitive UI, comprehensive documentation
- **Cost Effective**: Open source, self-hosted, unlimited flags
- **Customizable**: Extensible architecture, plugin support

## 🆚 Comparison

| Feature | FlexFlag | LaunchDarkly | Unleash | Flagsmith |
|---------|----------|--------------|---------|-----------|
| Open Source | ✅ | ❌ | ✅ | ✅ |
| Self-Hosted | ✅ | ❌ | ✅ | ✅ |
| GitOps Native | ✅ | ❌ | ❌ | ❌ |
| Performance | <10ms | <50ms | <30ms | <40ms |
| Pricing | Free/OSS | $70k+ | Free/Paid | Free/Paid |

## 📞 Support

- GitHub Issues: [github.com/flexflag/flexflag/issues](https://github.com/flexflag/flexflag/issues)
- Discord: [Join our community](https://discord.gg/flexflag)
- Documentation: [docs.flexflag.io](https://docs.flexflag.io)