#!/bin/bash

# Godoc Annotation Quality Check Script
# This script validates that all handler functions have complete godoc annotations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HANDLER_DIR="$PROJECT_ROOT/internal/handler"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Godoc Annotation Quality Check"
echo "=========================================="
echo ""

# Check if handler directory exists
if [ ! -d "$HANDLER_DIR" ]; then
    echo -e "${RED}❌ Error: handler directory not found at $HANDLER_DIR${NC}"
    exit 1
fi

# Required annotations for each handler
REQUIRED_ANNOTATIONS=("@Summary" "@Tags" "@Router")

# Find all handler files (excluding test files)
HANDLER_FILES=$(find "$HANDLER_DIR" -name "*_handler.go" -not -name "*_test.go")

if [ -z "$HANDLER_FILES" ]; then
    echo -e "${RED}❌ Error: No handler files found${NC}"
    exit 1
fi

TOTAL_HANDLERS=0
COMPLETE_HANDLERS=0
INCOMPLETE_HANDLERS=0
ISSUES_FOUND=0

# Store issues in a temp file instead of associative array
ISSUES_TEMP=$(mktemp)
trap "rm -f $ISSUES_TEMP" EXIT

# Function to check annotations for a handler function
check_handler_annotations() {
    local file=$1
    local handler_name=$2
    local start_line=$3
    
    # Extract godoc comment block (lines before the function)
    local godoc_block=$(awk -v start="$start_line" '
        NR < start && /^\/\/ / {
            lines[NR] = $0
        }
        NR == start {
            for (i in lines) print lines[i]
            exit
        }
    ' "$file")
    
    local missing_annotations=()
    
    # Check for required annotations
    for annotation in "${REQUIRED_ANNOTATIONS[@]}"; do
        if ! echo "$godoc_block" | grep -q "$annotation"; then
            missing_annotations+=("$annotation")
        fi
    done
    
    # Check for additional recommended annotations
    local has_params=$(echo "$godoc_block" | grep -q "@Param" && echo "yes" || echo "no")
    local has_success=$(echo "$godoc_block" | grep -q "@Success" && echo "yes" || echo "no")
    local has_failure=$(echo "$godoc_block" | grep -q "@Failure" && echo "yes" || echo "no")
    
    # Return results
    if [ ${#missing_annotations[@]} -eq 0 ]; then
        echo "COMPLETE|$has_params|$has_success|$has_failure"
    else
        echo "INCOMPLETE|${missing_annotations[*]}"
    fi
}

echo "📋 Analyzing handler files..."
echo ""

# Process each handler file
for file in $HANDLER_FILES; do
    filename=$(basename "$file")
    echo -e "${BLUE}Checking: $filename${NC}"
    
    # Find all exported handler functions (func (h *Handler) FunctionName)
    handlers=$(grep -n "^func (h \*.*Handler) [A-Z]" "$file" | grep -v "^//" || true)
    
    if [ -z "$handlers" ]; then
        echo -e "${YELLOW}  ⚠️  No handler functions found${NC}"
        echo ""
        continue
    fi
    
    while IFS= read -r line; do
        if [ -z "$line" ]; then
            continue
        fi
        
        line_num=$(echo "$line" | cut -d: -f1)
        handler_name=$(echo "$line" | sed -E 's/.*func \(h \*.*Handler\) ([A-Z][a-zA-Z0-9]*).*/\1/')
        
        TOTAL_HANDLERS=$((TOTAL_HANDLERS + 1))
        
        # Check annotations
        result=$(check_handler_annotations "$file" "$handler_name" "$line_num")
        status=$(echo "$result" | cut -d'|' -f1)
        
        if [ "$status" = "COMPLETE" ]; then
            COMPLETE_HANDLERS=$((COMPLETE_HANDLERS + 1))
            has_params=$(echo "$result" | cut -d'|' -f2)
            has_success=$(echo "$result" | cut -d'|' -f3)
            has_failure=$(echo "$result" | cut -d'|' -f4)
            
            echo -e "  ${GREEN}✅ $handler_name${NC}"
            
            # Check for recommended annotations
            warnings=""
            if [ "$has_params" = "no" ]; then
                warnings="${warnings}@Param "
            fi
            if [ "$has_success" = "no" ]; then
                warnings="${warnings}@Success "
            fi
            if [ "$has_failure" = "no" ]; then
                warnings="${warnings}@Failure "
            fi
            
            if [ -n "$warnings" ]; then
                echo -e "     ${YELLOW}⚠️  Missing recommended: $warnings${NC}"
                ISSUES_FOUND=$((ISSUES_FOUND + 1))
            fi
        else
            INCOMPLETE_HANDLERS=$((INCOMPLETE_HANDLERS + 1))
            ISSUES_FOUND=$((ISSUES_FOUND + 1))
            missing=$(echo "$result" | cut -d'|' -f2)
            echo -e "  ${RED}❌ $handler_name${NC}"
            echo -e "     ${RED}Missing required: $missing${NC}"
            
            # Store issue for summary
            echo "$filename::$handler_name|$missing" >> "$ISSUES_TEMP"
        fi
    done <<< "$handlers"
    
    echo ""
done

# Calculate quality score
if [ "$TOTAL_HANDLERS" -gt 0 ]; then
    QUALITY_SCORE=$((COMPLETE_HANDLERS * 100 / TOTAL_HANDLERS))
else
    QUALITY_SCORE=0
fi

echo "=========================================="
echo "Summary"
echo "=========================================="
echo "Total handlers analyzed: $TOTAL_HANDLERS"
echo "Complete annotations: $COMPLETE_HANDLERS"
echo "Incomplete annotations: $INCOMPLETE_HANDLERS"
echo "Quality score: $QUALITY_SCORE%"
echo ""

# List handlers with incomplete annotations
if [ "$INCOMPLETE_HANDLERS" -gt 0 ] && [ -f "$ISSUES_TEMP" ]; then
    echo "=========================================="
    echo "Handlers with Incomplete Annotations"
    echo "=========================================="
    while IFS='|' read -r key missing; do
        if [ -n "$key" ]; then
            echo -e "${RED}❌ $key${NC}"
            echo -e "   Missing: $missing"
        fi
    done < "$ISSUES_TEMP"
    echo ""
fi

# Annotation guidelines
echo "=========================================="
echo "Annotation Guidelines"
echo "=========================================="
echo "Required annotations for all handlers:"
echo "  • @Summary      - Brief description"
echo "  • @Tags         - API group tag"
echo "  • @Router       - Route path and method"
echo ""
echo "Recommended annotations:"
echo "  • @Description  - Detailed description"
echo "  • @Param        - Request parameters"
echo "  • @Success      - Success response"
echo "  • @Failure      - Error responses"
echo "  • @Accept       - Request content type (if applicable)"
echo "  • @Produce      - Response content type"
echo ""

echo "=========================================="
echo "Final Result"
echo "=========================================="

if [ "$INCOMPLETE_HANDLERS" -eq 0 ]; then
    if [ "$ISSUES_FOUND" -eq 0 ]; then
        echo -e "${GREEN}✅ SUCCESS: All handlers have complete annotations${NC}"
        echo -e "${GREEN}✅ No issues found${NC}"
        exit 0
    else
        echo -e "${GREEN}✅ SUCCESS: All handlers have required annotations${NC}"
        echo -e "${YELLOW}⚠️  WARNING: $ISSUES_FOUND handler(s) missing recommended annotations${NC}"
        echo -e "${YELLOW}⚠️  Consider adding @Param, @Success, and @Failure annotations${NC}"
        exit 0
    fi
else
    echo -e "${RED}❌ FAILURE: $INCOMPLETE_HANDLERS handler(s) have incomplete annotations${NC}"
    echo -e "${RED}❌ Quality score: $QUALITY_SCORE% (target: 100%)${NC}"
    echo ""
    echo "Please add missing annotations to the handlers listed above."
    echo "Refer to the annotation guidelines for proper format."
    exit 1
fi
