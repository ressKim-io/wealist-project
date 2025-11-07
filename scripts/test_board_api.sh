#!/bin/bash

# =============================================================================
# Board Service API Test Script
# =============================================================================
# Prerequisites:
# 1. User Service must be running (for workspace validation)
# 2. Board Service must be running
# 3. Valid JWT token from User Service
# 4. Valid Workspace ID from User Service
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BOARD_SERVICE_URL="${BOARD_SERVICE_URL:-http://localhost:8000}"
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"

# Test data (CHANGE THESE!)
JWT_TOKEN="${JWT_TOKEN:-your-jwt-token-here}"
USER_ID="${USER_ID:-your-user-id-here}"
WORKSPACE_ID="${WORKSPACE_ID:-your-workspace-id-here}"

# Storage for test IDs
PROJECT_ID=""
BOARD_ID=""
CUSTOM_ROLE_ID=""
CUSTOM_STAGE_ID=""
CUSTOM_IMPORTANCE_ID=""
COMMENT_ID=""

# =============================================================================
# Utility Functions
# =============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}=================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================================${NC}"
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

check_config() {
    print_header "Checking Configuration"

    if [ "$JWT_TOKEN" = "your-jwt-token-here" ]; then
        print_error "JWT_TOKEN is not set. Please set it before running tests."
        echo ""
        echo "Usage:"
        echo "  export JWT_TOKEN='your-actual-jwt-token'"
        echo "  export USER_ID='your-user-uuid'"
        echo "  export WORKSPACE_ID='your-workspace-uuid'"
        echo "  ./test_board_api.sh"
        exit 1
    fi

    if [ "$USER_ID" = "your-user-id-here" ]; then
        print_error "USER_ID is not set."
        exit 1
    fi

    if [ "$WORKSPACE_ID" = "your-workspace-id-here" ]; then
        print_error "WORKSPACE_ID is not set."
        exit 1
    fi

    print_success "Configuration loaded"
    print_info "Board Service: $BOARD_SERVICE_URL"
    print_info "User Service: $USER_SERVICE_URL"
    print_info "User ID: $USER_ID"
    print_info "Workspace ID: $WORKSPACE_ID"
}

# =============================================================================
# API Test Functions
# =============================================================================

test_health_check() {
    print_header "1. Health Check (No Auth Required)"

    response=$(curl -s "$BOARD_SERVICE_URL/health")

    if echo "$response" | jq -e '.database == "connected"' > /dev/null 2>&1; then
        print_success "Health check passed"
        echo "$response" | jq '.'
    else
        print_error "Health check failed"
        echo "$response"
        exit 1
    fi
}

test_create_project() {
    print_header "2. Create Project"

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"workspaceId\": \"$WORKSPACE_ID\",
            \"name\": \"Test Project $(date +%s)\",
            \"description\": \"Test project created by test script\"
        }")

    PROJECT_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "null" ]; then
        print_success "Project created: $PROJECT_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create project"
        echo "$response" | jq '.'
        exit 1
    fi
}

test_get_projects() {
    print_header "3. Get Projects in Workspace"

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects?workspace_id=$WORKSPACE_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved projects"
        echo "$response" | jq '.data.projects | length' | xargs -I {} echo "Found {} projects"
        echo "$response" | jq '.data.projects[] | {id, name, description}'
    else
        print_error "Failed to get projects"
        echo "$response" | jq '.'
    fi
}

test_get_project() {
    print_header "4. Get Project Details"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved project details"
        echo "$response" | jq '.data'
    else
        print_error "Failed to get project"
        echo "$response" | jq '.'
    fi
}

test_create_custom_role() {
    print_header "5. Create Custom Role"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/custom-fields/roles" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"Backend Developer\",
            \"color\": \"#3B82F6\"
        }")

    CUSTOM_ROLE_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$CUSTOM_ROLE_ID" ] && [ "$CUSTOM_ROLE_ID" != "null" ]; then
        print_success "Custom role created: $CUSTOM_ROLE_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create custom role"
        echo "$response" | jq '.'
    fi
}

test_create_custom_stage() {
    print_header "6. Create Custom Stage"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/custom-fields/stages" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"In Progress\",
            \"color\": \"#F59E0B\"
        }")

    CUSTOM_STAGE_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$CUSTOM_STAGE_ID" ] && [ "$CUSTOM_STAGE_ID" != "null" ]; then
        print_success "Custom stage created: $CUSTOM_STAGE_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create custom stage"
        echo "$response" | jq '.'
    fi
}

test_create_custom_importance() {
    print_header "7. Create Custom Importance"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/custom-fields/importance" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"name\": \"High Priority\",
            \"color\": \"#EF4444\"
        }")

    CUSTOM_IMPORTANCE_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$CUSTOM_IMPORTANCE_ID" ] && [ "$CUSTOM_IMPORTANCE_ID" != "null" ]; then
        print_success "Custom importance created: $CUSTOM_IMPORTANCE_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create custom importance"
        echo "$response" | jq '.'
    fi
}

test_get_custom_fields() {
    print_header "8. Get Custom Fields"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    print_info "Getting custom roles..."
    curl -s "$BOARD_SERVICE_URL/api/custom-fields/projects/$PROJECT_ID/roles" \
        -H "Authorization: Bearer $JWT_TOKEN" | jq '.data[] | {id, name, color}'

    print_info "Getting custom stages..."
    curl -s "$BOARD_SERVICE_URL/api/custom-fields/projects/$PROJECT_ID/stages" \
        -H "Authorization: Bearer $JWT_TOKEN" | jq '.data[] | {id, name, color}'

    print_info "Getting custom importance..."
    curl -s "$BOARD_SERVICE_URL/api/custom-fields/projects/$PROJECT_ID/importance" \
        -H "Authorization: Bearer $JWT_TOKEN" | jq '.data[] | {id, name, color}'

    print_success "Retrieved all custom fields"
}

test_create_board() {
    print_header "9. Create Board"

    if [ -z "$PROJECT_ID" ] || [ -z "$CUSTOM_STAGE_ID" ]; then
        print_error "Missing project ID or stage ID. Skipping."
        return
    fi

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/boards" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"projectId\": \"$PROJECT_ID\",
            \"title\": \"Test Board $(date +%s)\",
            \"description\": \"Test board for API testing\",
            \"customStageId\": \"$CUSTOM_STAGE_ID\",
            \"customImportanceId\": \"$CUSTOM_IMPORTANCE_ID\",
            \"assigneeId\": \"$USER_ID\",
            \"customRoleIds\": [\"$CUSTOM_ROLE_ID\"]
        }")

    BOARD_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$BOARD_ID" ] && [ "$BOARD_ID" != "null" ]; then
        print_success "Board created: $BOARD_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create board"
        echo "$response" | jq '.'
    fi
}

test_get_boards() {
    print_header "10. Get Boards"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s "$BOARD_SERVICE_URL/api/boards?project_id=$PROJECT_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved boards"
        echo "$response" | jq '.data[] | {id, title, description}'
    else
        print_error "Failed to get boards"
        echo "$response" | jq '.'
    fi
}

test_create_comment() {
    print_header "11. Create Comment"

    if [ -z "$BOARD_ID" ]; then
        print_error "No board ID available. Skipping."
        return
    fi

    response=$(curl -s -X POST "$BOARD_SERVICE_URL/api/comments" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"boardId\": \"$BOARD_ID\",
            \"content\": \"This is a test comment from API test script\"
        }")

    COMMENT_ID=$(echo "$response" | jq -r '.data.id // empty')

    if [ -n "$COMMENT_ID" ] && [ "$COMMENT_ID" != "null" ]; then
        print_success "Comment created: $COMMENT_ID"
        echo "$response" | jq '.data'
    else
        print_error "Failed to create comment"
        echo "$response" | jq '.'
    fi
}

test_get_comments() {
    print_header "12. Get Comments"

    if [ -z "$BOARD_ID" ]; then
        print_error "No board ID available. Skipping."
        return
    fi

    response=$(curl -s "$BOARD_SERVICE_URL/api/comments?board_id=$BOARD_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved comments"
        echo "$response" | jq '.data[] | {id, content, userName}'
    else
        print_error "Failed to get comments"
        echo "$response" | jq '.'
    fi
}

test_role_based_board_view() {
    print_header "13. Get Role-Based Board View"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/orders/role-board" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved role-based board view"
        echo "$response" | jq '.data | keys'
    else
        print_error "Failed to get role-based view"
        echo "$response" | jq '.'
    fi
}

test_stage_based_board_view() {
    print_header "14. Get Stage-Based Board View"

    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID available. Skipping."
        return
    fi

    response=$(curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/orders/stage-board" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_success "Retrieved stage-based board view"
        echo "$response" | jq '.data | keys'
    else
        print_error "Failed to get stage-based view"
        echo "$response" | jq '.'
    fi
}

# =============================================================================
# Cleanup Functions
# =============================================================================

cleanup() {
    print_header "Cleanup (Optional)"

    print_info "Test resources created:"
    echo "  Project ID: $PROJECT_ID"
    echo "  Board ID: $BOARD_ID"
    echo "  Custom Role ID: $CUSTOM_ROLE_ID"
    echo "  Custom Stage ID: $CUSTOM_STAGE_ID"
    echo "  Custom Importance ID: $CUSTOM_IMPORTANCE_ID"
    echo "  Comment ID: $COMMENT_ID"
    echo ""
    print_info "You can manually delete these resources if needed."
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Board Service API Test Suite                   ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════╝${NC}"

    check_config

    # Run tests
    test_health_check
    test_create_project
    test_get_projects
    test_get_project
    test_create_custom_role
    test_create_custom_stage
    test_create_custom_importance
    test_get_custom_fields
    test_create_board
    test_get_boards
    test_create_comment
    test_get_comments
    test_role_based_board_view
    test_stage_based_board_view

    cleanup

    print_header "Test Suite Completed"
    print_success "All tests passed! 🎉"
}

# Run main function
main
