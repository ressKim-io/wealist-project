#!/bin/bash

# Script to verify database schema changes for task 18
# This script checks if the new columns and tables exist in the database

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Database Schema Verification Script"
echo "Task 18: board-api-improvements"
echo "=========================================="
echo ""

# Database connection parameters (can be overridden by environment variables)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-project_board}"
DB_USER="${DB_USER:-postgres}"

echo "Connecting to database:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Function to run SQL query and check result
check_column() {
    local table=$1
    local column=$2
    local result=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT column_name FROM information_schema.columns WHERE table_name='$table' AND column_name='$column';" 2>/dev/null | xargs)
    
    if [ "$result" == "$column" ]; then
        echo -e "${GREEN}✓${NC} Column $table.$column exists"
        return 0
    else
        echo -e "${RED}✗${NC} Column $table.$column NOT FOUND"
        return 1
    fi
}

# Function to check if table exists
check_table() {
    local table=$1
    local result=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT table_name FROM information_schema.tables WHERE table_name='$table';" 2>/dev/null | xargs)
    
    if [ "$result" == "$table" ]; then
        echo -e "${GREEN}✓${NC} Table $table exists"
        return 0
    else
        echo -e "${RED}✗${NC} Table $table NOT FOUND"
        return 1
    fi
}

# Function to check if index exists
check_index() {
    local index=$1
    local result=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT indexname FROM pg_indexes WHERE indexname='$index';" 2>/dev/null | xargs)
    
    if [ "$result" == "$index" ]; then
        echo -e "${GREEN}✓${NC} Index $index exists"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} Index $index NOT FOUND (may be created with different name)"
        return 1
    fi
}

echo "Checking schema changes..."
echo ""

# Track overall status
all_passed=true

# Check 1: boards.start_date column
echo "1. Checking boards table for start_date column..."
if ! check_column "boards" "start_date"; then
    all_passed=false
fi
echo ""

# Check 2: projects.start_date and projects.due_date columns
echo "2. Checking projects table for start_date column..."
if ! check_column "projects" "start_date"; then
    all_passed=false
fi
echo ""

echo "3. Checking projects table for due_date column..."
if ! check_column "projects" "due_date"; then
    all_passed=false
fi
echo ""

# Check 3: attachments table
echo "4. Checking if attachments table exists..."
if ! check_table "attachments"; then
    all_passed=false
fi
echo ""

# Check attachments table columns if table exists
if check_table "attachments" &>/dev/null; then
    echo "5. Checking attachments table columns..."
    check_column "attachments" "entity_type"
    check_column "attachments" "entity_id"
    check_column "attachments" "file_name"
    check_column "attachments" "file_url"
    check_column "attachments" "file_size"
    check_column "attachments" "content_type"
    check_column "attachments" "uploaded_by"
    echo ""
fi

# Check indexes
echo "6. Checking indexes..."
check_index "idx_boards_start_date"
check_index "idx_attachments_entity"
check_index "idx_attachments_uploaded_by"
echo ""

# Summary
echo "=========================================="
if [ "$all_passed" = true ]; then
    echo -e "${GREEN}✓ All schema changes verified successfully!${NC}"
    echo ""
    echo "The following changes are in place:"
    echo "  - boards.start_date column"
    echo "  - projects.start_date column"
    echo "  - projects.due_date column"
    echo "  - attachments table with all required columns"
    exit 0
else
    echo -e "${RED}✗ Some schema changes are missing${NC}"
    echo ""
    echo "This is expected if:"
    echo "  1. The application hasn't been started yet"
    echo "  2. GORM auto-migration hasn't run yet"
    echo ""
    echo "To apply the changes:"
    echo "  1. Start the board-service application"
    echo "  2. GORM will automatically create missing columns/tables"
    echo "  3. Run this script again to verify"
    exit 1
fi
