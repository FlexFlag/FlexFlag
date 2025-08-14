#!/bin/bash

# FlexFlag Edge Server Deployment Script
set -e

echo "🚀 FlexFlag Edge Server Deployment"
echo "=================================="

# Configuration
EDGE_API_KEY=${FLEXFLAG_EDGE_API_KEY:-""}
ENVIRONMENT=${ENVIRONMENT:-"production"}
REPLICAS=${EDGE_REPLICAS:-"1"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    if [ -z "$EDGE_API_KEY" ]; then
        log_warning "FLEXFLAG_EDGE_API_KEY not set, using default"
        EDGE_API_KEY="your-api-key-here"
    fi
    
    log_success "Prerequisites check passed"
}

# Build edge server
build_edge_server() {
    log_info "Building FlexFlag Edge Server..."
    
    # Build Go binary first to check for compilation errors
    if ! go build -o edge-server ./cmd/edge-server; then
        log_error "Failed to build edge server binary"
        exit 1
    fi
    rm -f edge-server
    
    # Build Docker image
    docker-compose -f docker-compose.edge.yml build flexflag-edge
    
    log_success "Edge server built successfully"
}

# Deploy edge infrastructure
deploy_edge() {
    log_info "Deploying edge infrastructure..."
    
    # Export environment variables for docker-compose
    export FLEXFLAG_EDGE_API_KEY=$EDGE_API_KEY
    
    # Stop existing services
    docker-compose -f docker-compose.edge.yml down
    
    # Start all services
    docker-compose -f docker-compose.edge.yml up -d
    
    log_success "Edge infrastructure deployed"
}

# Scale edge servers
scale_edge_servers() {
    if [ "$REPLICAS" -gt 1 ]; then
        log_info "Scaling edge servers to $REPLICAS replicas..."
        docker-compose -f docker-compose.edge.yml up -d --scale flexflag-edge=$REPLICAS
        log_success "Scaled to $REPLICAS edge server replicas"
    fi
}

# Health check
health_check() {
    log_info "Performing health checks..."
    
    # Wait for services to start
    sleep 10
    
    # Check main API health
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "Main API is healthy"
    else
        log_warning "Main API health check failed"
    fi
    
    # Check edge server health
    if curl -f http://localhost:8081/health &> /dev/null; then
        log_success "Edge server is healthy"
    else
        log_warning "Edge server health check failed"
    fi
    
    # Check load balancer
    if curl -f http://localhost:8082/health &> /dev/null; then
        log_success "Load balancer is healthy"
    else
        log_warning "Load balancer health check failed"
    fi
}

# Show deployment status
show_status() {
    log_info "Deployment Status:"
    echo "=================="
    
    echo "🌐 Services:"
    echo "  Main API:        http://localhost:8080"
    echo "  Edge Server:     http://localhost:8081"
    echo "  Load Balancer:   http://localhost:8082"
    echo "  Database:        localhost:5433"
    echo "  Redis:           localhost:6379"
    echo ""
    
    echo "📊 Health Endpoints:"
    echo "  Main API:        http://localhost:8080/health"
    echo "  Edge Server:     http://localhost:8081/health"
    echo "  Load Balancer:   http://localhost:8082/health"
    echo ""
    
    echo "🔧 Management:"
    echo "  Edge Stats:      http://localhost:8081/api/v1/cache/stats"
    echo "  Nginx Status:    http://localhost:8082:8080/nginx_status"
    echo ""
    
    echo "📦 Running Containers:"
    docker-compose -f docker-compose.edge.yml ps
}

# Performance test
performance_test() {
    if command -v ab &> /dev/null; then
        log_info "Running basic performance test..."
        
        # Create a test API key if needed
        echo '{"name": "test-key", "environment_id": "production", "permissions": ["read"]}' > /tmp/api-key-request.json
        
        # Test edge evaluation performance
        echo ""
        echo "Testing edge server evaluation performance:"
        ab -n 1000 -c 10 -H "X-API-Key: your-api-key-here" \
           -T "application/json" \
           -p /dev/null \
           "http://localhost:8081/api/v1/evaluate"
    else
        log_warning "Apache Bench (ab) not available, skipping performance test"
    fi
}

# Main deployment flow
main() {
    case ${1:-deploy} in
        "build")
            check_prerequisites
            build_edge_server
            ;;
        "deploy")
            check_prerequisites
            build_edge_server
            deploy_edge
            scale_edge_servers
            health_check
            show_status
            ;;
        "scale")
            scale_edge_servers
            health_check
            show_status
            ;;
        "status")
            show_status
            ;;
        "test")
            performance_test
            ;;
        "stop")
            log_info "Stopping edge infrastructure..."
            docker-compose -f docker-compose.edge.yml down
            log_success "Edge infrastructure stopped"
            ;;
        "logs")
            docker-compose -f docker-compose.edge.yml logs -f ${2:-flexflag-edge}
            ;;
        *)
            echo "Usage: $0 {build|deploy|scale|status|test|stop|logs}"
            echo ""
            echo "Commands:"
            echo "  build   - Build edge server binary and Docker image"
            echo "  deploy  - Full deployment of edge infrastructure"
            echo "  scale   - Scale edge servers (set EDGE_REPLICAS)"
            echo "  status  - Show deployment status"
            echo "  test    - Run performance tests"
            echo "  stop    - Stop all edge services"
            echo "  logs    - Show logs (optionally specify service)"
            echo ""
            echo "Environment Variables:"
            echo "  FLEXFLAG_EDGE_API_KEY - API key for edge server authentication"
            echo "  ENVIRONMENT          - Deployment environment (default: production)"
            echo "  EDGE_REPLICAS        - Number of edge server replicas (default: 1)"
            exit 1
            ;;
    esac
}

main "$@"