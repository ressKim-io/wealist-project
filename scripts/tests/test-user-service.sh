#!/bin/bash

# ===== User Service 테스트 스크립트 =====
# user-service가 정상적으로 동작하는지 확인하는 간단한 스크립트

BASE_URL="http://localhost:8080"

# 색상 설정
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ===== 헬퍼 함수 =====
print_section() {
    echo -e "\n${BLUE}===== $1 =====${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# HTTP 상태 코드 확인 함수
check_http_status() {
    local status=$1
    local expected=$2
    local message=$3

    if [ "$status" -eq "$expected" ]; then
        print_success "$message (HTTP $status)"
        return 0
    else
        print_error "$message (Expected: HTTP $expected, Got: HTTP $status)"
        return 1
    fi
}

# ===== 1. Health Check =====
print_section "1. Health Check"

HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/actuator/health")
HEALTH_HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
HEALTH_BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if check_http_status "$HEALTH_HTTP_CODE" 200 "Health check"; then
    echo "$HEALTH_BODY" | head -5
fi

# ===== 2. Swagger UI 접근 가능 여부 =====
print_section "2. Swagger UI Check"

SWAGGER_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger-ui.html")

if check_http_status "$SWAGGER_RESPONSE" 200 "Swagger UI accessible"; then
    print_info "Swagger UI URL: $BASE_URL/swagger-ui.html"
fi

# ===== 3. API Docs 확인 =====
print_section "3. OpenAPI Docs Check"

API_DOCS_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v3/api-docs")
API_DOCS_HTTP_CODE=$(echo "$API_DOCS_RESPONSE" | tail -n1)
API_DOCS_BODY=$(echo "$API_DOCS_RESPONSE" | sed '$d')

if check_http_status "$API_DOCS_HTTP_CODE" 200 "OpenAPI docs available"; then
    # ProfileController 엔드포인트 확인
    if echo "$API_DOCS_BODY" | grep -q "/api/profiles/me"; then
        print_success "ProfileController endpoints found in API docs"
        print_info "  - GET /api/profiles/me"
        print_info "  - PUT /api/profiles/me"
    else
        print_error "ProfileController endpoints NOT found in API docs"
        print_warning "user-service may not be rebuilt with latest code"
    fi
fi

# ===== 4. 프로필 엔드포인트 테스트 (인증 없음) =====
print_section "4. Profile Endpoints Test (Without Auth)"

print_info "Testing GET /api/profiles/me without authentication..."
PROFILE_NO_AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/profiles/me")
PROFILE_NO_AUTH_HTTP_CODE=$(echo "$PROFILE_NO_AUTH_RESPONSE" | tail -n1)

# 인증 없으면 401 또는 403이 정상
if [ "$PROFILE_NO_AUTH_HTTP_CODE" -eq 401 ] || [ "$PROFILE_NO_AUTH_HTTP_CODE" -eq 403 ]; then
    print_success "Endpoint exists and requires authentication (HTTP $PROFILE_NO_AUTH_HTTP_CODE)"
elif [ "$PROFILE_NO_AUTH_HTTP_CODE" -eq 404 ]; then
    print_error "Endpoint NOT found (HTTP 404)"
    print_warning "Please rebuild user-service with: docker compose -f docker-compose.base.yml -f docker-compose.local.yml up -d --build user-service"
else
    print_warning "Unexpected status code: HTTP $PROFILE_NO_AUTH_HTTP_CODE"
fi

# ===== 5. 토큰을 사용한 프로필 엔드포인트 테스트 =====
print_section "5. Profile Endpoints Test (With Auth)"

# 환경 변수에서 토큰 읽기 또는 파일에서 읽기
ACCESS_TOKEN="${ACCESS_TOKEN:-}"

if [ -f ".test-token" ]; then
    ACCESS_TOKEN=$(cat .test-token)
fi

if [ -z "$ACCESS_TOKEN" ]; then
    print_warning "No ACCESS_TOKEN provided"
    print_info "To test authenticated endpoints:"
    print_info "  1. Login to the app and copy your access_token from localStorage"
    print_info "  2. Run: export ACCESS_TOKEN='your-token-here'"
    print_info "  3. Or save to .test-token file: echo 'your-token' > .test-token"
    print_info "  4. Then run this script again"
else
    print_info "Testing with provided ACCESS_TOKEN..."

    PROFILE_AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/profiles/me" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    PROFILE_AUTH_HTTP_CODE=$(echo "$PROFILE_AUTH_RESPONSE" | tail -n1)
    PROFILE_AUTH_BODY=$(echo "$PROFILE_AUTH_RESPONSE" | sed '$d')

    if check_http_status "$PROFILE_AUTH_HTTP_CODE" 200 "Profile fetch with authentication"; then
        echo "$PROFILE_AUTH_BODY" | head -10

        # email 필드가 있는지 확인
        if echo "$PROFILE_AUTH_BODY" | grep -q '"email"'; then
            print_success "Email field found in response"
        else
            print_warning "Email field NOT found in response (may be null)"
        fi
    fi
fi

# ===== 6. 워크스페이스 엔드포인트 테스트 =====
print_section "6. Workspace Endpoints Check"

if [ -n "$ACCESS_TOKEN" ]; then
    WORKSPACE_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/workspaces" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    WORKSPACE_HTTP_CODE=$(echo "$WORKSPACE_RESPONSE" | tail -n1)
    WORKSPACE_BODY=$(echo "$WORKSPACE_RESPONSE" | sed '$d')

    if check_http_status "$WORKSPACE_HTTP_CODE" 200 "Workspace list fetch"; then
        WORKSPACE_COUNT=$(echo "$WORKSPACE_BODY" | grep -o '"id"' | wc -l)
        print_info "Found $WORKSPACE_COUNT workspace(s)"
    fi
else
    print_warning "Skipping workspace test (no ACCESS_TOKEN)"
fi

# ===== 7. 요약 =====
print_section "Summary"

echo "Service URL: $BASE_URL"
echo "Health: $([ "$HEALTH_HTTP_CODE" -eq 200 ] && echo "✓ OK" || echo "✗ FAIL")"
echo "Swagger: $([ "$SWAGGER_RESPONSE" -eq 200 ] && echo "✓ OK" || echo "✗ FAIL")"
echo "Profile Endpoint: $([ "$PROFILE_NO_AUTH_HTTP_CODE" -ne 404 ] && echo "✓ EXISTS" || echo "✗ NOT FOUND")"

if [ "$PROFILE_NO_AUTH_HTTP_CODE" -eq 404 ]; then
    echo ""
    print_error "Profile endpoint not found! Please rebuild user-service:"
    echo "  docker compose -f docker-compose.base.yml -f docker-compose.local.yml stop user-service"
    echo "  docker compose -f docker-compose.base.yml -f docker-compose.local.yml rm -f user-service"
    echo "  docker compose -f docker-compose.base.yml -f docker-compose.local.yml build --no-cache user-service"
    echo "  docker compose -f docker-compose.base.yml -f docker-compose.local.yml up -d user-service"
fi

echo ""
print_success "Test completed!"
