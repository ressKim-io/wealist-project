#!/usr/bin/env bash

# Endpoint Coverage Validation Script
# This script validates that all routes defined in router.go are documented in swagger.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ROUTER_FILE="$PROJECT_ROOT/internal/router/router.go"
SWAGGER_FILE="$PROJECT_ROOT/docs/swagger.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Swagger Endpoint Coverage Validation"
echo "=========================================="
echo ""

# Check if required files exist
if [ ! -f "$ROUTER_FILE" ]; then
    echo -e "${RED}❌ Error: router.go not found at $ROUTER_FILE${NC}"
    exit 1
fi

if [ ! -f "$SWAGGER_FILE" ]; then
    echo -e "${RED}❌ Error: swagger.yaml not found at $SWAGGER_FILE${NC}"
    exit 1
fi

# Create temp files for comparison
ROUTER_TEMP=$(mktemp)
SWAGGER_TEMP=$(mktemp)

# Cleanup on exit
trap "rm -f $ROUTER_TEMP $SWAGGER_TEMP" EXIT

echo "📋 Extracting routes from router.go..."

# Extract routes from router.go using Python for more reliable parsing
python3 <<PYTHON_SCRIPT > "$ROUTER_TEMP"
import re

router_file = "$ROUTER_FILE"

with open(router_file, 'r') as f:
    content = f.read()

# Pattern to match route definitions
pattern = r'(projects|boards|comments|participants|fieldOptions|joinRequests)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]*)"'

matches = re.findall(pattern, content)

base_paths = {
    'projects': '/projects',
    'boards': '/boards',
    'comments': '/comments',
    'participants': '/participants',
    'fieldOptions': '/field-options',
    'joinRequests': '/join-requests'
}

routes = set()
for group, method, path in matches:
    base = base_paths[group]
    full_path = base + path if path else base
    # Convert :param to {param}
    full_path = re.sub(r':(\w+)', r'{\1}', full_path)
    routes.add(f"{method.upper()} /api{full_path}")

for route in sorted(routes):
    print(route)
PYTHON_SCRIPT

echo "📋 Extracting paths from swagger.yaml..."

# Extract endpoints from swagger.yaml using Python
python3 <<PYTHON_SCRIPT > "$SWAGGER_TEMP"
import re

swagger_file = "$SWAGGER_FILE"

with open(swagger_file, 'r') as f:
    lines = f.readlines()

endpoints = set()
current_path = None

for i, line in enumerate(lines):
    # Match path definitions (e.g., "  /boards:")
    path_match = re.match(r'^  (/(?:boards|projects|comments|participants|field-options|join-requests)[^:]*):$', line)
    if path_match:
        current_path = path_match.group(1)
        continue
    
    # Match method definitions (e.g., "    get:")
    if current_path:
        method_match = re.match(r'^    (get|post|put|patch|delete):$', line)
        if method_match:
            method = method_match.group(1).upper()
            endpoints.add(f"{method} /api{current_path}")
        # Reset if we hit another path
        elif re.match(r'^  /', line):
            current_path = None

for endpoint in sorted(endpoints):
    print(endpoint)
PYTHON_SCRIPT

# Count routes
ROUTER_COUNT=$(wc -l < "$ROUTER_TEMP" | tr -d ' ')
SWAGGER_COUNT=$(wc -l < "$SWAGGER_TEMP" | tr -d ' ')

echo ""
echo "=========================================="
echo "Route Statistics"
echo "=========================================="
echo "Router endpoints: $ROUTER_COUNT"
echo "Swagger endpoints: $SWAGGER_COUNT"
echo ""

# Find missing endpoints
echo "=========================================="
echo "Checking Coverage"
echo "=========================================="
echo ""

MISSING_COUNT=0
MISSING_ENDPOINTS=""

while IFS= read -r route; do
    if [ -n "$route" ]; then
        if ! grep -Fxq "$route" "$SWAGGER_TEMP"; then
            MISSING_COUNT=$((MISSING_COUNT + 1))
            MISSING_ENDPOINTS="${MISSING_ENDPOINTS}${route}\n"
            echo -e "${RED}❌ Missing: $route${NC}"
        fi
    fi
done < "$ROUTER_TEMP"

if [ "$MISSING_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ All router endpoints are documented${NC}"
fi

# Calculate coverage
if [ "$ROUTER_COUNT" -gt 0 ]; then
    DOCUMENTED_COUNT=$((ROUTER_COUNT - MISSING_COUNT))
    COVERAGE=$((DOCUMENTED_COUNT * 100 / ROUTER_COUNT))
else
    COVERAGE=0
fi

echo ""
echo "=========================================="
echo "Coverage Summary"
echo "=========================================="
echo "Documented: $DOCUMENTED_COUNT / $ROUTER_COUNT"
echo "Coverage: $COVERAGE%"
echo ""

# Check for extra endpoints in swagger (not in router)
echo "=========================================="
echo "Extra Endpoints in Swagger"
echo "=========================================="
echo ""

EXTRA_COUNT=0
while IFS= read -r endpoint; do
    if [ -n "$endpoint" ]; then
        if ! grep -Fxq "$endpoint" "$ROUTER_TEMP"; then
            EXTRA_COUNT=$((EXTRA_COUNT + 1))
            echo -e "${YELLOW}⚠️  Extra: $endpoint${NC}"
        fi
    fi
done < "$SWAGGER_TEMP"

if [ "$EXTRA_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ No extra endpoints found${NC}"
fi

echo ""
echo "=========================================="
echo "Final Result"
echo "=========================================="

if [ "$COVERAGE" -eq 100 ] && [ "$MISSING_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ SUCCESS: Full coverage achieved (100%)${NC}"
    echo -e "${GREEN}✅ All router endpoints are documented in Swagger${NC}"
    exit 0
else
    echo -e "${RED}❌ FAILURE: Incomplete coverage ($COVERAGE%)${NC}"
    echo -e "${RED}❌ $MISSING_COUNT endpoint(s) missing from Swagger documentation${NC}"
    if [ "$MISSING_COUNT" -gt 0 ]; then
        echo ""
        echo "Missing endpoints:"
        echo -e "$MISSING_ENDPOINTS"
    fi
    exit 1
fi
