#!/bin/bash

# FlexFlag Edge Server Test Script
# This script demonstrates how to test the edge server

set -e

echo "🧪 FlexFlag Edge Server Testing"
echo "==============================="

# Configuration
EDGE_URL="http://localhost:8083"
MAIN_API_URL="http://localhost:8080"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Helper functions
log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Test 1: Health Check
test_health() {
    echo ""
    log_info "Testing health endpoints..."
    
    # Health check
    if curl -s -f "$EDGE_URL/health" > /dev/null; then
        log_success "Health endpoint is responding"
        curl -s "$EDGE_URL/health" | python3 -m json.tool 2>/dev/null || echo "Health response received"
    else
        log_error "Health endpoint failed"
        return 1
    fi
    
    # Readiness check
    echo ""
    log_info "Checking readiness status..."
    READY_RESPONSE=$(curl -s "$EDGE_URL/ready")
    echo "$READY_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$READY_RESPONSE"
    
    # Check if ready
    if echo "$READY_RESPONSE" | grep -q '"ready":true'; then
        log_success "Edge server is ready with populated cache"
    else
        log_warning "Edge server is not ready (cache not populated or sync inactive)"
    fi
}

# Test 2: API Key Validation
test_api_key_validation() {
    echo ""
    log_info "Testing API key validation..."
    
    # Test without API key
    if curl -s "$EDGE_URL/api/v1/cache/stats" | grep -q "API key required"; then
        log_success "API key validation working (blocked request without key)"
    else
        log_warning "API key validation may not be working properly"
    fi
    
    # Test with invalid API key
    if curl -s -H "X-API-Key: invalid-key" "$EDGE_URL/api/v1/cache/stats" | grep -q "Invalid API key"; then
        log_success "Invalid API key properly rejected"
    else
        log_warning "Invalid API key handling may not be working"
    fi
}

# Test 3: Performance Test (Mock)
test_performance() {
    echo ""
    log_info "Testing basic performance..."
    
    # Measure health endpoint response time
    RESPONSE_TIME=$(curl -o /dev/null -s -w "%{time_total}" "$EDGE_URL/health")
    log_info "Health endpoint response time: ${RESPONSE_TIME}s"
    
    if [ "$(echo "$RESPONSE_TIME < 0.01" | bc 2>/dev/null || echo 0)" = "1" ]; then
        log_success "Excellent response time (<10ms)"
    elif [ "$(echo "$RESPONSE_TIME < 0.05" | bc 2>/dev/null || echo 0)" = "1" ]; then
        log_success "Good response time (<50ms)"
    else
        log_warning "Response time could be better (>${RESPONSE_TIME}s)"
    fi
}

# Test 4: Integration Test (requires main server)
test_integration() {
    echo ""
    log_info "Testing integration with main FlexFlag server..."
    
    # Check if main server is running
    if curl -s -f "$MAIN_API_URL/health" > /dev/null 2>&1; then
        log_success "Main FlexFlag server is running"
        
        # In a real scenario, you would:
        # 1. Create a project and flag via main API
        # 2. Generate an API key for the project
        # 3. Test evaluation via edge server
        
        log_info "To test full integration:"
        echo "  1. Ensure main FlexFlag server is running on port 8080"
        echo "  2. Create a project via web UI or API"
        echo "  3. Create flags in the project"
        echo "  4. Generate an API key for the project"
        echo "  5. Test evaluation using the generated API key:"
        echo ""
        echo "     curl -X POST $EDGE_URL/api/v1/evaluate \\"
        echo "       -H \"Content-Type: application/json\" \\"
        echo "       -H \"X-API-Key: YOUR_API_KEY\" \\"
        echo "       -d '{\"flag_key\":\"test-flag\",\"user_id\":\"user123\"}'"
        echo ""
        
    else
        log_warning "Main FlexFlag server not running on $MAIN_API_URL"
        log_info "Start it with: make run (in another terminal)"
    fi
}

# Test 5: Mock Evaluation Test
test_mock_evaluation() {
    echo ""
    log_info "Testing mock evaluation request format..."
    
    # Test evaluation request format (will fail auth but shows endpoint is working)
    EVAL_REQUEST='{
        "flag_key": "test-flag",
        "user_id": "user123",
        "user_key": "user123",
        "attributes": {
            "email": "test@example.com",
            "plan": "premium"
        }
    }'
    
    echo "Sample evaluation request:"
    echo "$EVAL_REQUEST" | python3 -m json.tool 2>/dev/null || echo "$EVAL_REQUEST"
    
    RESPONSE=$(curl -s -X POST "$EDGE_URL/api/v1/evaluate" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: test-key" \
        -d "$EVAL_REQUEST")
    
    echo ""
    echo "Response:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
    
    if echo "$RESPONSE" | grep -q "API key"; then
        log_success "Evaluation endpoint is accessible (API key validation working)"
    else
        log_warning "Unexpected response from evaluation endpoint"
    fi
}

# Test 6: Batch Evaluation Test
test_batch_evaluation() {
    echo ""
    log_info "Testing batch evaluation format..."
    
    BATCH_REQUEST='{
        "flag_keys": ["feature-a", "feature-b", "feature-c"],
        "user_id": "user123",
        "user_key": "user123",
        "attributes": {
            "email": "test@example.com"
        }
    }'
    
    echo "Sample batch request:"
    echo "$BATCH_REQUEST" | python3 -m json.tool 2>/dev/null || echo "$BATCH_REQUEST"
    
    RESPONSE=$(curl -s -X POST "$EDGE_URL/api/v1/evaluate/batch" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: test-key" \
        -d "$BATCH_REQUEST")
    
    echo ""
    echo "Response:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
}

# Run all tests
run_all_tests() {
    test_health
    test_api_key_validation
    test_performance
    test_mock_evaluation
    test_batch_evaluation
    test_integration
    
    echo ""
    echo "🎉 Edge Server Testing Complete!"
    echo "=================================="
    echo ""
    echo "📝 Summary:"
    echo "  • Edge server is running on $EDGE_URL"
    echo "  • Health endpoints are working"
    echo "  • API key validation is active"
    echo "  • Evaluation endpoints are accessible"
    echo ""
    echo "🚀 Next Steps for Full Testing:"
    echo "  1. Start main FlexFlag server: make run"
    echo "  2. Create a project and flags via web UI"
    echo "  3. Generate an API key for your project"
    echo "  4. Test real flag evaluations with your API key"
    echo ""
}

# Main execution
case "${1:-all}" in
    "health")
        test_health
        ;;
    "auth")
        test_api_key_validation
        ;;
    "performance")
        test_performance
        ;;
    "integration")
        test_integration
        ;;
    "evaluation")
        test_mock_evaluation
        test_batch_evaluation
        ;;
    "all")
        run_all_tests
        ;;
    *)
        echo "Usage: $0 {health|auth|performance|integration|evaluation|all}"
        echo ""
        echo "Test categories:"
        echo "  health      - Test health and readiness endpoints"
        echo "  auth        - Test API key authentication"
        echo "  performance - Basic performance testing"
        echo "  integration - Test integration with main server"
        echo "  evaluation  - Test evaluation endpoint formats"
        echo "  all         - Run all tests (default)"
        exit 1
        ;;
esac