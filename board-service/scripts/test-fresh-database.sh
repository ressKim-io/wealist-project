#!/bin/bash

# Test script for fresh database migration (Task 4.1)
# This script tests GORM auto-migration on a completely empty database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Task 4.1: Fresh Database Migration Test ===${NC}"
echo ""

# Configuration
TEST_DB_NAME="project_board_test_fresh"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo -e "${YELLOW}Step 1: Dropping existing test database (if exists)${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true
echo -e "${GREEN}✓ Existing test database dropped${NC}"
echo ""

echo -e "${YELLOW}Step 2: Creating fresh test database${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $TEST_DB_NAME;"
echo -e "${GREEN}✓ Fresh test database created: $TEST_DB_NAME${NC}"
echo ""

echo -e "${YELLOW}Step 3: Verifying database is empty${NC}"
TABLE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo "Tables in database: $TABLE_COUNT"
if [ "$TABLE_COUNT" -ne "0" ]; then
    echo -e "${RED}✗ Database is not empty!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Database is empty${NC}"
echo ""

echo -e "${YELLOW}Step 4: Building application${NC}"
cd "$(dirname "$0")/.."
go build -o board-api cmd/api/main.go
echo -e "${GREEN}✓ Application built${NC}"
echo ""

echo -e "${YELLOW}Step 5: Running application with GORM auto-migration${NC}"
echo "This will start the application and run auto-migration on the fresh database"
echo "Press Ctrl+C after you see 'Safe auto-migration completed successfully'"
echo ""

# Set environment variables for the test database
export DATABASE_HOST=$DB_HOST
export DATABASE_PORT=$DB_PORT
export DATABASE_USER=$DB_USER
export DATABASE_PASSWORD=$DB_PASSWORD
export DATABASE_DBNAME=$TEST_DB_NAME
export SERVER_PORT=8001
export LOG_LEVEL=info

# Run the application in background
./board-api > /tmp/board-api-fresh-test.log 2>&1 &
APP_PID=$!

echo "Application started with PID: $APP_PID"
echo "Waiting for migration to complete..."
sleep 5

# Check if application is still running
if ! kill -0 $APP_PID 2>/dev/null; then
    echo -e "${RED}✗ Application crashed during startup${NC}"
    echo "Last 50 lines of log:"
    tail -50 /tmp/board-api-fresh-test.log
    exit 1
fi

echo -e "${GREEN}✓ Application started successfully${NC}"
echo ""

echo -e "${YELLOW}Step 6: Checking migration logs${NC}"
echo "Last 30 lines of application log:"
tail -30 /tmp/board-api-fresh-test.log
echo ""

# Check for migration success message
if grep -q "Safe auto-migration completed successfully" /tmp/board-api-fresh-test.log; then
    echo -e "${GREEN}✓ Migration completed successfully${NC}"
else
    echo -e "${RED}✗ Migration success message not found in logs${NC}"
    kill $APP_PID 2>/dev/null || true
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 7: Verifying all tables were created${NC}"
EXPECTED_TABLES=("projects" "project_members" "project_join_requests" "boards" "participants" "comments" "field_options")

for table in "${EXPECTED_TABLES[@]}"; do
    EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');")
    if [[ $EXISTS == *"t"* ]]; then
        echo -e "${GREEN}✓ Table exists: $table${NC}"
    else
        echo -e "${RED}✗ Table missing: $table${NC}"
        kill $APP_PID 2>/dev/null || true
        exit 1
    fi
done
echo ""

echo -e "${YELLOW}Step 8: Checking table structures${NC}"
for table in "${EXPECTED_TABLES[@]}"; do
    COLUMN_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $TEST_DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '$table';")
    echo "Table $table has $COLUMN_COUNT columns"
done
echo ""

echo -e "${YELLOW}Step 9: Testing health check endpoint${NC}"
HEALTH_RESPONSE=$(curl -s http://localhost:8001/health)
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

echo -e "${YELLOW}Step 10: Cleaning up${NC}"
kill $APP_PID 2>/dev/null || true
sleep 2
echo -e "${GREEN}✓ Application stopped${NC}"
echo ""

echo -e "${GREEN}=== Task 4.1 Test Summary ===${NC}"
echo -e "${GREEN}✓ Fresh database created${NC}"
echo -e "${GREEN}✓ GORM auto-migration executed successfully${NC}"
echo -e "${GREEN}✓ All 7 tables created correctly${NC}"
echo -e "${GREEN}✓ Application started and health check passed${NC}"
echo ""
echo -e "${YELLOW}Test database '$TEST_DB_NAME' has been left intact for inspection.${NC}"
echo -e "${YELLOW}To drop it, run: PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c 'DROP DATABASE $TEST_DB_NAME;'${NC}"
echo ""
echo -e "${GREEN}=== Task 4.1: PASSED ===${NC}"
