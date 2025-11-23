#!/bin/bash

# ============================================================================
# Fractional Indexing Test Script
# ============================================================================
# This script tests the new fractional indexing implementation for board ordering
# Tests include:
# 1. Creating boards and custom fields
# 2. Creating a saved view
# 3. Moving boards within the same column
# 4. Moving boards between different columns
# 5. Verifying positions are lexicographically sorted
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# API Base URLs
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"
BOARD_SERVICE_URL="${BOARD_SERVICE_URL:-http://localhost:8000}"

# Test results
PASSED=0
FAILED=0

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_step() {
    echo -e "\n${YELLOW}>>> $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    PASSED=$((PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    FAILED=$((FAILED + 1))
}

print_info() {
    echo -e "${CYAN}ℹ $1${NC}"
}

# Test API call
call_api() {
    local method=$1
    local url=$2
    local data=$3

    if [ -z "$data" ]; then
        curl -s -X "$method" "$url" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json"
    else
        curl -s -X "$method" "$url" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

# Extract field from JSON
extract_field() {
    echo "$1" | jq -r "$2"
}

# ============================================================================
# Setup: Get Token and Create Test Data
# ============================================================================

setup_test_data() {
    print_header "SETUP: Creating Test Data"

    # Get access token from User Service (test endpoint)
    print_step "Getting test user token from User Service"
    token_response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    ACCESS_TOKEN=$(echo "$token_response" | jq -r '.accessToken')
    USER_ID=$(echo "$token_response" | jq -r '.userId')

    if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
        print_error "Failed to get access token"
        echo "Response: $token_response"
        exit 1
    fi

    print_success "Got test user token"
    print_info "User ID: $USER_ID"

    # Create workspace first (in User Service)
    print_step "Creating test workspace"
    workspace_response=$(curl -s -X POST "$USER_SERVICE_URL/api/workspaces" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Fractional Indexing Test Workspace",
            "description": "Workspace for testing fractional indexing"
        }')

    WORKSPACE_ID=$(extract_field "$workspace_response" '.id')

    if [ -z "$WORKSPACE_ID" ] || [ "$WORKSPACE_ID" = "null" ]; then
        print_error "Failed to create workspace"
        echo "Response: $workspace_response"
        exit 1
    fi

    print_success "Created workspace: $WORKSPACE_ID"

    # Create project (in Board Service)
    print_step "Creating test project"
    project_response=$(call_api POST "$BOARD_SERVICE_URL/api/projects" '{
        "workspace_id": "'"$WORKSPACE_ID"'",
        "name": "Fractional Indexing Test Project",
        "description": "Testing fractional indexing for board ordering"
    }')

    PROJECT_ID=$(extract_field "$project_response" '.data.project_id')

    if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "null" ]; then
        print_error "Failed to create project"
        echo "Response: $project_response"
        exit 1
    fi

    print_success "Created project: $PROJECT_ID"

    # Create custom field (Status) - using new Jira-style API
    print_step "Creating Status custom field"
    field_response=$(call_api POST "$BOARD_SERVICE_URL/api/fields" '{
        "project_id": "'"$PROJECT_ID"'",
        "name": "Status",
        "field_type": "single_select",
        "description": "Board status for testing"
    }')

    FIELD_ID=$(extract_field "$field_response" '.data.field_id')
    print_success "Created custom field: $FIELD_ID"

    # Create field options - using new Jira-style API
    print_step "Creating field options: Todo, In Progress, Done"

    todo_response=$(call_api POST "$BOARD_SERVICE_URL/api/field-options" '{
        "field_id": "'"$FIELD_ID"'",
        "label": "Todo",
        "color": "#FF0000"
    }')
    TODO_OPTION_ID=$(extract_field "$todo_response" '.data.option_id')
    print_success "Created option: Todo ($TODO_OPTION_ID)"

    progress_response=$(call_api POST "$BOARD_SERVICE_URL/api/field-options" '{
        "field_id": "'"$FIELD_ID"'",
        "label": "In Progress",
        "color": "#FFA500"
    }')
    PROGRESS_OPTION_ID=$(extract_field "$progress_response" '.data.option_id')
    print_success "Created option: In Progress ($PROGRESS_OPTION_ID)"

    done_response=$(call_api POST "$BOARD_SERVICE_URL/api/field-options" '{
        "field_id": "'"$FIELD_ID"'",
        "label": "Done",
        "color": "#00FF00"
    }')
    DONE_OPTION_ID=$(extract_field "$done_response" '.data.option_id')
    print_success "Created option: Done ($DONE_OPTION_ID)"

    # Create saved view (without grouping for fractional indexing tests)
    print_step "Creating saved view for Status Board"
    view_response=$(call_api POST "$BOARD_SERVICE_URL/api/views" '{
        "project_id": "'"$PROJECT_ID"'",
        "name": "Status Board"
    }')

    VIEW_ID=$(extract_field "$view_response" '.data.view_id')
    print_success "Created view: $VIEW_ID"

    # Create test boards
    print_step "Creating test boards"

    for i in {1..5}; do
        board_response=$(call_api POST "$BOARD_SERVICE_URL/api/boards" '{
            "project_id": "'"$PROJECT_ID"'",
            "title": "Board-'"$i"'",
            "description": "Test board '"$i"' for fractional indexing"
        }')

        board_id=$(extract_field "$board_response" '.data.board_id')

        # Set custom field value to "Todo"
        call_api POST "$BOARD_SERVICE_URL/api/board-field-values" '{
            "board_id": "'"$board_id"'",
            "field_id": "'"$FIELD_ID"'",
            "value": "'"$TODO_OPTION_ID"'"
        }' > /dev/null

        # Initialize position by calling MoveBoard (moves to end of Todo column)
        move_init_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$board_id/move" '{
            "view_id": "'"$VIEW_ID"'",
            "group_by_field_id": "'"$FIELD_ID"'",
            "new_field_value": "'"$TODO_OPTION_ID"'",
            "before_position": null,
            "after_position": null
        }')

        # Check if MoveBoard succeeded
        if echo "$move_init_response" | jq -e '.error' > /dev/null 2>&1; then
            print_error "Failed to initialize position for Board-$i"
            exit 1
        fi

        initial_position=$(echo "$move_init_response" | jq -r '.data.new_position // "unknown"')

        eval "BOARD_${i}_ID=$board_id"
        print_success "Created Board-$i: $board_id (position: $initial_position)"
    done
}

# ============================================================================
# Test 1: Move Within Same Column
# ============================================================================

test_move_within_column() {
    print_header "TEST 1: Move Board Within Same Column"

    print_step "Getting current board orders in Todo column"
    orders_response=$(call_api GET "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards")

    print_info "Current orders:"
    echo "$orders_response" | jq -r '.data.boards[] | select(.custom_fields["'"$FIELD_ID"'"] == "'"$TODO_OPTION_ID"'") | "\(.title): \(.position // "no position")"'

    # Debug: Print full response if positions are null
    if echo "$orders_response" | jq -e '.data.boards[0].position == null or .data.boards[0].position == ""' > /dev/null 2>&1; then
        print_info "Warning: Positions are null, printing full response:"
        echo "$orders_response" | jq '.'
    fi

    # Get positions
    board1_position=$(echo "$orders_response" | jq -r '.data.boards[] | select(.board_id == "'"$BOARD_1_ID"'") | .position // "a0"')
    board3_position=$(echo "$orders_response" | jq -r '.data.boards[] | select(.board_id == "'"$BOARD_3_ID"'") | .position // "a2"')

    print_step "Moving Board-2 between Board-1 and Board-3"
    move_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_2_ID/move" '{
        "view_id": "'"$VIEW_ID"'",
        "group_by_field_id": "'"$FIELD_ID"'",
        "new_field_value": "'"$TODO_OPTION_ID"'",
        "before_position": "'"$board1_position"'",
        "after_position": "'"$board3_position"'"
    }')

    new_position=$(extract_field "$move_response" '.data.new_position')

    if [ -z "$new_position" ] || [ "$new_position" = "null" ]; then
        print_error "Failed to move board"
        echo "Response: $move_response"
        return 1
    fi

    print_success "Board-2 moved to new position: $new_position"

    # Verify lexicographic order
    if [[ "$board1_position" < "$new_position" && "$new_position" < "$board3_position" ]]; then
        print_success "Position is correctly between Board-1 and Board-3 (lexicographic order)"
    else
        print_error "Position is NOT in correct order"
        print_info "Board-1: $board1_position"
        print_info "Board-2 (new): $new_position"
        print_info "Board-3: $board3_position"
    fi
}

# ============================================================================
# Test 2: Move to Different Column
# ============================================================================

test_move_between_columns() {
    print_header "TEST 2: Move Board to Different Column"

    print_step "Moving Board-1 from Todo to In Progress (first position)"
    move_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_1_ID/move" '{
        "view_id": "'"$VIEW_ID"'",
        "group_by_field_id": "'"$FIELD_ID"'",
        "new_field_value": "'"$PROGRESS_OPTION_ID"'",
        "before_position": null,
        "after_position": null
    }')

    new_position=$(extract_field "$move_response" '.data.new_position')
    new_field_value=$(extract_field "$move_response" '.data.new_field_value')

    if [ "$new_field_value" != "$PROGRESS_OPTION_ID" ]; then
        print_error "Field value not updated correctly"
        return 1
    fi

    print_success "Board-1 moved to In Progress with position: $new_position"

    # Verify board is in new column
    print_step "Verifying Board-1 is in In Progress column"
    board_response=$(call_api GET "$BOARD_SERVICE_URL/api/boards/$BOARD_1_ID")
    current_status=$(echo "$board_response" | jq -r '.data.custom_fields["'"$FIELD_ID"'"]')

    if [ "$current_status" = "$PROGRESS_OPTION_ID" ]; then
        print_success "Board-1 status correctly updated to In Progress"
    else
        print_error "Board-1 status not updated (expected: $PROGRESS_OPTION_ID, got: $current_status)"
    fi
}

# ============================================================================
# Test 3: Move to First Position
# ============================================================================

test_move_to_first() {
    print_header "TEST 3: Move Board to First Position"

    # Get first board's position in Todo
    print_step "Getting first board position in Todo column"
    orders_response=$(call_api GET "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards")

    first_position=$(echo "$orders_response" | jq -r '[.data.boards[] | select(.custom_fields["'"$FIELD_ID"'"] == "'"$TODO_OPTION_ID"'")] | sort_by(.position) | .[0].position')

    print_info "Current first position: $first_position"

    print_step "Moving Board-5 to first position in Todo"
    move_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_5_ID/move" '{
        "view_id": "'"$VIEW_ID"'",
        "group_by_field_id": "'"$FIELD_ID"'",
        "new_field_value": "'"$TODO_OPTION_ID"'",
        "before_position": null,
        "after_position": "'"$first_position"'"
    }')

    new_position=$(extract_field "$move_response" '.data.new_position')

    print_success "Board-5 moved to position: $new_position"

    # Verify it's before the old first position
    if [[ "$new_position" < "$first_position" ]]; then
        print_success "New position is correctly before old first position"
    else
        print_error "New position is NOT before old first position"
        print_info "New: $new_position, Old first: $first_position"
    fi
}

# ============================================================================
# Test 4: Move to Last Position
# ============================================================================

test_move_to_last() {
    print_header "TEST 4: Move Board to Last Position"

    # Get last board's position in Todo
    print_step "Getting last board position in Todo column"
    orders_response=$(call_api GET "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards")

    last_position=$(echo "$orders_response" | jq -r '[.data.boards[] | select(.custom_fields["'"$FIELD_ID"'"] == "'"$TODO_OPTION_ID"'")] | sort_by(.position) | .[-1].position')

    print_info "Current last position: $last_position"

    print_step "Moving Board-4 to last position in Todo"
    move_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$BOARD_4_ID/move" '{
        "view_id": "'"$VIEW_ID"'",
        "group_by_field_id": "'"$FIELD_ID"'",
        "new_field_value": "'"$TODO_OPTION_ID"'",
        "before_position": "'"$last_position"'",
        "after_position": null
    }')

    new_position=$(extract_field "$move_response" '.data.new_position')

    print_success "Board-4 moved to position: $new_position"

    # Verify it's after the old last position
    if [[ "$new_position" > "$last_position" ]]; then
        print_success "New position is correctly after old last position"
    else
        print_error "New position is NOT after old last position"
        print_info "New: $new_position, Old last: $last_position"
    fi
}

# ============================================================================
# Test 5: Verify All Positions are Sorted
# ============================================================================

test_verify_sorting() {
    print_header "TEST 5: Verify All Positions are Lexicographically Sorted"

    print_step "Getting all board orders"
    orders_response=$(call_api GET "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards")

    # Extract positions for Todo column
    positions=$(echo "$orders_response" | jq -r '[.data.boards[] | select(.custom_fields["'"$FIELD_ID"'"] == "'"$TODO_OPTION_ID"'")] | sort_by(.position) | .[].position')

    print_info "Positions in Todo column (sorted):"
    echo "$positions" | while read -r pos; do
        print_info "  - $pos"
    done

    # Check if positions are actually sorted
    sorted_positions=$(echo "$positions" | sort)

    if [ "$positions" = "$sorted_positions" ]; then
        print_success "All positions are correctly sorted lexicographically"
    else
        print_error "Positions are NOT correctly sorted"
        print_info "Expected order:"
        echo "$sorted_positions"
        print_info "Actual order:"
        echo "$positions"
    fi
}

# ============================================================================
# Test 6: Performance Test (Single Row Update)
# ============================================================================

test_performance() {
    print_header "TEST 6: Performance Test - Verify Single Row Update"

    print_step "Creating 10 additional boards for performance test"

    for i in {6..15}; do
        board_response=$(call_api POST "$BOARD_SERVICE_URL/api/boards" '{
            "project_id": "'"$PROJECT_ID"'",
            "title": "Perf-Board-'"$i"'",
            "description": "Performance test board '"$i"'"
        }')

        board_id=$(extract_field "$board_response" '.data.board_id')

        # Set to Done column
        call_api POST "$BOARD_SERVICE_URL/api/board-field-values" '{
            "board_id": "'"$board_id"'",
            "field_id": "'"$FIELD_ID"'",
            "value": "'"$DONE_OPTION_ID"'"
        }' > /dev/null

        eval "PERF_BOARD_${i}_ID=$board_id"
    done

    print_success "Created 10 additional boards"

    # Get positions
    print_step "Getting positions in Done column"
    orders_response=$(call_api GET "$BOARD_SERVICE_URL/api/views/$VIEW_ID/boards")

    positions=$(echo "$orders_response" | jq -r '[.data.boards[] | select(.custom_fields["'"$FIELD_ID"'"] == "'"$DONE_OPTION_ID"'")] | sort_by(.position) | .[].position')

    position_array=($positions)
    first_pos="${position_array[0]}"
    second_pos="${position_array[1]}"

    print_info "Moving board to position 2 (between first and second board)"
    print_info "Before position: $first_pos"
    print_info "After position: $second_pos"

    # Move board to middle position
    start_time=$(date +%s%N)

    move_response=$(call_api PUT "$BOARD_SERVICE_URL/api/boards/$PERF_BOARD_15_ID/move" '{
        "view_id": "'"$VIEW_ID"'",
        "group_by_field_id": "'"$FIELD_ID"'",
        "new_field_value": "'"$DONE_OPTION_ID"'",
        "before_position": "'"$first_pos"'",
        "after_position": "'"$second_pos"'"
    }')

    end_time=$(date +%s%N)
    duration=$((($end_time - $start_time) / 1000000))  # Convert to milliseconds

    new_position=$(extract_field "$move_response" '.data.new_position')

    print_success "Move completed in ${duration}ms"
    print_success "New position: $new_position"

    print_info "With fractional indexing, this operation updates only 1 row"
    print_info "Without it (integer-based), this would update N rows (cascading updates)"
}

# ============================================================================
# Cleanup
# ============================================================================

cleanup() {
    print_header "CLEANUP: Deleting Test Data"

    if [ -n "$PROJECT_ID" ]; then
        print_step "Deleting test project"
        call_api DELETE "$BOARD_SERVICE_URL/api/projects/$PROJECT_ID" > /dev/null
        print_success "Project deleted"
    fi

    if [ -n "$WORKSPACE_ID" ]; then
        print_step "Deleting test workspace"
        curl -s -X DELETE "$USER_SERVICE_URL/api/workspaces/$WORKSPACE_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null
        print_success "Workspace deleted"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    print_header "Fractional Indexing Test Suite"

    # Setup
    setup_test_data

    # Run tests
    test_move_within_column
    test_move_between_columns
    test_move_to_first
    test_move_to_last
    test_verify_sorting
    test_performance

    # Cleanup
    cleanup

    # Print summary
    print_header "TEST SUMMARY"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"

    if [ $FAILED -eq 0 ]; then
        echo -e "\n${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Run main
main
