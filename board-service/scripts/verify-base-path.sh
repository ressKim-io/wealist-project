#!/bin/bash

# Board Service Base Path Verification Script
# This script verifies that the base path configuration is working correctly

set -e

echo "=========================================="
echo "Board Service Base Path Verification"
echo "=========================================="
echo ""

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8000}"
BASE_PATH="${BASE_PATH:-/api/boards}"

echo "Testing Base Path Configuration:"
echo "  Base URL: $BASE_URL"
echo "  Base Path: $BASE_PATH"
echo ""

# Test 1: Health check with base path
echo "Test 1: Health check with base path"
echo "  URL: ${BASE_URL}${BASE_PATH}/health"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}${BASE_PATH}/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "  ✓ Health check successful (HTTP $HTTP_CODE)"
    echo "  Response: $BODY"
else
    echo "  ✗ Health check failed (HTTP $HTTP_CODE)"
    echo "  Response: $BODY"
    exit 1
fi
echo ""

# Test 2: Swagger documentation with base path
echo "Test 2: Swagger documentation with base path"
echo "  URL: ${BASE_URL}${BASE_PATH}/swagger/index.html"
SWAGGER_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${BASE_PATH}/swagger/index.html")

if [ "$SWAGGER_CODE" = "200" ]; then
    echo "  ✓ Swagger accessible (HTTP $SWAGGER_CODE)"
else
    echo "  ✗ Swagger not accessible (HTTP $SWAGGER_CODE)"
fi
echo ""

echo "=========================================="
echo "Verification Complete!"
echo "=========================================="
