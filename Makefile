.PHONY: help build run test clean docker-build docker-run lint fmt migrate-up migrate-down

APP_NAME=flexflag
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building ${APP_NAME}..."
	@go build ${LDFLAGS} -o bin/server cmd/server/main.go
	@go build ${LDFLAGS} -o bin/cli cmd/cli/main.go
	@go build ${LDFLAGS} -o bin/migrator cmd/migrator/main.go

run: ## Run the server
	@go run cmd/server/main.go

run-cli: ## Run the CLI
	@go run cmd/cli/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

coverage: ## Run tests and display coverage percentage
	@echo "Running tests and calculating coverage..."
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1 || true
	@go tool cover -func=coverage.out | tail -1
	@echo "📊 For detailed coverage report, run: make test"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./test/...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t ${APP_NAME}:${VERSION} -t ${APP_NAME}:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

migrate-up: ## Run database migrations up
	@go run cmd/migrator/main.go -database-url="postgres://flexflag:flexflag@localhost:5433/flexflag?sslmode=disable" -direction=up

migrate-down: ## Run database migrations down
	@go run cmd/migrator/main.go -database-url="postgres://flexflag:flexflag@localhost:5433/flexflag?sslmode=disable" -direction=down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_flags_table)
	@migrate create -ext sql -dir migrations -seq $(NAME)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/cosmtrek/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

dev: ## Run server with hot reload
	@air

generate: ## Generate code
	@echo "Generating code..."
	@go generate ./...

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if [ ! -f "$(shell go env GOPATH)/bin/swag" ]; then \
		echo "Installing swag tool..."; \
		go install github.com/swaggo/swag/cmd/swag@v1.16.3; \
	fi
	@$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o api/ -d ./
	@echo "Fixing compatibility issues..."
	@sed -i '' '/LeftDelim:/d' api/docs.go
	@sed -i '' '/RightDelim:/d' api/docs.go
	@echo "✅ Swagger documentation generated successfully!"
	@echo "📖 View at: http://localhost:8080/swagger/index.html"

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# Edge Server Commands
build-edge: ## Build edge server binary
	@echo "Building edge server..."
	@go build ${LDFLAGS} -o bin/edge-server cmd/edge-server/main.go

run-edge: ## Run edge server locally
	@go run cmd/edge-server/main.go

docker-build-edge: ## Build edge server Docker image
	@echo "Building edge server Docker image..."
	@docker build -f cmd/edge-server/Dockerfile -t flexflag-edge:${VERSION} -t flexflag-edge:latest .

edge-deploy: ## Deploy edge infrastructure
	@echo "Deploying edge infrastructure..."
	@./deployments/deploy-edge.sh deploy

edge-build: ## Build edge server only
	@echo "Building edge server..."
	@./deployments/deploy-edge.sh build

edge-scale: ## Scale edge servers (set EDGE_REPLICAS)
	@echo "Scaling edge servers..."
	@./deployments/deploy-edge.sh scale

edge-status: ## Show edge deployment status
	@./deployments/deploy-edge.sh status

edge-test: ## Run edge server performance tests
	@./deployments/deploy-edge.sh test

edge-stop: ## Stop edge infrastructure
	@./deployments/deploy-edge.sh stop

edge-logs: ## Show edge server logs
	@./deployments/deploy-edge.sh logs