#!/bin/bash

# DTO Schema Validation Script
# This script validates that all DTO types used in handlers are defined in swagger.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HANDLER_DIR="$PROJECT_ROOT/internal/handler"
DTO_DIR="$PROJECT_ROOT/internal/dto"
SWAGGER_FILE="$PROJECT_ROOT/docs/swagger.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Swagger DTO Schema Validation"
echo "=========================================="
echo ""

# Check if required files exist
if [ ! -d "$HANDLER_DIR" ]; then
    echo -e "${RED}❌ Error: handler directory not found at $HANDLER_DIR${NC}"
    exit 1
fi

if [ ! -d "$DTO_DIR" ]; then
    echo -e "${RED}❌ Error: dto directory not found at $DTO_DIR${NC}"
    exit 1
fi

if [ ! -f "$SWAGGER_FILE" ]; then
    echo -e "${RED}❌ Error: swagger.yaml not found at $SWAGGER_FILE${NC}"
    exit 1
fi

# Extract DTO types from handler files
# Pattern matches: dto.TypeName in function signatures and variable declarations
echo "📋 Extracting DTO types from handler files..."
HANDLER_DTOS=$(grep -rh "dto\." "$HANDLER_DIR"/*.go 2>/dev/null | \
    grep -v "^//" | \
    grep -v "^\s*//" | \
    grep -oE 'dto\.[A-Z][a-zA-Z0-9]+' | \
    sed 's/dto\.//' | \
    sort -u)

# Extract DTO struct definitions from dto package
echo "📋 Extracting DTO struct definitions..."
DTO_STRUCTS=$(grep -rh "^type [A-Z]" "$DTO_DIR"/*.go 2>/dev/null | \
    grep "struct" | \
    awk '{print $2}' | \
    sort -u)

# Extract definitions from swagger.yaml
# Pattern matches: project-board-api_internal_dto.TypeName:
echo "📋 Extracting DTO definitions from swagger.yaml..."
SWAGGER_DTOS=$(grep -E 'project-board-api_internal_dto\.[A-Z]' "$SWAGGER_FILE" | \
    sed 's/.*project-board-api_internal_dto\.//' | \
    sed 's/:$//' | \
    sed 's/^\s*//' | \
    sed 's/\s*$//' | \
    sort -u)

# Count DTOs
HANDLER_DTO_COUNT=$(echo "$HANDLER_DTOS" | grep -v '^$' | wc -l | tr -d ' ')
DTO_STRUCT_COUNT=$(echo "$DTO_STRUCTS" | grep -v '^$' | wc -l | tr -d ' ')
SWAGGER_DTO_COUNT=$(echo "$SWAGGER_DTOS" | grep -v '^$' | wc -l | tr -d ' ')

echo ""
echo "=========================================="
echo "DTO Statistics"
echo "=========================================="
echo "DTOs referenced in handlers: $HANDLER_DTO_COUNT"
echo "DTO structs defined: $DTO_STRUCT_COUNT"
echo "DTOs documented in Swagger: $SWAGGER_DTO_COUNT"
echo ""

# Check for missing DTOs in Swagger (used in handlers but not documented)
echo "=========================================="
echo "Checking Handler DTOs Coverage"
echo "=========================================="
echo ""

MISSING_COUNT=0
MISSING_DTOS=""

while IFS= read -r dto; do
    if [ -n "$dto" ]; then
        if ! echo "$SWAGGER_DTOS" | grep -q "^$dto$"; then
            MISSING_COUNT=$((MISSING_COUNT + 1))
            MISSING_DTOS="${MISSING_DTOS}${dto}\n"
            echo -e "${RED}❌ Missing in Swagger: $dto${NC}"
        fi
    fi
done <<< "$HANDLER_DTOS"

if [ "$MISSING_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ All handler DTOs are documented in Swagger${NC}"
fi

# Check for DTOs defined but not documented
echo ""
echo "=========================================="
echo "Checking Struct Definitions Coverage"
echo "=========================================="
echo ""

UNDOCUMENTED_COUNT=0
UNDOCUMENTED_DTOS=""

while IFS= read -r dto; do
    if [ -n "$dto" ]; then
        if ! echo "$SWAGGER_DTOS" | grep -q "^$dto$"; then
            UNDOCUMENTED_COUNT=$((UNDOCUMENTED_COUNT + 1))
            UNDOCUMENTED_DTOS="${UNDOCUMENTED_DTOS}${dto}\n"
            echo -e "${YELLOW}⚠️  Defined but not documented: $dto${NC}"
        fi
    fi
done <<< "$DTO_STRUCTS"

if [ "$UNDOCUMENTED_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ All DTO structs are documented in Swagger${NC}"
fi

# Check for extra DTOs in Swagger (documented but not defined)
echo ""
echo "=========================================="
echo "Checking for Extra Swagger Definitions"
echo "=========================================="
echo ""

EXTRA_COUNT=0
EXTRA_DTOS=""

while IFS= read -r dto; do
    if [ -n "$dto" ]; then
        if ! echo "$DTO_STRUCTS" | grep -q "^$dto$"; then
            EXTRA_COUNT=$((EXTRA_COUNT + 1))
            EXTRA_DTOS="${EXTRA_DTOS}${dto}\n"
            echo -e "${YELLOW}⚠️  Documented but not defined: $dto${NC}"
        fi
    fi
done <<< "$SWAGGER_DTOS"

if [ "$EXTRA_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ No extra Swagger definitions found${NC}"
fi

# Calculate coverage
if [ "$HANDLER_DTO_COUNT" -gt 0 ]; then
    DOCUMENTED_COUNT=$((HANDLER_DTO_COUNT - MISSING_COUNT))
    COVERAGE=$((DOCUMENTED_COUNT * 100 / HANDLER_DTO_COUNT))
else
    COVERAGE=100
fi

echo ""
echo "=========================================="
echo "Coverage Summary"
echo "=========================================="
echo "Handler DTOs documented: $DOCUMENTED_COUNT / $HANDLER_DTO_COUNT"
echo "Coverage: $COVERAGE%"
echo ""

# Detailed DTO list (optional, for debugging)
if [ "${VERBOSE:-0}" = "1" ]; then
    echo "=========================================="
    echo "Detailed DTO Lists"
    echo "=========================================="
    echo ""
    echo "Handler DTOs:"
    echo "$HANDLER_DTOS"
    echo ""
    echo "DTO Structs:"
    echo "$DTO_STRUCTS"
    echo ""
    echo "Swagger DTOs:"
    echo "$SWAGGER_DTOS"
    echo ""
fi

echo "=========================================="
echo "Final Result"
echo "=========================================="

if [ "$MISSING_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ SUCCESS: All handler DTOs are documented in Swagger${NC}"
    if [ "$UNDOCUMENTED_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  WARNING: $UNDOCUMENTED_COUNT DTO struct(s) defined but not used in handlers${NC}"
    fi
    exit 0
else
    echo -e "${RED}❌ FAILURE: $MISSING_COUNT DTO(s) missing from Swagger documentation${NC}"
    echo ""
    echo "Missing DTOs:"
    echo -e "$MISSING_DTOS"
    echo ""
    echo "These DTOs are used in handlers but not documented in swagger.yaml"
    echo "Please add godoc annotations for these DTOs or ensure they are referenced in handler annotations"
    exit 1
fi
