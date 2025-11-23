#!/bin/bash

# Test script for existing database migration (Task 4.2)
# This script tests GORM auto-migration on a database with existing tables

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Task 4.2: Existing Database Migration Test ===${NC}"
echo ""

# Configuration
TEST_DB_NAME="project_board_test_existing"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo -e "${YELLOW}Step 1: Setting up test database with existing tables${NC}"

# Drop and recreate test database
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $TEST_DB_NAME;"
echo -e "${GREEN}✓ Test database created: $TEST_DB_NAME${NC}"
echo ""

echo -e "${YELLOW}Step 2: Creating tables manually (simulating existing database)${NC}"
# Create tables with basic structure (simulating what might exist from SQL migrations)
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME << 'EOF'
-- Create projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create project_members table
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create project_join_requests table
CREATE TABLE project_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create boards table
CREATE TABLE boards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50),
    priority VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create participants table
CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create field_options table
CREATE TABLE field_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    field_type VARCHAR(50) NOT NULL,
    option_value VARCHAR(255) NOT NULL,
    option_label VARCHAR(255) NOT NULL,
    color VARCHAR(50),
    display_order INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);
EOF

echo -e "${GREEN}✓ All 7 tables created manually${NC}"
echo ""

echo -e "${YELLOW}Step 3: Verifying tables exist${NC}"
TABLE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo "Tables in database: $TABLE_COUNT"
if [ "$TABLE_COUNT" -ne "7" ]; then
    echo -e "${RED}✗ Expected 7 tables, found $TABLE_COUNT${NC}"
    exit 1
fi
echo -e "${GREEN}✓ All 7 tables exist${NC}"
echo ""

echo -e "${YELLOW}Step 4: Recording initial table structures${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -c "\d projects" > /tmp/projects_before.txt
echo -e "${GREEN}✓ Initial structures recorded${NC}"
echo ""

echo -e "${YELLOW}Step 5: Building application${NC}"
cd "$(dirname "$0")/.."
go build -o board-api cmd/api/main.go
echo -e "${GREEN}✓ Application built${NC}"
echo ""

echo -e "${YELLOW}Step 6: Running application with GORM auto-migration on existing database${NC}"
echo "This will test if GORM can handle existing tables without errors"
echo ""

# Set environment variables for the test database
export DATABASE_HOST=$DB_HOST
export DATABASE_PORT=$DB_PORT
export DATABASE_USER=$DB_USER
export DATABASE_PASSWORD=$DB_PASSWORD
export DATABASE_DBNAME=$TEST_DB_NAME
export SERVER_PORT=8002
export LOG_LEVEL=info

# Run the application in background
./board-api > /tmp/board-api-existing-test.log 2>&1 &
APP_PID=$!

echo "Application started with PID: $APP_PID"
echo "Waiting for migration to complete..."
sleep 5

# Check if application is still running
if ! kill -0 $APP_PID 2>/dev/null; then
    echo -e "${RED}✗ Application crashed during startup${NC}"
    echo "Last 50 lines of log:"
    tail -50 /tmp/board-api-existing-test.log
    exit 1
fi

echo -e "${GREEN}✓ Application started successfully${NC}"
echo ""

echo -e "${YELLOW}Step 7: Checking migration logs${NC}"
echo "Last 40 lines of application log:"
tail -40 /tmp/board-api-existing-test.log
echo ""

# Check for migration success message
if grep -q "Safe auto-migration completed successfully" /tmp/board-api-existing-test.log; then
    echo -e "${GREEN}✓ Migration completed successfully${NC}"
else
    echo -e "${RED}✗ Migration success message not found in logs${NC}"
    kill $APP_PID 2>/dev/null || true
    exit 1
fi

# Check that no "relation already exists" errors occurred
if grep -q "relation.*already exists" /tmp/board-api-existing-test.log; then
    echo -e "${RED}✗ 'relation already exists' error found in logs${NC}"
    kill $APP_PID 2>/dev/null || true
    exit 1
else
    echo -e "${GREEN}✓ No 'relation already exists' errors${NC}"
fi

# Check that tables were detected as existing
if grep -q "Table exists, updating schema only" /tmp/board-api-existing-test.log; then
    echo -e "${GREEN}✓ GORM correctly detected existing tables${NC}"
else
    echo -e "${YELLOW}⚠ Could not confirm GORM detected existing tables (check logs)${NC}"
fi
echo ""

echo -e "${YELLOW}Step 8: Verifying all tables still exist${NC}"
EXPECTED_TABLES=("projects" "project_members" "project_join_requests" "boards" "participants" "comments" "field_options")

for table in "${EXPECTED_TABLES[@]}"; do
    EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');")
    if [[ $EXISTS == *"t"* ]]; then
        echo -e "${GREEN}✓ Table still exists: $table${NC}"
    else
        echo -e "${RED}✗ Table missing after migration: $table${NC}"
        kill $APP_PID 2>/dev/null || true
        exit 1
    fi
done
echo ""

echo -e "${YELLOW}Step 9: Checking if schema was updated (not recreated)${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -c "\d projects" > /tmp/projects_after.txt
echo "Comparing table structures..."
if diff /tmp/projects_before.txt /tmp/projects_after.txt > /dev/null; then
    echo -e "${YELLOW}⚠ Table structure unchanged (expected if GORM models match exactly)${NC}"
else
    echo -e "${GREEN}✓ Table structure updated (GORM added missing columns/indexes)${NC}"
fi
echo ""

echo -e "${YELLOW}Step 10: Testing health check endpoint${NC}"
HEALTH_RESPONSE=$(curl -s http://localhost:8002/health)
if [[ $HEALTH_RESPONSE == *"ok"* ]] || [[ $HEALTH_RESPONSE == *"healthy"* ]]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "Response: $HEALTH_RESPONSE"
else
    echo -e "${RED}✗ Health check failed${NC}"
    echo "Response: $HEALTH_RESPONSE"
    kill $APP_PID 2>/dev/null || true
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 11: Testing application restart (idempotency)${NC}"
echo "Stopping application..."
kill $APP_PID 2>/dev/null || true
sleep 2

echo "Restarting application..."
./board-api > /tmp/board-api-existing-test-restart.log 2>&1 &
APP_PID=$!
sleep 5

if ! kill -0 $APP_PID 2>/dev/null; then
    echo -e "${RED}✗ Application crashed on restart${NC}"
    echo "Last 50 lines of log:"
    tail -50 /tmp/board-api-existing-test-restart.log
    exit 1
fi

echo -e "${GREEN}✓ Application restarted successfully${NC}"

# Check restart logs
if grep -q "Safe auto-migration completed successfully" /tmp/board-api-existing-test-restart.log; then
    echo -e "${GREEN}✓ Migration completed successfully on restart${NC}"
else
    echo -e "${RED}✗ Migration failed on restart${NC}"
    kill $APP_PID 2>/dev/null || true
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 12: Cleaning up${NC}"
kill $APP_PID 2>/dev/null || true
sleep 2
echo -e "${GREEN}✓ Application stopped${NC}"
echo ""

echo -e "${GREEN}=== Task 4.2 Test Summary ===${NC}"
echo -e "${GREEN}✓ Existing database with 7 tables created${NC}"
echo -e "${GREEN}✓ GORM auto-migration executed without errors${NC}"
echo -e "${GREEN}✓ No 'relation already exists' errors${NC}"
echo -e "${GREEN}✓ All tables preserved after migration${NC}"
echo -e "${GREEN}✓ Application restarted successfully (idempotent)${NC}"
echo -e "${GREEN}✓ Health check passed${NC}"
echo ""
echo -e "${YELLOW}Test database '$TEST_DB_NAME' has been left intact for inspection.${NC}"
echo -e "${YELLOW}To drop it, run: PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c 'DROP DATABASE $TEST_DB_NAME;'${NC}"
echo ""
echo -e "${GREEN}=== Task 4.2: PASSED ===${NC}"
