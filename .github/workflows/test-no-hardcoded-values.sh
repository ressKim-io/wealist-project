#!/bin/bash

# =============================================================================
# Property-Based Test: No hardcoded infrastructure values in workflows
# Feature: cd-workflow-improvement, Property 1
# Validates: Requirements 2.3
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKFLOW_DIR="${SCRIPT_DIR}"
EXIT_CODE=0

echo "🧪 Property Test: No hardcoded infrastructure values in workflows"
echo "=================================================================="
echo ""

# Files to check
WORKFLOW_FILES=$(find "${WORKFLOW_DIR}" -name "*.yml" -o -name "*.yaml")

if [ -z "$WORKFLOW_FILES" ]; then
    echo "❌ No workflow files found in ${WORKFLOW_DIR}"
    exit 1
fi

echo "📁 Checking workflow files:"
for file in $WORKFLOW_FILES; do
    echo "  - $(basename "$file")"
done
echo ""

# Check each file for hardcoded values
for file in $WORKFLOW_FILES; do
    filename=$(basename "$file")
    echo "🔍 Checking: ${filename}"
    
    # Skip test files
    if [[ "$filename" == *"test"* ]]; then
        echo "  ⏭️  Skipping test file"
        echo ""
        continue
    fi
    
    file_has_issues=false
    
    # Check for AWS Account ID (12-digit number)
    if grep -nE '[^a-zA-Z0-9_-][0-9]{12}[^a-zA-Z0-9_-]' "$file" | grep -v 'secrets\.' | grep -v 'Parameter' | grep -v '#'; then
        echo "  ❌ Found hardcoded AWS Account ID"
        file_has_issues=true
        EXIT_CODE=1
    fi
    
    # Check for EC2 Instance ID
    if grep -nE 'i-[0-9a-f]{8,17}' "$file" | grep -v 'secrets\.' | grep -v 'Parameter' | grep -v '#'; then
        echo "  ❌ Found hardcoded EC2 Instance ID"
        file_has_issues=true
        EXIT_CODE=1
    fi
    
    # Check for hardcoded AWS regions in env: section at the top level
    # Allow regions in:
    # - Comments
    # - Parameter names
    # - Bootstrap region for initial AWS credentials (needed to access Parameter Store)
    # - aws-actions/configure-aws-credentials step (needed for initial authentication)
    if grep -nE '^\s*AWS_REGION:\s*(ap-northeast-2|us-east-1|us-west-2|eu-west-1)' "$file" | grep -v '#' | grep -v 'Parameter' | grep -v 'Bootstrap'; then
        echo "  ❌ Found hardcoded AWS Region in top-level env section"
        file_has_issues=true
        EXIT_CODE=1
    fi
    
    # Note: aws-region in configure-aws-credentials is allowed as it's needed for initial authentication
    
    if [ "$file_has_issues" = false ]; then
        echo "  ✅ No hardcoded values found"
    fi
    
    echo ""
done

echo "=================================================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Property Test PASSED: No hardcoded infrastructure values found"
    echo ""
    echo "All workflow files properly use:"
    echo "  - Parameter Store for configuration"
    echo "  - GitHub Secrets for credentials"
    echo "  - Environment variables for dynamic values"
else
    echo "❌ Property Test FAILED: Hardcoded infrastructure values detected"
    echo ""
    echo "Hardcoded values found in workflow files."
    echo "Please replace them with:"
    echo "  - Parameter Store references"
    echo "  - GitHub Secrets"
    echo "  - Environment variables"
fi

exit $EXIT_CODE
