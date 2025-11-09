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
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "최소:" $min
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "최대:" $max
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "평균:" $avg
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "중간값:" $median
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P95:" $p95
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P99:" $p99

    # Performance evaluation
    if [ $avg -lt 50 ]; then
        print_success "우수한 성능 (평균 < 50ms)"
    elif [ $avg -lt 100 ]; then
        print_success "좋은 성능 (평균 < 100ms)"
    elif [ $avg -lt 200 ]; then
        print_info "적절한 성능 (평균 < 200ms)"
    else
        print_error "낮은 성능 (평균 >= 200ms)"
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
    print_info "워밍업 중 ($WARMUP_ITERATIONS 요청)..."
    for i in $(seq 1 $WARMUP_ITERATIONS); do
        measure_request "$method" "$url" "$data" > /dev/null
    done

    # Actual performance test
    print_info "성능 테스트 실행 중 ($PERFORMANCE_ITERATIONS 요청)..."

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

    print_section "$test_name (동시: $CONCURRENT_REQUESTS)"

    local pids=()
    local tmpdir=$(mktemp -d)

    print_info "$CONCURRENT_REQUESTS개 동시 요청 시작..."
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

    print_metric "총 소요 시간: ${total_duration}ms"
    print_metric "초당 요청 수: $(( (CONCURRENT_REQUESTS * 1000) / total_duration ))"
    print_metric "개별 요청 통계:"
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "최소:" $min
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "최대:" $max
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "평균:" $avg
    printf "    ${MAGENTA}%-15s${NC} %d ms\n" "중간값:" $median

    if [ $total_duration -lt 500 ]; then
        print_success "우수한 동시 처리 성능"
    elif [ $total_duration -lt 1000 ]; then
        print_success "좋은 동시 처리 성능"
    else
        print_info "적절한 동시 처리 성능"
    fi
    echo ""
}

# =============================================================================
# Setup Test Environment
# =============================================================================

setup_test_environment() {
    print_header "테스트 환경 설정"

    # Get test token
    print_section "테스트 토큰 가져오기"
    response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        JWT_TOKEN=$(echo "$response" | jq -r '.accessToken')
        USER_ID=$(echo "$response" | jq -r '.userId')
        print_success "토큰 수신 완료"
    else
        print_error "테스트 토큰 가져오기 실패"
        exit 1
    fi

    # Get or create workspace
    print_section "워크스페이스 설정"
    workspaces_response=$(curl -s "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$workspaces_response" | jq -e 'type == "array"' >/dev/null 2>&1; then
        WORKSPACE_ID=$(echo "$workspaces_response" | jq -r '.[0].id // empty')
    elif echo "$workspaces_response" | jq -e '.data' >/dev/null 2>&1; then
        WORKSPACE_ID=$(echo "$workspaces_response" | jq -r '.data[0].id // empty')
    fi

    if [ -z "$WORKSPACE_ID" ] || [ "$WORKSPACE_ID" = "null" ]; then
        print_info "워크스페이스를 찾을 수 없습니다. 새 워크스페이스 생성 중..."

        create_ws_response=$(curl -s -w "\n%{http_code}" -X POST "$USER_SERVICE_URL/api/workspaces" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "성능 테스트 워크스페이스",
                "description": "성능 테스트용 워크스페이스"
            }')

        http_code=$(echo "$create_ws_response" | tail -n1)
        ws_body=$(echo "$create_ws_response" | sed '$d')

        if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
            WORKSPACE_ID=$(echo "$ws_body" | jq -r '.id')
            print_success "워크스페이스 생성 완료: $WORKSPACE_ID"
        else
            print_error "워크스페이스 생성 실패 (HTTP $http_code)"
            echo "$ws_body" | jq '.'
            exit 1
        fi
    else
        print_success "기존 워크스페이스 사용: $WORKSPACE_ID"
    fi

    # Create test project
    print_section "테스트 프로젝트 생성"
    project_response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "workspace_id": "'$WORKSPACE_ID'",
            "name": "성능 테스트 프로젝트",
            "description": "성능 테스트용 프로젝트"
        }')

    http_code=$(echo "$project_response" | tail -n1)
    response_body=$(echo "$project_response" | sed '$d')

    if [ "$http_code" -eq 201 ]; then
        PROJECT_ID=$(echo "$response_body" | jq -r '.data.project_id')
        print_success "프로젝트 생성 완료: $PROJECT_ID"
    else
        print_error "프로젝트 생성 실패"
        exit 1
    fi

    # Create fields for testing
    print_section "테스트 필드 생성"

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

    print_success "테스트 환경 준비 완료"
    echo ""
    echo "테스트 데이터 ID:"
    echo "  프로젝트: $PROJECT_ID"
    echo "  텍스트 필드: $TEXT_FIELD_ID"
    echo "  우선순위 필드: $PRIORITY_FIELD_ID"
    echo "  보드: $BOARD_ID"
    echo "  뷰: $VIEW_ID"
}

# =============================================================================
# Performance Tests
# =============================================================================

test_read_performance() {
    print_header "읽기 성능 테스트"

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
    print_header "쓰기 성능 테스트"

    # Set field value
    run_performance_test \
        "POST /api/board-field-values" \
        "POST" \
        "$BOARD_SERVICE_URL/api/board-field-values" \
        '{"board_id": "'$BOARD_ID'", "field_id": "'$TEXT_FIELD_ID'", "value": "성능 테스트 값"}'
}

test_cache_performance() {
    print_header "캐시 성능 테스트"

    print_section "캐시 효과 테스트 (100번 요청)"

    # First, make one request to potentially populate cache
    curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null

    # Now test cache hit performance with many requests
    print_info "100번 연속 요청 실행 중 (캐시가 활성화되어야 함)..."
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

    print_metric "100번 캐시 요청 성능:"
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "최소:" $min
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "최대:" $max
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "평균:" $avg
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "중간값:" $median
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P95:" $p95
    printf "  ${MAGENTA}%-15s${NC} %d ms\n" "P99:" $p99

    # Cache effectiveness check
    if [ $avg -lt 20 ]; then
        print_success "우수한 캐시 성능! (평균 < 20ms)"
    elif [ $avg -lt 50 ]; then
        print_success "좋은 캐시 성능 (평균 < 50ms)"
    else
        print_error "캐시가 효과적이지 않을 수 있음 (평균 >= 50ms)"
    fi

    # Consistency check
    local variation=$(( max - min ))
    print_metric "응답 시간 변동폭: ${variation}ms (최대 - 최소)"
    if [ $variation -lt 50 ]; then
        print_success "매우 일관된 성능"
    elif [ $variation -lt 100 ]; then
        print_info "적절히 일관된 성능"
    else
        print_error "높은 변동성 - 캐시 문제 가능성"
    fi
    echo ""
}

test_concurrent_performance() {
    print_header "동시 요청 테스트"

    # Light load
    print_section "가벼운 부하: $CONCURRENT_REQUESTS_LIGHT개 동시 요청"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_LIGHT
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""

    # Medium load
    print_section "중간 부하: $CONCURRENT_REQUESTS_MEDIUM개 동시 요청"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_MEDIUM
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""

    # Heavy load
    print_section "높은 부하: $CONCURRENT_REQUESTS_HEAVY개 동시 요청"
    CONCURRENT_REQUESTS=$CONCURRENT_REQUESTS_HEAVY
    run_concurrent_test \
        "GET /api/projects/{id}/fields" \
        "GET" \
        "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        ""
}

test_sustained_load() {
    print_header "지속 부하 테스트"

    print_section "지속 부하: ${SUSTAINED_TEST_DURATION}초 동안 연속 요청"
    print_info "${SUSTAINED_TEST_DURATION}초 동안 최대한 빠르게 요청 전송 중..."

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

    print_metric "지속 부하 결과:"
    printf "  ${MAGENTA}%-25s${NC} %d\n" "총 요청 수:" $request_count
    printf "  ${MAGENTA}%-25s${NC} %d\n" "성공:" $success_count
    printf "  ${MAGENTA}%-25s${NC} %d\n" "에러:" $error_count
    printf "  ${MAGENTA}%-25s${NC} %d req/sec\n" "처리량:" $rps
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "평균 응답 시간:" $avg
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "P95 응답 시간:" $p95
    printf "  ${MAGENTA}%-25s${NC} %d ms\n" "P99 응답 시간:" $p99

    # Performance evaluation
    if [ $error_count -eq 0 ]; then
        print_success "지속 부하 중 에러 없음!"
    else
        local error_rate=$(( (error_count * 100) / request_count ))
        print_error "에러율: ${error_rate}%"
    fi

    if [ $rps -gt 100 ]; then
        print_success "우수한 처리량 (> 100 req/sec)"
    elif [ $rps -gt 50 ]; then
        print_success "좋은 처리량 (> 50 req/sec)"
    elif [ $rps -gt 20 ]; then
        print_info "적절한 처리량 (> 20 req/sec)"
    else
        print_error "낮은 처리량 (< 20 req/sec)"
    fi
    echo ""
}

test_load_scenarios() {
    print_header "부하 시나리오 테스트"

    print_section "시나리오 1: 일반 사용자 워크플로우"
    print_info "시뮬레이션: 프로젝트 조회 → 보드 조회 → 필드 값 업데이트"

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

    print_metric "총 워크플로우 시간: ${workflow_duration}ms"

    if [ $workflow_duration -lt 200 ]; then
        print_success "우수한 사용자 경험 (< 200ms)"
    elif [ $workflow_duration -lt 500 ]; then
        print_success "좋은 사용자 경험 (< 500ms)"
    else
        print_info "적절한 사용자 경험"
    fi
    echo ""
}

# =============================================================================
# Cleanup
# =============================================================================

cleanup() {
    print_header "정리"
    print_info "테스트 데이터 정리 중..."

    # Delete project (cascades to all related data)
    curl -s -X DELETE "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null

    print_success "테스트 데이터 정리 완료"
}

# =============================================================================
# Summary
# =============================================================================

print_performance_summary() {
    print_header "성능 테스트 요약"

    echo -e "${GREEN}모든 성능 테스트 완료!${NC}"
    echo ""
    echo "테스트 설정:"
    echo "  워밍업 반복: $WARMUP_ITERATIONS"
    echo "  성능 테스트 반복: $PERFORMANCE_ITERATIONS"
    echo "  가벼운 동시 부하: $CONCURRENT_REQUESTS_LIGHT 요청"
    echo "  중간 동시 부하: $CONCURRENT_REQUESTS_MEDIUM 요청"
    echo "  높은 동시 부하: $CONCURRENT_REQUESTS_HEAVY 요청"
    echo "  지속 부하 시간: ${SUSTAINED_TEST_DURATION}초"
    echo ""
    echo "성능 기준:"
    echo "  응답 시간:"
    echo "    - < 50ms: 우수 ⭐⭐⭐"
    echo "    - < 100ms: 좋음 ⭐⭐"
    echo "    - < 200ms: 적절 ⭐"
    echo ""
    echo "  처리량:"
    echo "    - > 100 req/sec: 우수 ⭐⭐⭐"
    echo "    - > 50 req/sec: 좋음 ⭐⭐"
    echo "    - > 20 req/sec: 적절 ⭐"
    echo ""
    echo "  캐시:"
    echo "    - < 20ms 평균: 우수한 캐시 히트"
    echo "    - < 50ms 변동: 일관된 성능"
    echo ""
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║       커스텀 필드 시스템 - 성능 테스트 스위트                 ║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Check if services are running
    if ! curl -s "$USER_SERVICE_URL/health" > /dev/null 2>&1; then
        print_error "User 서비스가 $USER_SERVICE_URL 에서 실행되지 않습니다"
        exit 1
    fi

    if ! curl -s "$BOARD_SERVICE_URL/health" > /dev/null 2>&1; then
        print_error "Board 서비스가 $BOARD_SERVICE_URL 에서 실행되지 않습니다"
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
