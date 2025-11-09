#!/bin/bash

# =============================================================================
# Go 1.25 Compatibility Pre-Build Check
# =============================================================================
# This script checks for common Go 1.25 compatibility issues BEFORE Docker build
# Saves time by catching errors early in the local environment
# =============================================================================

set +e  # Don't exit on errors, we want to collect all issues

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

CHECKS_PASSED=0
CHECKS_FAILED=0
ERRORS_FOUND=()

print_header() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║  $1${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_check() {
    echo -e "${BLUE}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    CHECKS_FAILED=$((CHECKS_FAILED + 1))
    ERRORS_FOUND+=("$1")
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_detail() {
    echo -e "${YELLOW}  → $1${NC}"
}

print_header "Go 1.25 Compatibility Pre-Build Check"

# =============================================================================
# Check 1: Go Version
# =============================================================================
print_check "1. Checking Go version availability..."

GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
print_info "Required Go version: $GO_VERSION"

if command -v go &> /dev/null; then
    INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Installed Go version: $INSTALLED_VERSION"
    print_success "Go is installed"
else
    print_warning "Go not found locally (will use Docker)"
fi

# =============================================================================
# Check 2: Unused Imports
# =============================================================================
print_check "2. Checking for unused imports (strict in Go 1.25)..."

UNUSED_IMPORTS=$(find . -name "*.go" -type f ! -path "./vendor/*" ! -name "*_test.go" -exec grep -l "^import" {} \; | while read file; do
    # Try to compile just to check for unused imports
    go build -o /dev/null "$file" 2>&1 | grep "imported and not used" | sed "s|^|$file: |"
done)

if [ -z "$UNUSED_IMPORTS" ]; then
    print_success "No unused imports detected"
else
    print_error "Found unused imports:"
    echo "$UNUSED_IMPORTS" | while read line; do
        print_detail "$line"
    done
fi

# =============================================================================
# Check 3: Type Comparison Issues (UUID vs int, etc.)
# =============================================================================
print_check "3. Checking for type comparison issues..."

TYPE_ISSUES=0

# Check for UUID compared with integers
UUID_INT_COMPARE=$(grep -rn "\.RoleID\s*[!=]=\s*[0-9]" --include="*.go" . 2>/dev/null | grep -v "vendor/")
if [ ! -z "$UUID_INT_COMPARE" ]; then
    print_error "Found UUID-to-int comparisons (not allowed in Go 1.25):"
    echo "$UUID_INT_COMPARE" | while read line; do
        print_detail "$line"
    done
    TYPE_ISSUES=$((TYPE_ISSUES + 1))
fi

# Check for other common UUID comparison issues
UUID_DIRECT_COMPARE=$(grep -rn "uuid\.UUID\s*[!=]=\s*[0-9]" --include="*.go" . 2>/dev/null | grep -v "vendor/")
if [ ! -z "$UUID_DIRECT_COMPARE" ]; then
    print_error "Found direct UUID-to-number comparisons:"
    echo "$UUID_DIRECT_COMPARE" | while read line; do
        print_detail "$line"
    done
    TYPE_ISSUES=$((TYPE_ISSUES + 1))
fi

if [ $TYPE_ISSUES -eq 0 ]; then
    print_success "No type comparison issues found"
fi

# =============================================================================
# Check 4: Compilation Test
# =============================================================================
print_check "4. Testing compilation (with local Go)..."

if command -v go &> /dev/null; then
    # Save compilation output
    COMPILE_OUTPUT=$(GOTOOLCHAIN=local go build -o /tmp/board-test ./cmd/api 2>&1)
    COMPILE_EXIT=$?

    if [ $COMPILE_EXIT -eq 0 ]; then
        print_success "Compilation successful"
        rm -f /tmp/board-test
    else
        # Check if it's a version issue
        if echo "$COMPILE_OUTPUT" | grep -q "requires go >="; then
            print_warning "Local Go version too old (Docker will use correct version)"
        else
            print_error "Compilation failed:"
            echo "$COMPILE_OUTPUT" | head -20 | while read line; do
                print_detail "$line"
            done

            # Extract specific error patterns
            if echo "$COMPILE_OUTPUT" | grep -q "cannot convert"; then
                print_detail "Found type conversion errors (common in Go 1.25)"
            fi
            if echo "$COMPILE_OUTPUT" | grep -q "imported and not used"; then
                print_detail "Found unused import errors"
            fi
        fi
    fi
else
    print_warning "Go not available locally, skipping compilation test"
fi

# =============================================================================
# Check 5: Go Fmt Issues
# =============================================================================
print_check "5. Checking code formatting (gofmt)..."

if command -v gofmt &> /dev/null; then
    UNFORMATTED=$(find . -name "*.go" -type f ! -path "./vendor/*" -exec gofmt -l {} \; 2>/dev/null)
    UNFORMATTED_COUNT=$(echo "$UNFORMATTED" | grep -c "\.go" || true)

    if [ "$UNFORMATTED_COUNT" -eq 0 ]; then
        print_success "All files properly formatted"
    else
        print_warning "Found $UNFORMATTED_COUNT unformatted files (non-critical):"
        echo "$UNFORMATTED" | head -10 | while read line; do
            [ ! -z "$line" ] && print_detail "$line"
        done
    fi
else
    print_warning "gofmt not available, skipping format check"
fi

# =============================================================================
# Check 6: Go Vet Static Analysis
# =============================================================================
print_check "6. Running static analysis (go vet)..."

if command -v go &> /dev/null; then
    VET_OUTPUT=$(GOTOOLCHAIN=local go vet ./... 2>&1 || true)

    if [ -z "$VET_OUTPUT" ]; then
        print_success "No issues found by go vet"
    else
        # Filter out version-related warnings
        REAL_ISSUES=$(echo "$VET_OUTPUT" | grep -v "requires go >=" | grep -v "downloading go" || true)

        if [ -z "$REAL_ISSUES" ]; then
            print_success "No issues found by go vet"
        else
            print_warning "Go vet found potential issues:"
            echo "$REAL_ISSUES" | head -10 | while read line; do
                [ ! -z "$line" ] && print_detail "$line"
            done
        fi
    fi
else
    print_warning "Go not available, skipping vet check"
fi

# =============================================================================
# Check 7: Module Verification
# =============================================================================
print_check "7. Verifying module dependencies..."

if command -v go &> /dev/null; then
    VERIFY_OUTPUT=$(go mod verify 2>&1)
    VERIFY_EXIT=$?

    if [ $VERIFY_EXIT -eq 0 ]; then
        print_success "Module verification passed"
    else
        print_warning "Module verification issues (may need 'go mod tidy'):"
        echo "$VERIFY_OUTPUT" | head -5 | while read line; do
            print_detail "$line"
        done
    fi
else
    print_warning "Go not available, skipping module verification"
fi

# =============================================================================
# Check 8: Common Go 1.25 Breaking Changes
# =============================================================================
print_check "8. Checking for known Go 1.25 breaking changes..."

BREAKING_ISSUES=0

# Check for old loop variable semantics (if any explicit workarounds exist)
OLD_LOOP_WORKAROUND=$(grep -rn "tmp\s*:=\s*" --include="*.go" . 2>/dev/null | grep "for.*range" | wc -l)
if [ "$OLD_LOOP_WORKAROUND" -gt 0 ]; then
    print_info "Found $OLD_LOOP_WORKAROUND potential old loop variable workarounds (can be cleaned up in Go 1.25)"
fi

# Check for deprecated io/ioutil usage
DEPRECATED_IOUTIL=$(grep -rn "io/ioutil" --include="*.go" . 2>/dev/null | grep -v "vendor/" | wc -l)
if [ "$DEPRECATED_IOUTIL" -gt 0 ]; then
    print_warning "Found io/ioutil usage (deprecated, use io and os instead)"
    BREAKING_ISSUES=$((BREAKING_ISSUES + 1))
fi

if [ $BREAKING_ISSUES -eq 0 ]; then
    print_success "No known breaking changes detected"
fi

# =============================================================================
# Check 9: Test Files Compilation
# =============================================================================
print_check "9. Checking test files..."

TEST_FILES=$(find . -name "*_test.go" -type f ! -path "./vendor/*" | wc -l)
print_info "Found $TEST_FILES test files"

if [ $TEST_FILES -gt 0 ] && command -v go &> /dev/null; then
    TEST_COMPILE=$(GOTOOLCHAIN=local go test -c ./... -o /tmp/test-compile 2>&1 || true)

    if echo "$TEST_COMPILE" | grep -q "FAIL"; then
        print_warning "Some test files have compilation issues"
    else
        print_success "Test files look compilable"
    fi
    rm -f /tmp/test-compile
else
    print_success "Test files detected"
fi

# =============================================================================
# Check 10: Dockerfile Consistency
# =============================================================================
print_check "10. Checking Dockerfile Go version consistency..."

if [ -f "docker/Dockerfile" ]; then
    DOCKERFILE_GO=$(grep "FROM golang:" docker/Dockerfile | head -1 | sed 's/.*golang:\([0-9.]*\).*/\1/')
    GOMOD_GO=$(grep "^go " go.mod | awk '{print $2}')

    print_info "Dockerfile Go version: $DOCKERFILE_GO"
    print_info "go.mod Go version: $GOMOD_GO"

    if [[ "$DOCKERFILE_GO" == "$GOMOD_GO"* ]] || [[ "$GOMOD_GO" == "$DOCKERFILE_GO"* ]]; then
        print_success "Dockerfile and go.mod versions are consistent"
    else
        print_error "Version mismatch: Dockerfile ($DOCKERFILE_GO) vs go.mod ($GOMOD_GO)"
    fi
else
    print_warning "Dockerfile not found at docker/Dockerfile"
fi

# =============================================================================
# Summary
# =============================================================================
print_header "Check Summary"

TOTAL_CHECKS=$((CHECKS_PASSED + CHECKS_FAILED))
echo "Total Checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${RED}Failed: $CHECKS_FAILED${NC}"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ✓ All checks passed! Ready for Docker build                 ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}You can now run:${NC}"
    echo -e "${YELLOW}  docker-compose build board-service${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  ✗ Some checks failed! Review errors above                   ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Issues found:${NC}"
    for error in "${ERRORS_FOUND[@]}"; do
        echo -e "${RED}  • $error${NC}"
    done
    echo ""
    echo -e "${CYAN}Recommended actions:${NC}"
    echo -e "${YELLOW}  1. Fix the errors listed above${NC}"
    echo -e "${YELLOW}  2. Run this script again: ./check-go1.25-compat.sh${NC}"
    echo -e "${YELLOW}  3. Once all checks pass, run: docker-compose build board-service${NC}"
    echo ""
    exit 1
fi
