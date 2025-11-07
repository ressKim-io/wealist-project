#!/bin/bash

# =============================================================================
# User Service Login Helper Script
# Get JWT token and user info from User Service
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8080}"

print_header() {
    echo -e "${BLUE}=================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# =============================================================================
# Get Test Token (Development Only)
# =============================================================================

get_test_token() {
    print_header "Getting Test Token (Development Mode)"
    print_info "Calling: $USER_SERVICE_URL/api/auth/test"

    response=$(curl -s "$USER_SERVICE_URL/api/auth/test")

    # Check if test endpoint returned token
    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        JWT_TOKEN=$(echo "$response" | jq -r '.accessToken')
        USER_ID=$(echo "$response" | jq -r '.userId')

        print_success "Test token received!"
        echo ""
        echo "User ID: $USER_ID"
        echo "Token: ${JWT_TOKEN:0:50}..." # Show first 50 chars only
        echo ""

        # Export for child processes
        export JWT_TOKEN
        export USER_ID

        # Get workspaces
        get_workspaces
    else
        print_error "Failed to get test token!"
        echo "$response" | jq '.'
        echo ""
        print_info "Test endpoint might not be available. Try manual login instead."
        exit 1
    fi
}

# =============================================================================
# Login to User Service (Manual)
# =============================================================================

login() {
    local email="$1"
    local password="$2"

    print_header "Logging in to User Service"
    print_info "Email: $email"

    response=$(curl -s -X POST "$USER_SERVICE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\"
        }")

    # Check if login was successful
    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        JWT_TOKEN=$(echo "$response" | jq -r '.accessToken')
        USER_ID=$(echo "$response" | jq -r '.userId')

        print_success "Login successful!"
        echo ""
        echo "User ID: $USER_ID"
        echo "Token: ${JWT_TOKEN:0:50}..." # Show first 50 chars only
        echo ""

        # Export for child processes
        export JWT_TOKEN
        export USER_ID

        # Get workspaces
        get_workspaces
    else
        print_error "Login failed!"
        echo "$response" | jq '.'
        exit 1
    fi
}

# =============================================================================
# Get Workspaces from User Service
# =============================================================================

get_workspaces() {
    print_header "Fetching Workspaces"

    response=$(curl -s "$USER_SERVICE_URL/api/workspace" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | jq -e '.data' > /dev/null 2>&1; then
        workspaces=$(echo "$response" | jq -r '.data[] | {id, name}')

        if [ -n "$workspaces" ]; then
            print_success "Found workspaces:"
            echo "$response" | jq -r '.data[] | "  - \(.name) (ID: \(.id))"'

            # Get first workspace ID
            WORKSPACE_ID=$(echo "$response" | jq -r '.data[0].id')
            export WORKSPACE_ID

            echo ""
            print_info "Using first workspace: $WORKSPACE_ID"
        else
            print_info "No workspaces found. Please create a workspace first."
        fi
    else
        echo "Failed to get workspaces"
        echo "$response" | jq '.'
    fi
}

# =============================================================================
# Main
# =============================================================================

main() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   User Service Login Helper                      ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════╝${NC}"
    echo ""

    # If no arguments provided, use test endpoint (development mode)
    if [ $# -eq 0 ]; then
        print_info "No credentials provided - using test endpoint"
        get_test_token
    else
        # Manual login with credentials
        if [ $# -lt 2 ]; then
            echo "Usage:"
            echo "  $0                          # Use test token (development only)"
            echo "  $0 <email> <password>       # Manual login"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Get test token"
            echo "  $0 user@example.com mypassword       # Login with credentials"
            exit 1
        fi

        EMAIL="$1"
        PASSWORD="$2"
        login "$EMAIL" "$PASSWORD"
    fi

    echo ""
    print_header "Environment Variables Set"
    echo "export JWT_TOKEN='$JWT_TOKEN'"
    echo "export USER_ID='$USER_ID'"
    echo "export WORKSPACE_ID='$WORKSPACE_ID'"
    echo ""
    print_info "Copy and paste the export commands above to set environment variables,"
    print_info "then run: ./test_board_api.sh"
}

main "$@"
