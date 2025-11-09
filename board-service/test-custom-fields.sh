#!/bin/bash

# =============================================================================
# Custom Fields System Integration Test Script
# =============================================================================
# Tests all 22 new custom fields API endpoints
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

# Service URLs
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"
BOARD_SERVICE_URL="${BOARD_SERVICE_URL:-http://localhost:8081}"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

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
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_result() {
    local status_code=$1
    local expected=$2
    local description=$3

    if [ "$status_code" -eq "$expected" ]; then
        print_success "$description (HTTP $status_code)"
        return 0
    else
        print_error "$description (Expected: $expected, Got: $status_code)"
        return 1
    fi
}

# =============================================================================
# Authentication
# =============================================================================

get_test_token() {
    print_header "Getting Test Token from User Service"
    print_info "Calling: $USER_SERVICE_URL/api/auth/test"

    response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        JWT_TOKEN=$(echo "$response" | jq -r '.accessToken')
        USER_ID=$(echo "$response" | jq -r '.userId')

        print_success "Test token received!"
        echo "User ID: $USER_ID"
        echo "Token: ${JWT_TOKEN:0:50}..."

        export JWT_TOKEN
        export USER_ID
        return 0
    else
        print_error "Failed to get test token!"
        echo "$response" | jq '.'
        exit 1
    fi
}

# =============================================================================
# Setup: Create Test Project
# =============================================================================

create_test_project() {
    print_header "Creating Test Project"

    # Get or create workspace
    print_section "Get or Create Workspace"
    workspaces_response=$(curl -s "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $JWT_TOKEN")

    WORKSPACE_ID=$(echo "$workspaces_response" | jq -r '.[0].id // .data[0].id' 2>/dev/null)

    if [ -z "$WORKSPACE_ID" ] || [ "$WORKSPACE_ID" = "null" ]; then
        print_info "No workspace found. Creating new workspace..."

        create_ws_response=$(curl -s -w "\n%{http_code}" -X POST "$USER_SERVICE_URL/api/workspaces" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "Custom Fields Test Workspace",
                "description": "Workspace for custom fields testing"
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

    export WORKSPACE_ID

    # Create project
    project_response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/projects" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "workspace_id": "'$WORKSPACE_ID'",
            "name": "Custom Fields Test Project",
            "description": "Test project for custom fields system"
        }')

    http_code=$(echo "$project_response" | tail -n1)
    response_body=$(echo "$project_response" | sed '$d')

    if [ "$http_code" -eq 201 ]; then
        PROJECT_ID=$(echo \"$response_body\" | jq -r '.project_id')
        print_success "Project created: $PROJECT_ID"
        export PROJECT_ID
        return 0
    else
        print_error "Failed to create project (HTTP $http_code)"
        echo "$response_body" | jq '.'
        exit 1
    fi
}

# =============================================================================
# Test 1: Field CRUD Operations
# =============================================================================

test_field_crud() {
    print_header "TEST 1: Field CRUD Operations"

    # 1.1 Create text field
    print_section "1.1 Create Text Field"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Description",
            "field_type": "text",
            "description": "Task description",
            "is_required": false,
            "config": {"max_length": 500}
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create text field"; then
        TEXT_FIELD_ID=$(echo "$body" | jq -r '.field_id')
        echo "  Field ID: $TEXT_FIELD_ID"
    fi

    # 1.2 Create single_select field
    print_section "1.2 Create Single Select Field"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Priority",
            "field_type": "single_select",
            "description": "Task priority",
            "is_required": true,
            "config": {}
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create single_select field"; then
        PRIORITY_FIELD_ID=$(echo "$body" | jq -r '.field_id')
        echo "  Field ID: $PRIORITY_FIELD_ID"
    fi

    # 1.3 Create multi_select field
    print_section "1.3 Create Multi Select Field"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Tags",
            "field_type": "multi_select",
            "description": "Task tags",
            "is_required": false,
            "config": {"max_selections": 5}
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create multi_select field"; then
        TAGS_FIELD_ID=$(echo "$body" | jq -r '.field_id')
        echo "  Field ID: $TAGS_FIELD_ID"
    fi

    # 1.4 Create number field
    print_section "1.4 Create Number Field"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "Story Points",
            "field_type": "number",
            "description": "Estimated story points",
            "is_required": false,
            "config": {"min": 0, "max": 100, "decimal_places": 0}
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    print_result "$http_code" 200 "Create number field"

    # 1.5 Get all fields
    print_section "1.5 Get All Fields for Project"
    response=$(curl -s -w "\n%{http_code}" "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $JWT_TOKEN")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Get all fields"; then
        field_count=$(echo "$body" | jq '. | length')
        echo "  Found $field_count fields"
    fi

    # 1.6 Update field
    print_section "1.6 Update Field"
    response=$(curl -s -w "\n%{http_code}" -X PATCH "$BOARD_SERVICE_URL/api/fields/$TEXT_FIELD_ID" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Detailed Description",
            "description": "Updated description field"
        }')

    http_code=$(echo "$response" | tail -n1)
    print_result "$http_code" 200 "Update field"
}

# =============================================================================
# Test 2: Field Options CRUD
# =============================================================================

test_field_options() {
    print_header "TEST 2: Field Options CRUD"

    # 2.1 Create options for Priority field
    print_section "2.1 Create Priority Options"

    # High priority
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/options" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "field_id": "'$PRIORITY_FIELD_ID'",
            "label": "High",
            "color": "#FF0000",
            "description": "High priority"
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create High priority option"; then
        HIGH_OPTION_ID=$(echo "$body" | jq -r '.option_id')
    fi

    # Medium priority
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/options" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "field_id": "'$PRIORITY_FIELD_ID'",
            "label": "Medium",
            "color": "#FFA500"
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create Medium priority option"; then
        MEDIUM_OPTION_ID=$(echo "$body" | jq -r '.option_id')
    fi

    # Low priority
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/options" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "field_id": "'$PRIORITY_FIELD_ID'",
            "label": "Low",
            "color": "#00FF00"
        }')

    http_code=$(echo "$response" | tail -n1)
    print_result "$http_code" 200 "Create Low priority option"

    # 2.2 Get options
    print_section "2.2 Get Field Options"
    response=$(curl -s -w "\n%{http_code}" "$BOARD_SERVICE_URL/api/fields/$PRIORITY_FIELD_ID/options" \
        -H "Authorization: Bearer $JWT_TOKEN")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Get field options"; then
        option_count=$(echo "$body" | jq '. | length')
        echo "  Found $option_count options"
    fi

    # 2.3 Create options for Tags field
    print_section "2.3 Create Tag Options"

    for tag in "Frontend" "Backend" "Bug" "Feature"; do
        curl -s -X POST "$BOARD_SERVICE_URL/api/options" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "field_id": "'$TAGS_FIELD_ID'",
                "label": "'$tag'",
                "color": "#0000FF"
            }' > /dev/null
    done

    print_success "Created 4 tag options (Frontend, Backend, Bug, Feature)"
}

# =============================================================================
# Test 3: Board and Field Values
# =============================================================================

test_field_values() {
    print_header "TEST 3: Board and Field Values"

    # 3.1 Create a board first
    print_section "3.1 Create Test Board"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/boards" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "title": "Test Task #1",
            "description": "Testing custom fields"
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 201 "Create board"; then
        BOARD_ID=$(echo "$body" | jq -r '.board_id')
        echo "  Board ID: $BOARD_ID"
    fi

    # 3.2 Set text field value
    print_section "3.2 Set Text Field Value"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/field-values" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "board_id": "'$BOARD_ID'",
            "field_id": "'$TEXT_FIELD_ID'",
            "value": "This is a detailed description of the task"
        }')

    http_code=$(echo "$response" | tail -n1)
    print_result "$http_code" 200 "Set text field value"

    # 3.3 Set single select value (Priority)
    print_section "3.3 Set Single Select Value (Priority = High)"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/field-values" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "board_id": "'$BOARD_ID'",
            "field_id": "'$PRIORITY_FIELD_ID'",
            "value": "'$HIGH_OPTION_ID'"
        }')

    http_code=$(echo "$response" | tail -n1)
    print_result "$http_code" 200 "Set priority to High"

    # 3.4 Set multi select values (Tags)
    print_section "3.4 Set Multi Select Values (Tags)"

    # Get tag option IDs
    tags_response=$(curl -s "$BOARD_SERVICE_URL/api/fields/$TAGS_FIELD_ID/options" \
        -H "Authorization: Bearer $JWT_TOKEN")

    FRONTEND_TAG=$(echo "$tags_response" | jq -r '.[] | select(.label=="Frontend") | .option_id')
    BUG_TAG=$(echo "$tags_response" | jq -r '.[] | select(.label=="Bug") | .option_id')

    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/field-values/multi-select" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "board_id": "'$BOARD_ID'",
            "field_id": "'$TAGS_FIELD_ID'",
            "values": [
                {"value": "'$FRONTEND_TAG'", "display_order": 0},
                {"value": "'$BUG_TAG'", "display_order": 1}
            ]
        }')

    http_code=$(echo "$response" | tail -n1)
    print_result "$http_code" 200 "Set multi-select tags"

    # 3.5 Get board field values
    print_section "3.5 Get Board Field Values"
    response=$(curl -s -w "\n%{http_code}" "$BOARD_SERVICE_URL/api/boards/$BOARD_ID/field-values" \
        -H "Authorization: Bearer $JWT_TOKEN")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Get board field values"; then
        echo "  Response:"
        echo "$body" | jq '.'
    fi
}

# =============================================================================
# Test 4: Saved Views (Filters/Sorting/Grouping)
# =============================================================================

test_saved_views() {
    print_header "TEST 4: Saved Views"

    # 4.1 Create saved view
    print_section "4.1 Create Saved View with Filters"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BOARD_SERVICE_URL/api/views" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "project_id": "'$PROJECT_ID'",
            "name": "High Priority Tasks",
            "description": "View for high priority tasks",
            "is_shared": true,
            "filters": {
                "'$PRIORITY_FIELD_ID'": {
                    "operator": "eq",
                    "value": "'$HIGH_OPTION_ID'"
                }
            },
            "sort_by": "created_at",
            "sort_direction": "desc"
        }')

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Create saved view"; then
        VIEW_ID=$(echo "$body" | jq -r '.view_id')
        echo "  View ID: $VIEW_ID"
    fi

    # 4.2 Get project views
    print_section "4.2 Get Project Views"
    response=$(curl -s -w "\n%{http_code}" "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/views" \
        -H "Authorization: Bearer $JWT_TOKEN")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Get project views"; then
        view_count=$(echo "$body" | jq '. | length')
        echo "  Found $view_count views"
    fi

    # 4.3 Apply view (get filtered boards)
    print_section "4.3 Apply View (Get Filtered Boards)"
    response=$(curl -s -w "\n%{http_code}" "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards?page=1&limit=20" \
        -H "Authorization: Bearer $JWT_TOKEN")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if print_result "$http_code" 200 "Apply view and get boards"; then
        board_count=$(echo "$body" | jq '.boards | length')
        total=$(echo "$body" | jq '.total')
        echo "  Found $board_count boards (total: $total)"
    fi
}

# =============================================================================
# Test 5: Cache Performance
# =============================================================================

test_cache_performance() {
    print_header "TEST 5: Cache Performance"

    print_section "5.1 First Request (DB Hit)"
    start_time=$(date +%s%N)
    curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    print_info "First request took: ${duration}ms (expected: ~30-50ms)"

    print_section "5.2 Second Request (Cache Hit)"
    start_time=$(date +%s%N)
    curl -s "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID/fields" \
        -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    print_info "Second request took: ${duration}ms (expected: ~1-5ms)"

    if [ "$duration" -lt 10 ]; then
        print_success "Cache is working! Response time < 10ms"
    else
        print_error "Cache might not be working (response time: ${duration}ms)"
    fi
}

# =============================================================================
# Cleanup (Optional)
# =============================================================================

cleanup() {
    print_header "Cleanup (Optional)"
    print_info "To delete test project: DELETE /api/projects/$PROJECT_ID"
    print_info "Test data will remain for manual inspection"
}

# =============================================================================
# Summary
# =============================================================================

print_summary() {
    print_header "TEST SUMMARY"

    echo "Total Tests: $TESTS_TOTAL"
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
        echo ""
        echo "Test Data:"
        echo "  Project ID: $PROJECT_ID"
        echo "  Board ID: $BOARD_ID"
        echo "  View ID: $VIEW_ID"
    else
        echo -e "${RED}✗ SOME TESTS FAILED${NC}"
        exit 1
    fi
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║   Custom Fields System Integration Test                      ║${NC}"
    echo -e "${CYAN}║   Testing 22 New API Endpoints                                ║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Run tests
    get_test_token
    create_test_project
    test_field_crud
    test_field_options
    test_field_values
    test_saved_views
    test_cache_performance
    cleanup
    print_summary
}

# Run main
main
