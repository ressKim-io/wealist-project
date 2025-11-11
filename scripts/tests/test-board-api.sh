#!/bin/bash

# ============================================================================
# Board Service API Integration Test Script
# ============================================================================
# This script tests the entire Board Service API flow:
# 1. Get test user token from User Service
# 2. Create workspace in User Service
# 3. Test all Board Service endpoints
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API Base URLs
USER_SERVICE_URL="http://localhost:8080"
BOARD_SERVICE_URL="http://localhost:8000"

# Test results
PASSED=0
FAILED=0

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "\n${BLUE}========================================${NC}" >&2
    echo -e "${BLUE}$1${NC}" >&2
    echo -e "${BLUE}========================================${NC}" >&2
}

print_step() {
    echo -e "\n${YELLOW}>>> $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}" >&2
    PASSED=$((PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}" >&2
    FAILED=$((FAILED + 1))
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}" >&2
}

# Test API call
test_api() {
    local method=$1
    local url=$2
    local description=$3
    local data=$4
    local expected_status=${5:-200}

    print_step "Testing: $description"

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_status" ]; then
        print_success "$method $url - Status: $http_code"
        # Print formatted JSON to stderr (for display only)
        # Skip JSON formatting for empty responses (like 204 No Content)
        if [ -n "$body" ]; then
            if echo "$body" | jq '.' >/dev/null 2>&1; then
                echo "$body" | jq '.' >&2
            else
                print_info "Response is not valid JSON:" >&2
                echo "$body" >&2
            fi
        fi
        # Return raw body to stdout (for capture)
        echo "$body"
    else
        print_error "$method $url - Expected: $expected_status, Got: $http_code"
        print_info "Response body:"
        echo "$body" >&2
        return 1
    fi
}

# ============================================================================
# Main Test Flow
# ============================================================================

print_header "Board Service API Integration Test"

# ============================================================================
# Step 1: Get Test User Token
# ============================================================================
print_header "Step 1: Get Test User Token from User Service"

print_step "Calling /api/auth/test to get test user token"
response=$(curl -s "${USER_SERVICE_URL}/api/auth/test")

ACCESS_TOKEN=$(echo "$response" | jq -r '.accessToken')
USER_ID=$(echo "$response" | jq -r '.userId')
USER_EMAIL=$(echo "$response" | jq -r '.email')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
    print_error "Failed to get access token from User Service"
    echo "$response" >&2
    exit 1
fi

print_success "Got access token for user: $USER_EMAIL"
print_info "User ID: $USER_ID"
print_info "Token: ${ACCESS_TOKEN:0:20}..."

# ============================================================================
# Step 2: Create Workspace in User Service
# ============================================================================
print_header "Step 2: Create Workspace in User Service"

WORKSPACE_NAME="Test Workspace $(date +%s)"
WORKSPACE_DESC="Workspace for Board Service API testing"

print_step "Creating workspace: $WORKSPACE_NAME"
workspace_response=$(curl -s -X POST "${USER_SERVICE_URL}/api/workspaces" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"$WORKSPACE_NAME\",
        \"description\": \"$WORKSPACE_DESC\"
    }")

WORKSPACE_ID=$(echo "$workspace_response" | jq -r '.id')

if [ -z "$WORKSPACE_ID" ] || [ "$WORKSPACE_ID" = "null" ]; then
    print_error "Failed to create workspace"
    echo "$workspace_response" >&2
    exit 1
fi

print_success "Created workspace: $WORKSPACE_ID"
echo "$workspace_response" | jq '.' >&2

# ============================================================================
# Step 3: Test Board Service - Project APIs
# ============================================================================
print_header "Step 3: Test Project APIs"

# Create Project
PROJECT_NAME="Test Project $(date +%s)"
project_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/projects" \
    "Create Project" \
    "{
        \"workspace_id\": \"$WORKSPACE_ID\",
        \"name\": \"$PROJECT_NAME\",
        \"description\": \"Test project for API testing\"
    }" \
    201)

PROJECT_ID=$(echo "$project_data" | jq -r '.data.project_id' 2>/dev/null)

if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "null" ]; then
    print_error "Failed to get project ID from response"
    echo "$project_data" >&2
    exit 1
fi

print_info "Project ID: $PROJECT_ID"

# Get Project
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}" \
    "Get Project Details" \
    "" \
    200

# Get Projects by Workspace
test_api "GET" "${BOARD_SERVICE_URL}/api/projects?workspace_id=${WORKSPACE_ID}" \
    "Get Projects in Workspace" \
    "" \
    200

# Update Project
test_api "PUT" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}" \
    "Update Project" \
    "{
        \"name\": \"Updated Project Name\",
        \"description\": \"Updated description\"
    }" \
    200

# Search Projects
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/search?workspace_id=${WORKSPACE_ID}&query=Test" \
    "Search Projects" \
    "" \
    200

# Get Project Members
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/members" \
    "Get Project Members" \
    "" \
    200

# ============================================================================
# Step 4: Test Custom Fields APIs
# ============================================================================
print_header "Step 4: Test Custom Fields APIs"

# Get Custom Roles (should have defaults)
roles_response=$(test_api "GET" "${BOARD_SERVICE_URL}/api/custom-fields/projects/${PROJECT_ID}/roles" \
    "Get Custom Roles" \
    "" \
    200)

ROLE_ID=$(echo "$roles_response" | jq -r '.data[0].role_id' 2>/dev/null)
print_info "Default Role ID: $ROLE_ID"

# Create Custom Role
custom_role_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/custom-fields/roles" \
    "Create Custom Role" \
    "{
        \"project_id\": \"$PROJECT_ID\",
        \"name\": \"Frontend Developer\",
        \"color\": \"#3B82F6\"
    }" \
    201)

CUSTOM_ROLE_ID=$(echo "$custom_role_data" | jq -r '.data.role_id' 2>/dev/null)
print_info "Custom Role ID: $CUSTOM_ROLE_ID"

# Get Custom Role
test_api "GET" "${BOARD_SERVICE_URL}/api/custom-fields/roles/${CUSTOM_ROLE_ID}" \
    "Get Custom Role" \
    "" \
    200

# Update Custom Role
test_api "PUT" "${BOARD_SERVICE_URL}/api/custom-fields/roles/${CUSTOM_ROLE_ID}" \
    "Update Custom Role" \
    "{
        \"name\": \"Senior Frontend Developer\",
        \"color\": \"#10B981\"
    }" \
    200

# Get Custom Stages (should have defaults)
stages_response=$(test_api "GET" "${BOARD_SERVICE_URL}/api/custom-fields/projects/${PROJECT_ID}/stages" \
    "Get Custom Stages" \
    "" \
    200)

STAGE_ID=$(echo "$stages_response" | jq -r '.data[0].stage_id' 2>/dev/null)
print_info "Default Stage ID: $STAGE_ID"

# Create Custom Stage
custom_stage_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/custom-fields/stages" \
    "Create Custom Stage" \
    "{
        \"project_id\": \"$PROJECT_ID\",
        \"name\": \"Code Review\",
        \"color\": \"#F59E0B\"
    }" \
    201)

CUSTOM_STAGE_ID=$(echo "$custom_stage_data" | jq -r '.data.stage_id' 2>/dev/null)
print_info "Custom Stage ID: $CUSTOM_STAGE_ID"

# Get Custom Importance (should have defaults)
importance_response=$(test_api "GET" "${BOARD_SERVICE_URL}/api/custom-fields/projects/${PROJECT_ID}/importance" \
    "Get Custom Importance" \
    "" \
    200)

IMPORTANCE_ID=$(echo "$importance_response" | jq -r '.data[0].importance_id' 2>/dev/null)
print_info "Default Importance ID: $IMPORTANCE_ID"

# Create Custom Importance
custom_importance_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/custom-fields/importance" \
    "Create Custom Importance" \
    "{
        \"project_id\": \"$PROJECT_ID\",
        \"name\": \"Critical\",
        \"color\": \"#DC2626\",
        \"level\": 5
    }" \
    201)

CUSTOM_IMPORTANCE_ID=$(echo "$custom_importance_data" | jq -r '.data.importance_id' 2>/dev/null)
print_info "Custom Importance ID: $CUSTOM_IMPORTANCE_ID"

# ============================================================================
# Step 5: Test Board APIs
# ============================================================================
print_header "Step 5: Test Board APIs"

# Create Board
board_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/boards" \
    "Create Board" \
    "{
        \"project_id\": \"$PROJECT_ID\",
        \"title\": \"Implement Authentication\",
        \"content\": \"Add JWT authentication to the API\",
        \"role_ids\": [\"$ROLE_ID\"],
        \"stage_id\": \"$STAGE_ID\",
        \"importance_id\": \"$IMPORTANCE_ID\"
    }" \
    201)

BOARD_ID=$(echo "$board_data" | jq -r '.data.board_id' 2>/dev/null)

if [ -z "$BOARD_ID" ] || [ "$BOARD_ID" = "null" ]; then
    print_error "Failed to get board ID from response"
    echo "$board_data" >&2
    exit 1
fi

print_info "Board ID: $BOARD_ID"

# Get Board
test_api "GET" "${BOARD_SERVICE_URL}/api/boards/${BOARD_ID}" \
    "Get Board Details" \
    "" \
    200

# Get Boards by Project
test_api "GET" "${BOARD_SERVICE_URL}/api/boards?project_id=${PROJECT_ID}" \
    "Get Boards in Project" \
    "" \
    200

# Update Board
test_api "PUT" "${BOARD_SERVICE_URL}/api/boards/${BOARD_ID}" \
    "Update Board" \
    "{
        \"title\": \"Implement JWT Authentication\",
        \"content\": \"Add JWT authentication and authorization\",
        \"stage_id\": \"$CUSTOM_STAGE_ID\"
    }" \
    200

# ============================================================================
# Step 6: Test Comment APIs
# ============================================================================
print_header "Step 6: Test Comment APIs"

# Create Comment
comment_data=$(test_api "POST" "${BOARD_SERVICE_URL}/api/comments" \
    "Create Comment" \
    "{
        \"board_id\": \"$BOARD_ID\",
        \"content\": \"This is a test comment for the board\"
    }" \
    201)

COMMENT_ID=$(echo "$comment_data" | jq -r '.data.comment_id' 2>/dev/null)

if [ -z "$COMMENT_ID" ] || [ "$COMMENT_ID" = "null" ]; then
    print_error "Failed to get comment ID from response"
else
    print_info "Comment ID: $COMMENT_ID"

    # Get Comments by Board
    test_api "GET" "${BOARD_SERVICE_URL}/api/comments?board_id=${BOARD_ID}" \
        "Get Comments for Board" \
        "" \
        200

    # Update Comment
    test_api "PUT" "${BOARD_SERVICE_URL}/api/comments/${COMMENT_ID}" \
        "Update Comment" \
        "{
            \"content\": \"Updated comment content\"
        }" \
        200

    # Delete Comment
    test_api "DELETE" "${BOARD_SERVICE_URL}/api/comments/${COMMENT_ID}" \
        "Delete Comment" \
        "" \
        204
fi

# ============================================================================
# Step 7: Test User Order APIs (Drag & Drop)
# ============================================================================
print_header "Step 7: Test User Order APIs"

# Get Role-Based Board View
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/orders/role-board" \
    "Get Role-Based Board View" \
    "" \
    200

# Get Stage-Based Board View
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/orders/stage-board" \
    "Get Stage-Based Board View" \
    "" \
    200

# Update Role Column Order
test_api "PUT" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/orders/role-columns" \
    "Update Role Column Order" \
    "{
        \"itemIds\": [\"$ROLE_ID\", \"$CUSTOM_ROLE_ID\"]
    }" \
    200

# Update Stage Column Order
test_api "PUT" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/orders/stage-columns" \
    "Update Stage Column Order" \
    "{
        \"itemIds\": [\"$STAGE_ID\", \"$CUSTOM_STAGE_ID\"]
    }" \
    200

# Update Board Order in Role
if [ -n "$BOARD_ID" ] && [ "$BOARD_ID" != "null" ]; then
    test_api "PUT" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/orders/role-boards/${ROLE_ID}" \
        "Update Board Order in Role" \
        "{
            \"itemIds\": [\"$BOARD_ID\"]
        }" \
        200
fi

# ============================================================================
# Step 8: Test Project Join Request APIs
# ============================================================================
print_header "Step 8: Test Join Request APIs"

# Note: Since we're the owner, we can't create a join request for ourselves
# But we can test getting join requests
test_api "GET" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}/join-requests" \
    "Get Join Requests" \
    "" \
    200

# ============================================================================
# Step 9: Cleanup (Delete Board)
# ============================================================================
print_header "Step 9: Cleanup"

# Delete Board
test_api "DELETE" "${BOARD_SERVICE_URL}/api/boards/${BOARD_ID}" \
    "Delete Board" \
    "" \
    200

# Delete Custom Fields
test_api "DELETE" "${BOARD_SERVICE_URL}/api/custom-fields/roles/${CUSTOM_ROLE_ID}" \
    "Delete Custom Role" \
    "" \
    200

test_api "DELETE" "${BOARD_SERVICE_URL}/api/custom-fields/stages/${CUSTOM_STAGE_ID}" \
    "Delete Custom Stage" \
    "" \
    200

test_api "DELETE" "${BOARD_SERVICE_URL}/api/custom-fields/importance/${CUSTOM_IMPORTANCE_ID}" \
    "Delete Custom Importance" \
    "" \
    200

# Delete Project (soft delete)
test_api "DELETE" "${BOARD_SERVICE_URL}/api/projects/${PROJECT_ID}" \
    "Delete Project" \
    "" \
    200

# ============================================================================
# Test Summary
# ============================================================================
print_header "Test Summary"

TOTAL=$((PASSED + FAILED))
echo -e "${BLUE}Total Tests: $TOTAL${NC}" >&2
echo -e "${GREEN}Passed: $PASSED${NC}" >&2
echo -e "${RED}Failed: $FAILED${NC}" >&2

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}========================================${NC}" >&2
    echo -e "${GREEN}🎉 All tests passed!${NC}" >&2
    echo -e "${GREEN}========================================${NC}" >&2
    exit 0
else
    echo -e "\n${RED}========================================${NC}" >&2
    echo -e "${RED}❌ Some tests failed!${NC}" >&2
    echo -e "${RED}========================================${NC}" >&2
    exit 1
fi
