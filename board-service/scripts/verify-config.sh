#!/bin/bash

# =============================================================================
# Board Service Configuration Verification Script
# =============================================================================
# This script verifies that the board-service configuration is correct,
# especially the User Service connection settings.
#
# Usage:
#   ./scripts/verify-config.sh
# =============================================================================

set -e

echo "=========================================="
echo "Board Service Configuration Verification"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if config file exists
echo "1. Checking config file..."
if [ -f "configs/config.yaml" ]; then
    echo -e "${GREEN}✓${NC} config.yaml found"
    
    # Extract user_api.base_url from config.yaml
    BASE_URL=$(grep -A 1 "user_api:" configs/config.yaml | grep "base_url:" | awk '{print $2}' | tr -d '"')
    echo "   Base URL from config.yaml: $BASE_URL"
    
    # Check for trailing slash
    if [[ "$BASE_URL" == */ ]]; then
        echo -e "${YELLOW}⚠${NC}  Warning: Base URL has trailing slash (will be auto-removed)"
    fi
else
    echo -e "${YELLOW}⚠${NC}  config.yaml not found (will use environment variables)"
fi

echo ""

# Check environment variables
echo "2. Checking environment variables..."

if [ -n "$USER_SERVICE_URL" ]; then
    echo -e "${GREEN}✓${NC} USER_SERVICE_URL is set: $USER_SERVICE_URL"
    EFFECTIVE_URL="$USER_SERVICE_URL"
elif [ -n "$USER_API_BASE_URL" ]; then
    echo -e "${GREEN}✓${NC} USER_API_BASE_URL is set: $USER_API_BASE_URL"
    EFFECTIVE_URL="$USER_API_BASE_URL"
elif [ -n "$BASE_URL" ]; then
    echo -e "${YELLOW}⚠${NC}  Using config.yaml value: $BASE_URL"
    EFFECTIVE_URL="$BASE_URL"
else
    echo -e "${RED}✗${NC} No User Service URL configured!"
    exit 1
fi

echo ""

# Validate URL format
echo "3. Validating URL format..."

# Check for trailing slash
if [[ "$EFFECTIVE_URL" == */ ]]; then
    echo -e "${YELLOW}⚠${NC}  URL has trailing slash: $EFFECTIVE_URL"
    echo "   This will be automatically removed at runtime"
    EFFECTIVE_URL="${EFFECTIVE_URL%/}"
fi

# Check scheme
if [[ "$EFFECTIVE_URL" =~ ^https?:// ]]; then
    echo -e "${GREEN}✓${NC} URL has valid scheme (http/https)"
else
    echo -e "${RED}✗${NC} URL missing scheme (http:// or https://)"
    exit 1
fi

# Extract host and port
HOST_PORT=$(echo "$EFFECTIVE_URL" | sed -E 's|^https?://||')
echo "   Host: $HOST_PORT"

# Check if port is specified
if [[ "$HOST_PORT" =~ :[0-9]+$ ]]; then
    echo -e "${GREEN}✓${NC} Port is specified"
else
    echo -e "${YELLOW}⚠${NC}  Port not specified (will use default: 80 for http, 443 for https)"
fi

echo ""

# Show expected endpoints
echo "4. Expected User Service endpoints:"
echo "   - Validate member: ${EFFECTIVE_URL}/api/workspaces/{workspaceId}/validate-member/{userId}"
echo "   - Get user:        ${EFFECTIVE_URL}/api/users/{userId}"
echo "   - Get workspace profile: ${EFFECTIVE_URL}/api/profiles/workspace/{workspaceId}"
echo "   - Get workspace:   ${EFFECTIVE_URL}/api/workspaces/{workspaceId}"

echo ""

# Test connectivity (optional)
echo "5. Testing connectivity to User Service..."

# Extract just the host:port for curl
CURL_URL=$(echo "$EFFECTIVE_URL" | sed -E 's|^(https?://[^/]+).*|\1|')

if command -v curl &> /dev/null; then
    # Try to connect to health endpoint
    if curl -s -f -m 5 "${CURL_URL}/actuator/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} User Service is reachable at ${CURL_URL}"
    else
        echo -e "${YELLOW}⚠${NC}  Cannot reach User Service at ${CURL_URL}"
        echo "   This is normal if services are not running yet"
        echo "   Try: curl ${CURL_URL}/actuator/health"
    fi
else
    echo -e "${YELLOW}⚠${NC}  curl not available, skipping connectivity test"
fi

echo ""

# Summary
echo "=========================================="
echo "Configuration Summary"
echo "=========================================="
echo "Effective User Service URL: $EFFECTIVE_URL"
echo ""
echo -e "${GREEN}Configuration verification complete!${NC}"
echo ""
echo "To start the service with this configuration:"
echo "  go run cmd/api/main.go"
echo ""
echo "To override with environment variable:"
echo "  USER_SERVICE_URL=http://localhost:8080 go run cmd/api/main.go"
