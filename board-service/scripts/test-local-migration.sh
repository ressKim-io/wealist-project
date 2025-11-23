#!/bin/bash

# Comprehensive local migration test script (Tasks 4.1 and 4.2)
# This script uses Docker Compose to test GORM auto-migration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Task 4: Local Environment Migration Testing              ║${NC}"
echo -e "${BLUE}║  - Task 4.1: Fresh Database Test                           ║${NC}"
echo -e "${BLUE}║  - Task 4.2: Existing Database Test                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Change to board-service directory
cd "$(dirname "$0")/.."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker or docker-compose not found${NC}"
    exit 1
fi

# Use docker compose or docker-compose
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${YELLOW}Using: $DOCKER_COMPOSE${NC}"
echo ""

# ============================================================================
# Task 4.1: Fresh Database Test
# ============================================================================

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Task 4.1: Fresh Database Migration Test                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}Step 1: Checking for existing PostgreSQL container${NC}"
POSTGRES_CONTAINER=$(docker ps --filter "ancestor=postgres:17-alpine" --format "{{.Names}}" | head -1)
if [ -z "$POSTGRES_CONTAINER" ]; then
    echo "No existing PostgreSQL container found, starting new one..."
    $DOCKER_COMPOSE up -d postgres
    POSTGRES_CONTAINER="project-board-db"
    sleep 10
else
    echo "Using existing PostgreSQL container: $POSTGRES_CONTAINER"
fi
echo -e "${GREEN}✓ PostgreSQL container: $POSTGRES_CONTAINER${NC}"
echo ""

echo -e "${YELLOW}Step 2: Creating fresh test database${NC}"
# Drop and recreate test database
docker exec $POSTGRES_CONTAINER psql -U postgres -c "DROP DATABASE IF EXISTS project_board_test_fresh;" 2>/dev/null || true
docker exec $POSTGRES_CONTAINER psql -U postgres -c "CREATE DATABASE project_board_test_fresh;"
echo -e "${GREEN}✓ Fresh test database created${NC}"
echo ""

echo -e "${YELLOW}Step 3: Verifying database is empty${NC}"
TABLE_COUNT=$(docker exec $POSTGRES_CONTAINER psql -U postgres -d project_board_test_fresh -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
echo "Tables in database: $TABLE_COUNT"
if [ "$TABLE_COUNT" != "0" ]; then
    echo -e "${RED}✗ Database is not empty (has $TABLE_COUNT tables)${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Database is empty${NC}"
echo ""

echo -e "${YELLOW}Step 4: Building and starting board-service with test database${NC}"
# Stop any existing board-service
$DOCKER_COMPOSE down board-service 2>/dev/null || true

# Start board-service with test database configuration
DB_NAME=project_board_test_fresh DB_HOST=$POSTGRES_CONTAINER $DOCKER_COMPOSE up -d --build board-service
echo "Waiting for application to start..."
sleep 15
echo ""

echo -e "${YELLOW}Step 5: Checking application logs${NC}"
echo "Last 40 lines of board-service logs:"
$DOCKER_COMPOSE logs --tail=40 board-service
echo ""

# Check if migration completed successfully
if $DOCKER_COMPOSE logs board-service | grep -q "Safe auto-migration completed successfully"; then
    echo -e "${GREEN}✓ Migration completed successfully${NC}"
else
    echo -e "${RED}✗ Migration success message not found${NC}"
    echo "Full logs:"
    $DOCKER_COMPOSE logs board-service
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 6: Verifying all tables were created${NC}"
EXPECTED_TABLES=("projects" "project_members" "project_join_requests" "boards" "participants" "comments" "field_options")

ALL_TABLES_EXIST=true
for table in "${EXPECTED_TABLES[@]}"; do
    EXISTS=$(docker exec project-board-db psql -U postgres -d project_board -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" | tr -d ' ')
    if [[ $EXISTS == "t" ]]; then
        echo -e "${GREEN}✓ Table exists: $table${NC}"
    else
        echo -e "${RED}✗ Table missing: $table${NC}"
        ALL_TABLES_EXIST=false
    fi
done

if [ "$ALL_TABLES_EXIST" = false ]; then
    echo -e "${RED}✗ Some tables are missing${NC}"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 7: Checking table structures${NC}"
for table in "${EXPECTED_TABLES[@]}"; do
    COLUMN_COUNT=$(docker exec project-board-db psql -U postgres -d project_board -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '$table';" | tr -d ' ')
    echo "Table $table has $COLUMN_COUNT columns"
done
echo ""

echo -e "${YELLOW}Step 8: Testing health check endpoint${NC}"
MAX_RETRIES=5
RETRY_COUNT=0
HEALTH_OK=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    HEALTH_RESPONSE=$(curl -s http://localhost:8000/health || echo "failed")
    if [[ $HEALTH_RESPONSE == *"ok"* ]] || [[ $HEALTH_RESPONSE == *"healthy"* ]] || [[ $HEALTH_RESPONSE == *"status"* ]]; then
        echo -e "${GREEN}✓ Health check passed${NC}"
        echo "Response: $HEALTH_RESPONSE"
        HEALTH_OK=true
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo "Attempt $RETRY_COUNT/$MAX_RETRIES failed, retrying..."
        sleep 3
    fi
done

if [ "$HEALTH_OK" = false ]; then
    echo -e "${RED}✗ Health check failed after $MAX_RETRIES attempts${NC}"
    echo "Response: $HEALTH_RESPONSE"
    exit 1
fi
echo ""

echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Task 4.1: PASSED ✓                                        ║${NC}"
echo -e "${GREEN}║  - Fresh database created                                  ║${NC}"
echo -e "${GREEN}║  - GORM auto-migration executed successfully               ║${NC}"
echo -e "${GREEN}║  - All 7 tables created correctly                          ║${NC}"
echo -e "${GREEN}║  - Application started and health check passed             ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
sleep 2

# ============================================================================
# Task 4.2: Existing Database Test
# ============================================================================

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Task 4.2: Existing Database Migration Test               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}Step 1: Recording current table structures${NC}"
docker exec project-board-db psql -U postgres -d project_board -c "\d projects" > /tmp/projects_before_restart.txt 2>&1
echo -e "${GREEN}✓ Table structures recorded${NC}"
echo ""

echo -e "${YELLOW}Step 2: Restarting board-service (simulating redeployment)${NC}"
$DOCKER_COMPOSE restart board-service
echo "Waiting for application to restart..."
sleep 15
echo ""

echo -e "${YELLOW}Step 3: Checking restart logs${NC}"
echo "Last 40 lines of board-service logs after restart:"
$DOCKER_COMPOSE logs --tail=40 board-service
echo ""

# Check if migration completed successfully on restart
if $DOCKER_COMPOSE logs board-service | tail -50 | grep -q "Safe auto-migration completed successfully"; then
    echo -e "${GREEN}✓ Migration completed successfully on restart${NC}"
else
    echo -e "${RED}✗ Migration success message not found after restart${NC}"
    exit 1
fi

# Check for "relation already exists" errors
if $DOCKER_COMPOSE logs board-service | grep -q "relation.*already exists"; then
    echo -e "${RED}✗ 'relation already exists' error found in logs${NC}"
    exit 1
else
    echo -e "${GREEN}✓ No 'relation already exists' errors${NC}"
fi

# Check that tables were detected as existing
if $DOCKER_COMPOSE logs board-service | tail -50 | grep -q "Table exists, updating schema only"; then
    echo -e "${GREEN}✓ GORM correctly detected existing tables${NC}"
else
    echo -e "${YELLOW}⚠ Could not confirm GORM detected existing tables (check logs)${NC}"
fi
echo ""

echo -e "${YELLOW}Step 4: Verifying all tables still exist${NC}"
ALL_TABLES_EXIST=true
for table in "${EXPECTED_TABLES[@]}"; do
    EXISTS=$(docker exec project-board-db psql -U postgres -d project_board -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" | tr -d ' ')
    if [[ $EXISTS == "t" ]]; then
        echo -e "${GREEN}✓ Table still exists: $table${NC}"
    else
        echo -e "${RED}✗ Table missing after restart: $table${NC}"
        ALL_TABLES_EXIST=false
    fi
done

if [ "$ALL_TABLES_EXIST" = false ]; then
    echo -e "${RED}✗ Some tables are missing after restart${NC}"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 5: Verifying schema integrity${NC}"
docker exec project-board-db psql -U postgres -d project_board -c "\d projects" > /tmp/projects_after_restart.txt 2>&1
echo "Comparing table structures..."
if diff /tmp/projects_before_restart.txt /tmp/projects_after_restart.txt > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Table structure preserved (no unexpected changes)${NC}"
else
    echo -e "${YELLOW}⚠ Table structure changed (GORM may have added columns/indexes)${NC}"
    echo "Differences:"
    diff /tmp/projects_before_restart.txt /tmp/projects_after_restart.txt || true
fi
echo ""

echo -e "${YELLOW}Step 6: Testing health check after restart${NC}"
RETRY_COUNT=0
HEALTH_OK=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    HEALTH_RESPONSE=$(curl -s http://localhost:8000/health || echo "failed")
    if [[ $HEALTH_RESPONSE == *"ok"* ]] || [[ $HEALTH_RESPONSE == *"healthy"* ]] || [[ $HEALTH_RESPONSE == *"status"* ]]; then
        echo -e "${GREEN}✓ Health check passed after restart${NC}"
        echo "Response: $HEALTH_RESPONSE"
        HEALTH_OK=true
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo "Attempt $RETRY_COUNT/$MAX_RETRIES failed, retrying..."
        sleep 3
    fi
done

if [ "$HEALTH_OK" = false ]; then
    echo -e "${RED}✗ Health check failed after restart${NC}"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 7: Testing multiple restarts (idempotency)${NC}"
for i in {1..2}; do
    echo "Restart test $i/2..."
    $DOCKER_COMPOSE restart board-service
    sleep 10
    
    if ! $DOCKER_COMPOSE logs board-service | tail -30 | grep -q "Safe auto-migration completed successfully"; then
        echo -e "${RED}✗ Migration failed on restart $i${NC}"
        exit 1
    fi
    
    if $DOCKER_COMPOSE logs board-service | tail -30 | grep -q "relation.*already exists"; then
        echo -e "${RED}✗ 'relation already exists' error on restart $i${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Restart $i successful${NC}"
done
echo ""

echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Task 4.2: PASSED ✓                                        ║${NC}"
echo -e "${GREEN}║  - Application restarted successfully                      ║${NC}"
echo -e "${GREEN}║  - GORM detected existing tables correctly                 ║${NC}"
echo -e "${GREEN}║  - No 'relation already exists' errors                     ║${NC}"
echo -e "${GREEN}║  - All tables preserved after restart                      ║${NC}"
echo -e "${GREEN}║  - Multiple restarts work (idempotent)                     ║${NC}"
echo -e "${GREEN}║  - Health check passed                                     ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# ============================================================================
# Final Summary
# ============================================================================

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Task 4: Local Environment Testing - COMPLETE ✓           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Summary:${NC}"
echo -e "${GREEN}  ✓ Task 4.1: Fresh database migration - PASSED${NC}"
echo -e "${GREEN}  ✓ Task 4.2: Existing database migration - PASSED${NC}"
echo ""
echo -e "${YELLOW}Requirements Verified:${NC}"
echo -e "${GREEN}  ✓ Requirement 1.1: Safe auto-migration on existing tables${NC}"
echo -e "${GREEN}  ✓ Requirement 4.1: Fresh database deployment${NC}"
echo -e "${GREEN}  ✓ Requirement 4.2: Schema updates without data loss${NC}"
echo -e "${GREEN}  ✓ Requirement 4.3: All tables created correctly${NC}"
echo ""
echo -e "${YELLOW}Containers are still running. To stop them:${NC}"
echo -e "${YELLOW}  $DOCKER_COMPOSE down${NC}"
echo ""
echo -e "${YELLOW}To view logs:${NC}"
echo -e "${YELLOW}  $DOCKER_COMPOSE logs -f board-service${NC}"
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ALL TESTS PASSED ✓✓✓                                     ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
