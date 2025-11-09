#!/bin/bash

# =============================================================================
# Custom Fields System Performance Test Script
# =============================================================================
# Tests performance and load handling of custom fields API endpoints
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Service URLs
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"
BOARD_SERVICE_URL="${BOARD_SERVICE_URL:-http://localhost:8000}"

# Performance test configuration
WARMUP_ITERATIONS=10
PERFORMANCE_ITERATIONS=100
CONCURRENT_REQUESTS_LIGHT=10
CONCURRENT_REQUESTS_MEDIUM=50
CONCURRENT_REQUESTS_HEAVY=100
SUSTAINED_TEST_DURATION=30  # seconds

# Result arrays
declare -a response_times

# =============================================================================
# Helper Functions
# =============================================================================

print_header() {
    echo ""
    echo -e "${CYAN}=================================================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}=================================================================${NC}"
    echo ""
}

print_section() {
    echo ""
    echo -e "${BLUE}>>> $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_metric() {
    echo -e "${MAGENTA}  $1${NC}"
}

# Calculate statistics from response times
calculate_stats() {
    local times=("$@")
    local count=${#times[@]}

    if [ $count -eq 0 ]; then
        echo "0 0 0 0"
        return
    fi

    # Sort array
    IFS=$'\n' sorted=($(sort -n <<<"${times[*]}"))
    unset IFS

    # Min and Max
    local min=${sorted[0]}
    local max=${sorted[$((count-1))]}

    # Average
    local sum=0
    for time in "${times[@]}"; do
        sum=$((sum + time))
    done
    local avg=$((sum / count))

    # Median
    local median
    if [ $((count % 2)) -eq 0 ]; then
        local mid1=${sorted[$((count/2 - 1))]}
        local mid2=${sorted[$((count/2))]}
        median=$(( (mid1 + mid2) / 2 ))
    else
        median=${sorted[$((count/2))]}
    fi

    # P95 (95th percentile)
    local p95_index=$(( (count * 95) / 100 ))
    if [ $p95_index -ge $count ]; then
        p95_index=$((count - 1))
    fi
    local p95=${sorted[$p95_index]}

    # P99 (99th percentile)
    local p99_index=$(( (count * 99) / 100 ))
    if [ $p99_index -ge $count ]; then
        p99_index=$((count - 1))
    fi
    local p99=${sorted[$p99_index]}

    echo "$min $max $avg $median $p95 $p99"
}

# Print statistics table
print_stats() {
    local endpoint=$1
    local times=("${@:2}")

    stats=$(calculate_stats "${times[@]}")
    read min max avg median p95 p99 <<< "$stats"

    printf "${BLUE}%-50s${NC}\n" "$endpoint"
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Min:" $min
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Max:" $max
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Average:" $avg
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Median:" $median
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P95:" $p95
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P99:" $p99

    # Performance evaluation
    if [ $avg -lt 50 ]; then
        print_success "Excellent performance (avg < 50ms)"
    elif [ $avg -lt 100 ]; then
        print_success "Good performance (avg < 100ms)"
    elif [ $avg -lt 200 ]; then
        print_info "Acceptable performance (avg < 200ms)"
    else
        print_error "Poor performance (avg >= 200ms)"
    fi
    echo ""
}

# Measure single request
measure_request() {
    local method=$1
    local url=$2
    local data=$3

    local start=$(date +%s%N)

    if [ "$method" = "GET" ]; then
        curl -s -o /dev/null -w "%{http_code}" "$url" \
            -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    else
        curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" > /dev/null
    fi

    local end=$(date +%s%N)
    local duration=$(( (end - start) / 1000000 ))

    echo $duration
}

# Run performance test for an endpoint
run_performance_test() {
    local test_name=$1
    local method=$2
    local url=$3
    local data=$4

    print_section "$test_name"

    # Warmup
    print_info "Warming up ($WARMUP_ITERATIONS requests)..."
    for i in $(seq 1 $WARMUP_ITERATIONS); do
        measure_request "$method" "$url" "$data" > /dev/null
    done

    # Actual performance test
    print_info "Running performance test ($PERFORMANCE_ITERATIONS requests)..."

    local times=()
    local progress_step=$((PERFORMANCE_ITERATIONS / 10))

    for i in $(seq 1 $PERFORMANCE_ITERATIONS); do
        local duration=$(measure_request "$method" "$url" "$data")
        times+=($duration)

        # Progress indicator
        if [ $((i % progress_step)) -eq 0 ]; then
            echo -n "."
        fi
    done
    echo ""

    # Print statistics
    print_stats "$test_name" "${times[@]}"
}

# Run concurrent requests test
run_concurrent_test() {
    local test_name=$1
    local method=$2
    local url=$3
    local data=$4

    print_section "$test_name (Concurrent: $CONCURRENT_REQUESTS)"

    local pids=()
    local tmpdir=$(mktemp -d)

    print_info "Starting $CONCURRENT_REQUESTS concurrent requests..."
    local start=$(date +%s%N)

    for i in $(seq 1 $CONCURRENT_REQUESTS); do
        {
            local req_start=$(date +%s%N)
            if [ "$method" = "GET" ]; then
                curl -s -o /dev/null "$url" \
                    -H "Authorization: Bearer $JWT_TOKEN"
            else
                curl -s -o /dev/null -X "$method" "$url" \
                    -H "Authorization: Bearer $JWT_TOKEN" \
                    -H "Content-Type: application/json" \
                    -d "$data"
            fi
            local req_end=$(date +%s%N)
            local duration=$(( (req_end - req_start) / 1000000 ))
            echo $duration > "$tmpdir/$i.time"
        } &
        pids+=($!)
    done

    # Wait for all requests to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done

    local end=$(date +%s%N)
    local total_duration=$(( (end - start) / 1000000 ))

    # Collect results
    local times=()
    for i in $(seq 1 $CONCURRENT_REQUESTS); do
        if [ -f "$tmpdir/$i.time" ]; then
            times+=($(cat "$tmpdir/$i.time"))
        fi
    done

    rm -rf "$tmpdir"

    # Calculate statistics
    stats=$(calculate_stats "${times[@]}")
    read min max avg median p95 p99 <<< "$stats"

    print_metric "Total time: ${total_duration}ms"
    print_metric "Requests/sec: $(( (CONCURRENT_REQUESTS * 1000) / total_duration ))"
    print_metric "Individual request stats:"
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "Min:" $min
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "Max:" $max
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "Average:" $avg
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "Median:" $median

    if [ $total_duration -lt 500 ]; then
        print_success "Excellent concurrent performance"
    elif [ $total_duration -lt 1000 ]; then
        print_success "Good concurrent performance"
    else
        print_info "Acceptable concurrent performance"
    fi
    echo ""
}

# =============================================================================
# Setup Test Environment
# =============================================================================

setup_test_environment() {
    print_header "Setting Up Test Environment"

    # Get test token
    print_section "Getting Test Token"
    response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        JWT_TOKEN=$(echo "$response" | jq -r '.accessToken')
        USER_ID=$(echo "$response" | jq -r '.userId')
        print_success "Token received"
    else
        print_error "Failed to get test token"
        exit 1
    fi

    # Get or create workspace
    print_section "Setting Up Workspace"
    workspaces_response=$(curl -s "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$workspaces_response" | jq -e 'type == "array"' >/dev/null 2>&1; then
        WORKSPACE_ID=$(echo "$workspaces_response" | jq -r '.[0].id // empty')
    elif echo "$workspaces_response" | jq -e '.data' >/dev/null 2>&1; then
        WORKSPACE_ID=$(echo "$workspaces_response" | jq -r '.data[0].id // empty')
    fi

    if [ -z "$WORKSPACE_ID" ] || [ "$WORKSPACE_ID" = "null" ]; then
        print_info "No workspace found. Creating new workspace..."

        create_ws_response=$(curl -s -w "\n%{http_code}" -X POST "$USER_SERVICE_URL/api/workspaces" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "Performance Test Workspace",
                "description": "Workspace for performance testing"
            }')

        http_code=$(echo "$create_ws_response" | tail -n1)
        ws_body=$(echo "$create_ws_response" | sed '$d')

        if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
            WORKSPACE_ID=$(echo "$ws_body" | jq -r '.id')
            print_success "Created workspace: $WORKSPACE_ID"
        else
            print_error "Failed to create workspace (HTTP $http_code)"
            echo "$ws_body" | jq '.'
            exit 1
        fi
    else
        print_success "Using existing workspace: $WORKSPACE_ID"
    fi

    # Create test project
    print_section "Creating Test Project"
    project_response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "workspace_id": "'$WORKSPACE_ID'",
            "name": "Performance Test Project",
            "description": "Project for performance testing"
        }')

    http_code=$(echo "$project_response" | tail -n1)
    response_body=$(echo "$project_response" | sed '$d')

    if [ "$http_code" -eq 201 ]; then
        PROJECT_ID=$(echo "$response_body" | jq -r '.data.project_id')
        print_success "Project created: $PROJECT_ID"
    else
        print_error "Failed to create project"
        exit 1
    fi

    # Create fields for testing
    print_section "Creating Test Fields"

    # Text field
    field_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Description",
            "field_type": "text",
            "is_required": false,
            "config": {}
        }')
    TEXT_FIELD_ID=$(echo "$field_response" | jq -r '.data.field_id')

    # Single select field
    field_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Priority",
            "field_type": "single_select",
            "is_required": true,
            "config": {}
        }')
    PRIORITY_FIELD_ID=$(echo "$field_response" | jq -r '.data.field_id')

    # Create options
    option_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "field_id": "'$PRIORITY_FIELD_ID'",
            "label": "High",
            "color": "#FF0000"
        }')
    HIGH_OPTION_ID=$(echo "$option_response" | jq -r '.data.option_id')

    # Create stage and role
    stage_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/custom-fields/stages" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "To Do",
            "color": "#808080"
        }')
    STAGE_ID=$(echo "$stage_response" | jq -r '.data.stage_id // .data.id')

    role_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/custom-fields/roles" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Developer",
            "color": "#0000FF"
        }')
    ROLE_ID=$(echo "$role_response" | jq -r '.data.role_id // .data.id')

    # Create board
    board_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/boards" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "title": "Test Board",
            "content": "Performance test board",
            "stage_id": "'$STAGE_ID'",
            "role_ids": ["'$ROLE_ID'"]
        }')
    BOARD_ID=$(echo "$board_response" | jq -r '.data.board_id')

    # Create view
    view_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/views" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Test View",
            "description": "Performance test view",
            "is_shared": true,
            "filters": {},
            "sort_by": "created_at",
            "sort_direction": "desc"
        }')
    VIEW_ID=$(echo "$view_response" | jq -r '.data.view_id')

    print_success "Test environment ready"
    echo ""
    echo "Test Data IDs:"
    echo "  Project: $PROJECT_ID"
    echo "  Text Field: $TEXT_FIELD_ID"
    echo "  Priority Field: $PRIORITY_FIELD_ID"
    echo "  Board: $BOARD_ID"
    echo "  View: $VIEW_ID"
}

# =============================================================================
# Performance Tests
# =============================================================================

test_read_performance() {
    print_header "READ Performance Tests"

    # Get fields
    run_performance_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""

    # Get field options
    run_performance_test \
        "GET /api/fields/{id}/options" \
        "GET" \
        "$BOARD_SERVICE_URL/api/fields/$PRIORITY_FIELD_ID/options" \
        ""

    # Get board field values
    run_performance_test \
        "GET /api/boards/{id}/field-values" \
        "GET" \
        "$BOARD_SERVICE_URL/api/boards/$BOARD_ID/field-values" \
        ""

    # Get project views
    run_performance_test \
        "GET /api/projects/{id}/views" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/views" \
        ""
}

test_write_performance() {
    print_header "WRITE Performance Tests"

    # Set field value
    run_performance_test \
        "POST /api/board-field-values" \
        "POST" \
        "$BOARD_SERVICE_URL/api/board-field-values" \
        '{"board_id": "'$BOARD_ID'", "field_id": "'$TEXT_FIELD_ID'", "value": "Performance test value"}'
}

test_cache_performance() {
    print_header "CACHE Performance Tests"

    print_section "Testing cache effectiveness (100 requests each)"

    # First, make one request to potentially populate cache
    curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null

    # Now test cache hit performance with many requests
    print_info "Running 100 consecutive requests (cache should be hot)..."
    local cached_times=()
    for i in $(seq 1 100); do
        local duration=$(measure_request "GET" "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" "")
        cached_times+=($duration)
        if [ $((i % 20)) -eq 0 ]; then
            echo -n "."
        fi
    done
    echo ""

    stats=$(calculate_stats "${cached_times[@]}")
    read min max avg median p95 p99 <<< "$stats"

    print_metric "100 Cached Requests Performance:"
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Min:" $min
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Max:" $max
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Average:" $avg
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "Median:" $median
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P95:" $p95
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P99:" $p99

    # Cache effectiveness check
    if [ $avg -lt 20 ]; then
        print_success "Excellent cache performance! (avg < 20ms)"
    elif [ $avg -lt 50 ]; then
        print_success "Good cache performance (avg < 50ms)"
    else
        print_error "Cache might not be effective (avg >= 50ms)"
    fi

    # Consistency check
    local variation=$(( max - min ))
    print_metric "Response time variation: ${variation}ms (max - min)"
    if [ $variation -lt 50 ]; then
        print_success "Very consistent performance"
    elif [ $variation -lt 100 ]; then
        print_info "Reasonably consistent performance"
    else
        print_error "High variance - might indicate cache issues"
    fi
    echo ""
}

test_concurrent_performance() {
    print_header "CONCURRENT Request Tests"

    # Light load
    print_section "Light Load: $CONCURRENT_REQUESTS_LIGHT concurrent requests"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_LIGHT
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""

    # Medium load
    print_section "Medium Load: $CONCURRENT_REQUESTS_MEDIUM concurrent requests"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_MEDIUM
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""

    # Heavy load
    print_section "Heavy Load: $CONCURRENT_REQUESTS_HEAVY concurrent requests"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_HEAVY
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""
}

test_sustained_load() {
    print_header "SUSTAINED Load Tests"

    print_section "Sustained load: ${SUSTAINED_TEST_DURATION} seconds continuous requests"
    print_info "Sending requests as fast as possible for ${SUSTAINED_TEST_DURATION} seconds..."

    local request_count=0
    local success_count=0
    local error_count=0
    local total_time=0
    local times=()

    local end_time=$(($(date +%s) + SUSTAINED_TEST_DURATION))

    while [ $(date +%s) -lt $end_time ]; do
        local start=$(date +%s%N)
        local status=$(curl -s -o /dev/null -w "%{http_code}" "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
            -H "Authorization: Bearer $JWT_TOKEN")
        local end=$(date +%s%N)
        local duration=$(( (end - start) / 1000000 ))

        request_count=$((request_count + 1))
        times+=($duration)
        total_time=$((total_time + duration))

        if [ "$status" = "200" ]; then
            success_count=$((success_count + 1))
        else
            error_count=$((error_count + 1))
        fi

        # Progress indicator every 20 requests
        if [ $((request_count % 20)) -eq 0 ]; then
            echo -n "."
        fi
    done
    echo ""

    # Calculate statistics
    stats=$(calculate_stats "${times[@]}")
    read min max avg median p95 p99 <<< "$stats"

    local rps=$(( request_count / SUSTAINED_TEST_DURATION ))

    print_metric "Sustained Load Results:"
    printf "  ${MAGENTA}%-25s${NC} %d\n" "Total requests:" $request_count
    printf "  ${MAGENTA}%-25s${NC} %d\n" "Successful:" $success_count
    printf "  ${MAGENTA}%-25s${NC} %d\n" "Errors:" $error_count
    printf "  ${MAGENTA}%-25s${NC} %d req/sec\n" "Throughput:" $rps
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "Avg response time:" $avg
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "P95 response time:" $p95
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "P99 response time:" $p99

    # Performance evaluation
    if [ $error_count -eq 0 ]; then
        print_success "No errors during sustained load!"
    else
        local error_rate=$(( (error_count * 100) / request_count ))
        print_error "Error rate: ${error_rate}%"
    fi

    if [ $rps -gt 100 ]; then
        print_success "Excellent throughput (> 100 req/sec)"
    elif [ $rps -gt 50 ]; then
        print_success "Good throughput (> 50 req/sec)"
    elif [ $rps -gt 20 ]; then
        print_info "Acceptable throughput (> 20 req/sec)"
    else
        print_error "Low throughput (< 20 req/sec)"
    fi
    echo ""
}

test_load_scenarios() {
    print_header "LOAD Scenario Tests"

    print_section "Scenario 1: Typical User Workflow"
    print_info "Simulating: View project → View board → Update field value"

    local workflow_start=$(date +%s%N)

    # Step 1: Get fields
    measure_request "GET" "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" "" > /dev/null

    # Step 2: Get board field values
    measure_request "GET" "$BOARD_SERVICE_URL/api/boards/$BOARD_ID/field-values" "" > /dev/null

    # Step 3: Update field value
    measure_request "POST" "$BOARD_SERVICE_URL/api/board-field-values" \
        '{"board_id": "'$BOARD_ID'", "field_id": "'$TEXT_FIELD_ID'", "value": "Workflow test"}' > /dev/null

    local workflow_end=$(date +%s%N)
    local workflow_duration=$(( (workflow_end - workflow_start) / 1000000 ))

    print_metric "Total workflow time: ${workflow_duration}ms"

    if [ $workflow_duration -lt 200 ]; then
        print_success "Excellent user experience (< 200ms)"
    elif [ $workflow_duration -lt 500 ]; then
        print_success "Good user experience (< 500ms)"
    else
        print_info "Acceptable user experience"
    fi
    echo ""
}

# =============================================================================
# Cleanup
# =============================================================================

cleanup() {
    print_header "Cleanup"
    print_info "Cleaning up test data..."

    # Delete project (cascades to all related data)
    curl -s -X DELETE "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null

    print_success "Test data cleaned up"
}

# =============================================================================
# Summary
# =============================================================================

print_performance_summary() {
    print_header "PERFORMANCE TEST SUMMARY"

    echo -e "${GREEN}All performance tests completed!${NC}"
    echo ""
    echo "Test Configuration:"
    echo "  Warmup iterations: $WARMUP_ITERATIONS"
    echo "  Performance iterations: $PERFORMANCE_ITERATIONS"
    echo "  Light concurrent load: $CONCURRENT_REQUESTS_LIGHT requests"
    echo "  Medium concurrent load: $CONCURRENT_REQUESTS_MEDIUM requests"
    echo "  Heavy concurrent load: $CONCURRENT_REQUESTS_HEAVY requests"
    echo "  Sustained load duration: ${SUSTAINED_TEST_DURATION}s"
    echo ""
    echo "Performance Benchmarks:"
    echo "  Response Time:"
    echo "    - < 50ms: Excellent ⭐⭐⭐"
    echo "    - < 100ms: Good ⭐⭐"
    echo "    - < 200ms: Acceptable ⭐"
    echo ""
    echo "  Throughput:"
    echo "    - > 100 req/sec: Excellent ⭐⭐⭐"
    echo "    - > 50 req/sec: Good ⭐⭐"
    echo "    - > 20 req/sec: Acceptable ⭐"
    echo ""
    echo "  Cache:"
    echo "    - < 20ms avg: Excellent cache hit"
    echo "    - < 50ms variation: Consistent performance"
    echo ""
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║   Custom Fields System - Performance Test Suite              ║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Check if services are running
    if ! curl -s "$USER_SERVICE_URL/health" > /dev/null 2>&1; then
        print_error "User service is not running at $USER_SERVICE_URL"
        exit 1
    fi

    if ! curl -s "$BOARD_SERVICE_URL/health" > /dev/null 2>&1; then
        print_error "Board service is not running at $BOARD_SERVICE_URL"
        exit 1
    fi

    # Run tests
    setup_test_environment
    test_read_performance
    test_write_performance
    test_cache_performance
    test_concurrent_performance
    test_sustained_load
    test_load_scenarios
    cleanup
    print_performance_summary
}

# Handle interrupts
trap cleanup EXIT

# Run main
main
