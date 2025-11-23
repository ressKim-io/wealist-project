#!/bin/bash

# Script to clean up boards with nil UUID author_id
# This script runs the Go cleanup utility

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Board Service - Nil Author ID Cleanup Utility"
echo "=============================================="
echo ""

# Check if .env.test exists and load it
if [ -f "$PROJECT_ROOT/.env.test" ]; then
    echo "Loading database configuration from .env.test..."
    export $(grep -v '^#' "$PROJECT_ROOT/.env.test" | xargs)
else
    echo "Warning: .env.test not found. Using default database configuration."
    echo "You can set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME environment variables."
fi

echo ""
echo "Database Configuration:"
echo "  Host: ${DB_HOST:-localhost}"
echo "  Port: ${DB_PORT:-5432}"
echo "  User: ${DB_USER:-postgres}"
echo "  Database: ${DB_NAME:-project_board}"
echo ""

# Run the cleanup script
cd "$PROJECT_ROOT"
go run scripts/cleanup-nil-author-ids.go
