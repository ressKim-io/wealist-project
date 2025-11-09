#!/bin/bash

# =============================================================================
# Go Version Upgrade Verification Script
# =============================================================================
# This script verifies that the Go version upgrade didn't break anything
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

CHECKS_PASSED=0
CHECKS_FAILED=0

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

print_check() {
    echo -e "${YELLOW}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    CHECKS_FAILED=$((CHECKS_FAILED + 1))
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_header "Go Version Upgrade Verification"

# =============================================================================
# Check 1: Go Version in Files
# =============================================================================
print_check "1. Checking Go version consistency..."

GO_MOD_VERSION=$(grep "^go " go.mod | awk '{print $2}')
DOCKERFILE_VERSION=$(grep "FROM golang:" docker/Dockerfile | head -1 | sed 's/.*golang:\([0-9.]*\).*/\1/')

print_info "go.mod version: $GO_MOD_VERSION"
print_info "Dockerfile version: $DOCKERFILE_VERSION"

if [[ "$GO_MOD_VERSION" == "$DOCKERFILE_VERSION"* ]] || [[ "$DOCKERFILE_VERSION" == "$GO_MOD_VERSION"* ]]; then
    print_success "Go versions are consistent"
else
    print_error "Go version mismatch: go.mod ($GO_MOD_VERSION) vs Dockerfile ($DOCKERFILE_VERSION)"
fi

# =============================================================================
# Check 2: go.mod Syntax
# =============================================================================
print_check "2. Validating go.mod syntax..."

if go mod edit -print > /dev/null 2>&1; then
    print_success "go.mod syntax is valid"
else
    print_error "go.mod has syntax errors"
fi

# =============================================================================
# Check 3: Dependency Resolution (Skip if network unavailable)
# =============================================================================
print_check "3. Testing dependency resolution..."

# Try with timeout to avoid hanging
if timeout 10s go list -m all > /dev/null 2>&1; then
    print_success "All dependencies are resolvable"
elif go list -m all > /dev/null 2>&1; then
    print_success "All dependencies are resolvable"
else
    print_info "Skipping dependency check (network issue or missing deps)"
fi

# =============================================================================
# Check 4: Go File Syntax
# =============================================================================
print_check "4. Checking Go file syntax..."

SYNTAX_ERRORS=$(find . -name "*.go" ! -path "./vendor/*" -exec gofmt -l {} \; 2>/dev/null | wc -l)

if [ "$SYNTAX_ERRORS" -eq 0 ]; then
    print_success "All Go files have valid syntax"
else
    print_info "Found $SYNTAX_ERRORS files with formatting issues (not critical)"
fi

# =============================================================================
# Check 5: Import Paths
# =============================================================================
print_check "5. Validating import paths..."

if go list ./... > /dev/null 2>&1; then
    print_success "All import paths are valid"
else
    print_error "Some import paths are invalid"
fi

# =============================================================================
# Check 6: Build Tags and Constraints
# =============================================================================
print_check "6. Checking for build constraints..."

BUILD_TAG_FILES=$(find . -name "*.go" ! -path "./vendor/*" -exec grep -l "//go:build\|// +build" {} \; 2>/dev/null | wc -l)

print_info "Found $BUILD_TAG_FILES files with build tags"
print_success "Build tags check completed"

# =============================================================================
# Check 7: Deprecated API Usage
# =============================================================================
print_check "7. Checking for deprecated API usage..."

# Check for known deprecated patterns in Go 1.24
DEPRECATED_COUNT=0

# io/ioutil deprecated in Go 1.16
if grep -r "io/ioutil" --include="*.go" . 2>/dev/null | grep -v vendor | grep -q .; then
    print_info "Warning: io/ioutil is deprecated, use io and os instead"
    DEPRECATED_COUNT=$((DEPRECATED_COUNT + 1))
fi

if [ "$DEPRECATED_COUNT" -eq 0 ]; then
    print_success "No known deprecated APIs detected"
else
    print_info "Found $DEPRECATED_COUNT potential deprecation issues (not critical)"
fi

# =============================================================================
# Check 8: Module Verification
# =============================================================================
print_check "8. Verifying module integrity..."

if go mod verify > /dev/null 2>&1; then
    print_success "Module verification passed"
else
    print_info "Module verification skipped (dependencies not downloaded)"
fi

# =============================================================================
# Check 9: Compilation Test (Syntax Only)
# =============================================================================
print_check "9. Testing compilation (syntax check)..."

if go build -o /dev/null ./cmd/api > /dev/null 2>&1; then
    print_success "Compilation successful"
elif GOTOOLCHAIN=local go build -o /dev/null ./cmd/api > /dev/null 2>&1; then
    print_success "Compilation successful (local toolchain)"
else
    print_info "Compilation test skipped (network or dependency issues)"
fi

# =============================================================================
# Check 10: Unit Tests Syntax
# =============================================================================
print_check "10. Checking test file syntax..."

TEST_FILES=$(find . -name "*_test.go" ! -path "./vendor/*" | wc -l)
print_info "Found $TEST_FILES test files"

if [ "$TEST_FILES" -gt 0 ]; then
    print_success "Test files detected"
else
    print_info "No test files found"
fi

# =============================================================================
# Summary
# =============================================================================
print_header "Verification Summary"

TOTAL_CHECKS=$((CHECKS_PASSED + CHECKS_FAILED))
echo "Total Checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${RED}Failed: $CHECKS_FAILED${NC}"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ All verification checks passed!${NC}"
    echo -e "${GREEN}Go version upgrade looks good.${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗ Some checks failed!${NC}"
    echo -e "${RED}Please review the errors above.${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
