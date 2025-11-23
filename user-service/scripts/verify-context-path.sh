#!/bin/bash
# =============================================================================
# User Service Context Path Verification Script
# =============================================================================
# This script verifies that the context path is correctly configured
# for both local and AWS environments.
#
# Usage:
#   ./scripts/verify-context-path.sh [local|aws]
#
# Examples:
#   ./scripts/verify-context-path.sh local    # Test local environment
#   ./scripts/verify-context-path.sh aws      # Test AWS environment
# =============================================================================

set -e

ENVIRONMENT=${1:-local}
BASE_URL=""
CONTEXT_PATH=""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "User Service Context Path Verification"
echo "=========================================="
echo ""

# Set URLs based on environment
if [ "$ENVIRONMENT" = "aws" ]; then
    BASE_URL="https://api.wealist.co.kr"
    CONTEXT_PATH="/api/users"
    echo "Environment: AWS (Production/Dev)"
    echo "Base URL: ${BASE_URL}"
    echo "Context Path: ${CONTEXT_PATH}"
elif [ "$ENVIRONMENT" = "local" ]; then
    BASE_URL="http://localhost:8080"
    CONTEXT_PATH=""
    echo "Environment: Local Development"
    echo "Base URL: ${BASE_URL}"
    echo "Context Path: (none)"
else
    echo -e "${RED}❌ Invalid environment: ${ENVIRONMENT}${NC}"
    echo "Usage: $0 [local|aws]"
    exit 1
fi

echo ""
echo "=========================================="
echo "Running Health Checks..."
echo "=========================================="
echo ""

# Test 1: Health Check Endpoint
echo "Test 1: Health Check Endpoint"
echo "URL: ${BASE_URL}${CONTEXT_PATH}/actuator/health"
echo ""

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${CONTEXT_PATH}/actuator/health" || echo "000")

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Health check passed (HTTP ${HTTP_CODE})${NC}"
    
    # Show health details
    HEALTH_RESPONSE=$(curl -s "${BASE_URL}${CONTEXT_PATH}/actuator/health")
    echo "Response:"
    echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"
else
    echo -e "${RED}❌ Health check failed (HTTP ${HTTP_CODE})${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo "Test 2: API Endpoint (Workspaces)"
echo "=========================================="
echo ""

# Test 2: Sample API Endpoint
echo "URL: ${BASE_URL}${CONTEXT_PATH}/api/workspaces/all"
echo ""

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${CONTEXT_PATH}/api/workspaces/all" || echo "000")

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    echo -e "${GREEN}✅ API endpoint accessible (HTTP ${HTTP_CODE})${NC}"
    echo "Note: 401/403 is expected without authentication"
else
    echo -e "${YELLOW}⚠️  API endpoint returned HTTP ${HTTP_CODE}${NC}"
    echo "This may be expected depending on authentication requirements"
fi

echo ""
echo "=========================================="
echo "Test 3: Swagger UI"
echo "=========================================="
echo ""

# Test 3: Swagger UI
echo "URL: ${BASE_URL}${CONTEXT_PATH}/swagger-ui.html"
echo ""

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${CONTEXT_PATH}/swagger-ui.html" || echo "000")

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Swagger UI accessible (HTTP ${HTTP_CODE})${NC}"
elif [ "$HTTP_CODE" = "302" ] || [ "$HTTP_CODE" = "301" ]; then
    echo -e "${GREEN}✅ Swagger UI redirecting (HTTP ${HTTP_CODE})${NC}"
else
    echo -e "${YELLOW}⚠️  Swagger UI returned HTTP ${HTTP_CODE}${NC}"
fi

echo ""
echo "=========================================="
echo "Verification Summary"
echo "=========================================="
echo ""
echo -e "${GREEN}✅ Context path verification completed${NC}"
echo ""
echo "Environment: ${ENVIRONMENT}"
echo "Base URL: ${BASE_URL}"
echo "Context Path: ${CONTEXT_PATH}"
echo ""

if [ "$ENVIRONMENT" = "aws" ]; then
    echo "Next steps:"
    echo "1. Verify ALB routing rules are configured"
    echo "2. Check Target Group health status"
    echo "3. Test end-to-end flow from client"
fi

echo ""
