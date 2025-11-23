#!/bin/bash

# =============================================================================
# Board Service 통합 테스트 스크립트 (New Custom Fields System)
# =============================================================================
# Prerequisites:
# 1. User Service must be running (localhost:8080)
# 2. Board Service must be running (localhost:8000)
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
BOARD_SERVICE_URL="${BOARD_SERVICE_URL:-http://localhost:8000}"
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"

# Global variables
TOKEN=""
USER_ID=""
WORKSPACE_ID=""
PROJECT_ID=""
FIELD_STATUS_ID=""
FIELD_PRIORITY_ID=""
FIELD_TAGS_ID=""
OPTION_TODO_ID=""
OPTION_INPROGRESS_ID=""
OPTION_DONE_ID=""
OPTION_HIGH_ID=""
OPTION_MEDIUM_ID=""
OPTION_LOW_ID=""
BOARD_ID=""
COMMENT_ID=""

# =============================================================================
# Utility Functions
# =============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
}

print_step() {
    echo ""
    echo -e "${CYAN}📋 STEP $1: $2${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_json() {
    echo "$1" | jq '.' 2>/dev/null || echo "$1"
}

# =============================================================================
# Test Functions
# =============================================================================

test_health_check() {
    print_step "1" "Health Check"

    response=$(curl -s "$BOARD_SERVICE_URL/health")

    if echo "$response" | grep -q '"status":"healthy"'; then
        print_success "Board Service is healthy"
        print_json "$response"
    else
        print_error "Health check failed: $response"
    fi
}

get_test_token() {
    print_step "2" "Get Test Token from User Service"

    echo "Calling: $USER_SERVICE_URL/api/auth/test"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$USER_SERVICE_URL/api/auth/test")

    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    response_body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')

    echo "HTTP Code: $http_code"
    echo "Response: $response_body"

    if echo "$response_body" | grep -q '"accessToken"'; then
        TOKEN=$(echo "$response_body" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
        USER_ID=$(echo "$response_body" | grep -o '"userId":"[^"]*"' | cut -d'"' -f4)

        echo "Extracted TOKEN: ${TOKEN:0:50}..."
        echo "Extracted USER_ID: $USER_ID"

        if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
            print_success "토큰 생성 성공 (User ID: ${USER_ID:0:8}...)"
        else
            print_error "토큰 추출 실패"
        fi
    else
        print_error "테스트 토큰 생성 실패: $response_body"
    fi
}

create_workspace() {
    print_step "3" "Workspace 생성 (User Service)"

    workspace_data="{\"workspaceName\":\"Test Workspace $(date +%s)\",\"workspaceDescription\":\"자동 테스트용 워크스페이스\"}"

    workspace_response=$(curl -s -X POST "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$workspace_data")

    # HTTP 상태 코드도 함께 확인
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$workspace_data")

    if [ "$http_code" = "200" ] && echo "$workspace_response" | grep -q '"workspaceId"'; then
        WORKSPACE_ID=$(echo "$workspace_response" | grep -o '"workspaceId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Workspace 생성 성공 (ID: ${WORKSPACE_ID:0:8}...)"
    else
        print_error "Workspace 생성 실패: HTTP $http_code, Response: $workspace_response"
    fi
}



create_project() {
    print_step "4" "Project 생성"

    echo "Using TOKEN: ${TOKEN:0:50}..."
    echo "Using WORKSPACE_ID: $WORKSPACE_ID"

    project_data="{\"workspaceId\":\"$WORKSPACE_ID\",\"name\":\"Test Project $(date +%s)\",\"description\":\"자동 테스트용 프로젝트\"}"
    echo "Request data: $project_data"

    project_response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$project_data")

    http_code=$(echo $project_response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    response_body=$(echo $project_response | sed -e 's/HTTPSTATUS\:.*//g')

    echo "HTTP Code: $http_code"
    echo "Response: $response_body"

    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        PROJECT_ID=$(echo "$response_body" | grep -o '"projectId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Project 생성 성공 (ID: ${PROJECT_ID:0:8}...)"
    else
        print_error "Project 생성 실패: HTTP $http_code, $response_body"
    fi
}

create_field_status() {
    print_step "5" "Custom Field 생성: Status (single_select)"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"Status\",
            \"fieldType\": \"single_select\",
            \"description\": \"Task status\",
            \"isRequired\": true,
            \"config\": {}
        }")

    if echo "$response" | grep -q '"data"'; then
        FIELD_STATUS_ID=$(echo "$response" | grep -o '"fieldId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Status 필드 생성 성공 (ID: ${FIELD_STATUS_ID:0:8}...)"
    else
        print_error "Status 필드 생성 실패: $response"
    fi
}

create_status_options() {
    print_step "6" "Status Options 생성 (To Do, In Progress, Done)"

    # To Do
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"label\": \"To Do\",
            \"color\": \"#94A3B8\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_TODO_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "To Do 옵션 생성 성공"
    fi

    # In Progress
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"label\": \"In Progress\",
            \"color\": \"#3B82F6\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_INPROGRESS_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "In Progress 옵션 생성 성공"
    fi

    # Done
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"label\": \"Done\",
            \"color\": \"#10B981\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_DONE_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Done 옵션 생성 성공"
    fi
}

create_field_priority() {
    print_step "7" "Custom Field 생성: Priority (single_select)"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"Priority\",
            \"fieldType\": \"single_select\",
            \"description\": \"Task priority level\",
            \"isRequired\": false,
            \"config\": {}
        }")

    if echo "$response" | grep -q '"data"'; then
        FIELD_PRIORITY_ID=$(echo "$response" | grep -o '"fieldId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Priority 필드 생성 성공 (ID: ${FIELD_PRIORITY_ID:0:8}...)"
    else
        print_error "Priority 필드 생성 실패: $response"
    fi
}

create_priority_options() {
    print_step "8" "Priority Options 생성 (High, Medium, Low)"

    # High
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"label\": \"High\",
            \"color\": \"#EF4444\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_HIGH_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "High 옵션 생성 성공"
    fi

    # Medium
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"label\": \"Medium\",
            \"color\": \"#F59E0B\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_MEDIUM_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Medium 옵션 생성 성공"
    fi

    # Low
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"label\": \"Low\",
            \"color\": \"#6B7280\"
        }")

    if echo "$response" | grep -q '"data"'; then
        OPTION_LOW_ID=$(echo "$response" | grep -o '"optionId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Low 옵션 생성 성공"
    fi
}

create_field_tags() {
    print_step "9" "Create Custom Field: Tags (multi_select)"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"Tags\",
            \"fieldType\": \"multi_select\",
            \"description\": \"Task tags\",
            \"isRequired\": false,
            \"config\": {\"max_selections\": 5}
        }")

    if echo "$response" | grep -q '"data"'; then
        FIELD_TAGS_ID=$(echo "$response" | grep -o '"fieldId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Tags 필드 생성 성공 (ID: ${FIELD_TAGS_ID:0:8}...)"
    else
        print_error "Tags 필드 생성 실패: $response"
    fi
}

list_project_fields() {
    print_step "10" "List All Project Fields"

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q '"data"'; then
        field_count=$(echo "$response" | grep -o '"fieldId"' | wc -l)
        print_success "Retrieved $field_count custom fields"
    else
        print_error "Failed to list project fields"
    fi
}

create_board() {
    print_step "11" "Create Board (custom_fields_cache will be auto-populated)"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/boards" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"title\": \"Test Board $(date +%s)\",
            \"content\": \"Board with custom fields\",
            \"assigneeId\": \"$USER_ID\"
        }")

    if echo "$response" | grep -q '"data"'; then
        BOARD_ID=$(echo "$response" | grep -o '"boardId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Board 생성 성공 (ID: ${BOARD_ID:0:8}...)"
    else
        print_error "Board 생성 실패: $response"
    fi
}

set_board_field_values() {
    print_step "12" "Set Board Field Values"

    # Set Status = In Progress
    print_info "Setting Status to 'In Progress'..."
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/board-field-values" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"boardId\": \"$BOARD_ID\",
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"value\": \"$OPTION_INPROGRESS_ID\"
        }")

    if echo "$response" | grep -q '"data"'; then
        print_success "Status 필드 값 설정 성공"
    fi

    # Set Priority = High
    print_info "Setting Priority to 'High'..."
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/board-field-values" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"boardId\": \"$BOARD_ID\",
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"value\": \"$OPTION_HIGH_ID\"
        }")

    if echo "$response" | grep -q '"data"'; then
        print_success "Priority 필드 값 설정 성공"
    fi
}

get_board_with_fields() {
    print_step "13" "Board 조회 (Custom Fields 포함)"

    response=$(curl -s "$BOARD_SERVICE_URL/api/boards/$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q '"data"'; then
        print_success "Board 조회 성공"
    else
        print_error "Board 조회 실패"
    fi
}

get_boards_in_project() {
    print_step "14" "Project의 모든 Board 조회"

    response=$(curl -s "$BOARD_SERVICE_URL/api/boards?projectId=$PROJECT_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q '"data"'; then
        board_count=$(echo "$response" | grep -o '"boardId"' | wc -l)
        print_success "$board_count개의 Board 조회 성공"
    else
        print_error "Boards 조회 실패"
    fi
}

create_comment() {
    print_step "15" "Create Comment on Board"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/comments" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"boardId\": \"$BOARD_ID\",
            \"content\": \"This is a test comment from integration test script\"
        }")

    if echo "$response" | grep -q '"data"'; then
        COMMENT_ID=$(echo "$response" | grep -o '"commentId":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Comment 생성 성공 (ID: ${COMMENT_ID:0:8}...)"
    else
        print_error "Comment 생성 실패"
    fi
}

get_comments() {
    print_step "16" "Board Comments 조회"

    response=$(curl -s "$BOARD_SERVICE_URL/api/comments?boardId=$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q '"data"'; then
        comment_count=$(echo "$response" | grep -o '"commentId"' | wc -l)
        print_success "$comment_count개의 Comment 조회 성공"
    else
        print_error "Comments 조회 실패"
    fi
}

test_board_filtering() {
    print_step "17" "Board 필터링 테스트"

    print_info "Filter by Status = In Progress..."
    response=$(curl -s "$BOARD_SERVICE_URL/api/boards?projectId=$PROJECT_ID&status=In%20Progress" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q '"data"'; then
        filtered_count=$(echo "$response" | grep -o '"boardId"' | wc -l)
        print_success "필터링된 Board: $filtered_count개"
    fi
}

update_board() {
    print_step "18" "Board 업데이트"

    response=$(curl -s -X PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"title\": \"Updated Board Title\",
            \"description\": \"Updated description\"
        }")

    if echo "$response" | grep -q '"data"'; then
        print_success "Board 업데이트 성공"
    else
        print_error "Board 업데이트 실패"
    fi
}

# =============================================================================
# Summary & Cleanup
# =============================================================================

print_summary() {
    print_header "Test Summary"

    echo -e "${GREEN}✅ Created Resources:${NC}"
    echo -e "   User ID:           ${USER_ID:0:8}..."
    echo -e "   Workspace ID:      ${WORKSPACE_ID:0:8}..."
    echo -e "   Project ID:        ${PROJECT_ID:0:8}..."
    echo -e ""
    echo -e "   Status Field ID:   ${FIELD_STATUS_ID:0:8}..."
    echo -e "   Priority Field ID: ${FIELD_PRIORITY_ID:0:8}..."
    echo -e "   Tags Field ID:     ${FIELD_TAGS_ID:0:8}..."
    echo -e ""
    echo -e "   Board ID:          ${BOARD_ID:0:8}..."
    echo -e "   Comment ID:        ${COMMENT_ID:0:8}..."
    echo ""

    print_info "You can continue testing with:"
    echo "  export TOKEN=\"$TOKEN\""
    echo "  export PROJECT_ID=\"$PROJECT_ID\""
    echo "  export BOARD_ID=\"$BOARD_ID\""
    echo ""
    echo "  curl -H \"Authorization: Bearer \$TOKEN\" $BOARD_SERVICE_URL/api/boards/\$BOARD_ID"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  Board Service Integration Test Suite                 ║${NC}"
    echo -e "${GREEN}║  (New Custom Fields System)                            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"

    # Run all tests
    test_health_check
    get_test_token
    create_workspace
    create_project
    create_field_status
    create_status_options
    create_field_priority
    create_priority_options
    create_field_tags
    list_project_fields
    create_board
    set_board_field_values
    get_board_with_fields
    get_boards_in_project
    create_comment
    get_comments
    test_board_filtering
    update_board

    print_summary

    print_header "All Tests Passed! 🎉"
    print_success "Integration test suite completed successfully"
}

# Run main
main
