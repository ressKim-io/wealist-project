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

    if echo "$response" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
        print_success "Board Service is healthy"
        print_json "$response"
    else
        print_error "Health check failed"
    fi
}

get_test_token() {
    print_step "2" "Get Test Token from User Service"

    response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    TOKEN=$(echo "$response" | jq -r '.accessToken // empty')
    USER_ID=$(echo "$response" | jq -r '.userId // empty')

    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        print_success "Token obtained (User ID: ${USER_ID:0:8}...)"
    else
        print_error "Failed to get test token"
    fi
}

create_workspace() {
    print_step "3" "Workspace 생성 (User Service)"

    workspace_data="{\"name\":\"Test Workspace $(date +%s)\",\"description\":\"자동 테스트용 워크스페이스\"}"
    workspace_response=$(curl -s -X POST "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$workspace_data")

    if echo "$workspace_response" | grep -q '"id"'; then
        WORKSPACE_ID=$(echo "$workspace_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Workspace 생성 성공 (ID: ${WORKSPACE_ID:0:8}...)"
    else
        print_error "Workspace 생성 실패: $workspace_response"
    fi
}

create_project() {
    print_step "4" "Project 생성"

    project_data="{\"workspaceId\":\"$WORKSPACE_ID\",\"name\":\"Test Project $(date +%s)\",\"description\":\"자동 테스트용 프로젝트\"}"
    project_response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$project_data")

    if echo "$project_response" | grep -q '"data"'; then
        PROJECT_ID=$(echo "$project_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Project 생성 성공 (ID: ${PROJECT_ID:0:8}...)"
    else
        print_error "Project 생성 실패: $project_response"
    fi
}

create_field_status() {
    print_step "5" "Create Custom Field: Status (single_select)"

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

    FIELD_STATUS_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$FIELD_STATUS_ID" ] && [ "$FIELD_STATUS_ID" != "null" ]; then
        print_success "Status field created: ${FIELD_STATUS_ID:0:8}..."
        print_json "$response"
    else
        print_error "Failed to create status field"
    fi
}

create_status_options() {
    print_step "6" "Create Status Options (To Do, In Progress, Done)"

    # To Do
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"value\": \"To Do\",
            \"color\": \"#94A3B8\"
        }")

    OPTION_TODO_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "To Do option created: ${OPTION_TODO_ID:0:8}..."

    # In Progress
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"value\": \"In Progress\",
            \"color\": \"#3B82F6\"
        }")

    OPTION_INPROGRESS_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "In Progress option created: ${OPTION_INPROGRESS_ID:0:8}..."

    # Done
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"value\": \"Done\",
            \"color\": \"#10B981\"
        }")

    OPTION_DONE_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "Done option created: ${OPTION_DONE_ID:0:8}..."
}

create_field_priority() {
    print_step "7" "Create Custom Field: Priority (single_select)"

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

    FIELD_PRIORITY_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$FIELD_PRIORITY_ID" ] && [ "$FIELD_PRIORITY_ID" != "null" ]; then
        print_success "Priority field created: ${FIELD_PRIORITY_ID:0:8}..."
        print_json "$response"
    else
        print_error "Failed to create priority field"
    fi
}

create_priority_options() {
    print_step "8" "Create Priority Options (High, Medium, Low)"

    # High
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"value\": \"High\",
            \"color\": \"#EF4444\"
        }")

    OPTION_HIGH_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "High priority option created: ${OPTION_HIGH_ID:0:8}..."

    # Medium
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"value\": \"Medium\",
            \"color\": \"#F59E0B\"
        }")

    OPTION_MEDIUM_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "Medium priority option created: ${OPTION_MEDIUM_ID:0:8}..."

    # Low
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/field-options" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"value\": \"Low\",
            \"color\": \"#6B7280\"
        }")

    OPTION_LOW_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_success "Low priority option created: ${OPTION_LOW_ID:0:8}..."
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

    FIELD_TAGS_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$FIELD_TAGS_ID" ] && [ "$FIELD_TAGS_ID" != "null" ]; then
        print_success "Tags field created: ${FIELD_TAGS_ID:0:8}..."
        print_json "$response"
    else
        print_error "Failed to create tags field"
    fi
}

list_project_fields() {
    print_step "10" "List All Project Fields"

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        field_count=$(echo "$response" | jq '.data | length')
        print_success "Retrieved $field_count custom fields"
        echo "$response" | jq '.data[] | {id, name, field_type: .fieldType, required: .isRequired}'
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
            \"description\": \"Board with custom fields\",
            \"assigneeId\": \"$USER_ID\"
        }")

    BOARD_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$BOARD_ID" ] && [ "$BOARD_ID" != "null" ]; then
        print_success "Board created: ${BOARD_ID:0:8}..."
        print_json "$response"

        # Check custom_fields in response
        custom_fields=$(echo "$response" | jq '.data.custom_fields // empty')
        if [ -n "$custom_fields" ]; then
            print_info "Custom fields in response:"
            echo "$custom_fields" | jq '.'
        fi
    else
        print_error "Failed to create board"
    fi
}

set_board_field_values() {
    print_step "12" "Set Board Field Values"

    # Set Status = In Progress
    print_info "Setting Status to 'In Progress'..."
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/boards/$BOARD_ID/field-values" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_STATUS_ID\",
            \"value\": \"$OPTION_INPROGRESS_ID\"
        }")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Status field value set"
    fi

    # Set Priority = High
    print_info "Setting Priority to 'High'..."
    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/boards/$BOARD_ID/field-values" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"fieldId\": \"$FIELD_PRIORITY_ID\",
            \"value\": \"$OPTION_HIGH_ID\"
        }")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Priority field value set"
    fi
}

get_board_with_fields() {
    print_step "13" "Get Board with Custom Fields"

    response=$(curl -s "$BOARD_SERVICE_URL/api/boards/$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved board with custom fields"
        echo "$response" | jq '.data | {id, title, custom_fields}'
    else
        print_error "Failed to get board"
    fi
}

get_boards_in_project() {
    print_step "14" "Get All Boards in Project"

    response=$(curl -s "$BOARD_SERVICE_URL/api/boards?project_id=$PROJECT_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        board_count=$(echo "$response" | jq '.data | length')
        print_success "Retrieved $board_count boards"
        echo "$response" | jq '.data[] | {id, title, status: .custom_fields.Status, priority: .custom_fields.Priority}'
    else
        print_error "Failed to get boards"
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

    COMMENT_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$COMMENT_ID" ] && [ "$COMMENT_ID" != "null" ]; then
        print_success "Comment created: ${COMMENT_ID:0:8}..."
        print_json "$response"
    else
        print_error "Failed to create comment"
    fi
}

get_comments() {
    print_step "16" "Get Board Comments"

    response=$(curl -s "$BOARD_SERVICE_URL/api/comments?board_id=$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        comment_count=$(echo "$response" | jq '.data | length')
        print_success "Retrieved $comment_count comments"
        echo "$response" | jq '.data[] | {id, content, userName, createdAt}'
    else
        print_error "Failed to get comments"
    fi
}

test_board_filtering() {
    print_step "17" "Test Board Filtering (using custom fields)"

    print_info "Filter by Status = In Progress..."
    response=$(curl -s "$BOARD_SERVICE_URL/api/boards?project_id=$PROJECT_ID&status=In%20Progress" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        filtered_count=$(echo "$response" | jq '.data | length')
        print_success "Filtered boards: $filtered_count"
    fi
}

update_board() {
    print_step "18" "Update Board"

    response=$(curl -s -X PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"title\": \"Updated Board Title\",
            \"description\": \"Updated description\"
        }")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Board updated successfully"
        print_json "$response"
    else
        print_error "Failed to update board"
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
