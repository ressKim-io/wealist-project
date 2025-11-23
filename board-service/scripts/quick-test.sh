#!/bin/bash
# =============================================================================
# Quick Test Script for Board Service
# =============================================================================
# Usage: ./quick-test.sh [endpoint]
# Examples:
#   ./quick-test.sh projects          # Test GET /api/projects
#   ./quick-test.sh create-project    # Test POST /api/projects
#   ./quick-test.sh boards            # Test GET /api/projects/{id}/boards
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BOARD_API_URL="${BOARD_API_URL:-http://localhost:8000}"
USER_API_URL="${USER_API_URL:-http://localhost:8080}"
TEST_USER_EMAIL="user1@example.com"

# Get test user info
echo -e "${BLUE}🔍 Fetching test user info...${NC}"

# Get user ID from database
# Try different possible container names and database configurations
for CONTAINER_NAME in "wealist-postgres" "postgres" "wealist_postgres"; do
  for DB_USER in "wealist_user" "wealist" "postgres"; do
    for DB_NAME in "wealist_user_db" "wealist" "postgres"; do
      USER_ID=$(docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -t -c "SELECT user_id FROM users WHERE email = '${TEST_USER_EMAIL}' LIMIT 1;" 2>/dev/null | tr -d ' \n')
      if [ -n "$USER_ID" ]; then
        break 3
      fi
    done
  done
done

if [ -z "$USER_ID" ]; then
  # Fallback: get any active user
  for CONTAINER_NAME in "wealist-postgres" "postgres" "wealist_postgres"; do
    for DB_USER in "wealist_user" "wealist" "postgres"; do
      for DB_NAME in "wealist_user_db" "wealist" "postgres"; do
        USER_ID=$(docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -t -c "SELECT user_id FROM users WHERE is_active = true LIMIT 1;" 2>/dev/null | tr -d ' \n')
        if [ -n "$USER_ID" ]; then
          break 3
        fi
      done
    done
  done
fi

if [ -z "$USER_ID" ]; then
  echo -e "${RED}❌ Failed to get user ID${NC}"
  echo -e "${YELLOW}💡 Make sure PostgreSQL and user-service are running${NC}"
  exit 1
fi

# Get workspace from database
WORKSPACE_ID=$(docker exec wealist-postgres psql -U wealist_user -d wealist_user_db -t -c "SELECT workspace_id FROM workspace_members WHERE user_id = '${USER_ID}' AND is_default = true LIMIT 1;" 2>/dev/null | tr -d ' \n')

if [ -z "$WORKSPACE_ID" ]; then
  # Fallback: get any workspace for this user
  WORKSPACE_ID=$(docker exec wealist-postgres psql -U wealist_user -d wealist_user_db -t -c "SELECT workspace_id FROM workspace_members WHERE user_id = '${USER_ID}' LIMIT 1;" 2>/dev/null | tr -d ' \n')
fi

if [ -z "$WORKSPACE_ID" ]; then
  echo -e "${RED}❌ Failed to get workspace ID${NC}"
  exit 1
fi

echo -e "${GREEN}✅ User ID: ${USER_ID}${NC}"
echo -e "${GREEN}✅ Workspace ID: ${WORKSPACE_ID}${NC}"

# Get test token
TEST_TOKEN=$(curl -s "${USER_API_URL}/api/users/test/${USER_ID}")
echo -e "${GREEN}✅ Test token generated${NC}"
echo ""

# Determine which test to run
TEST_TYPE="${1:-projects}"

case "$TEST_TYPE" in
  projects)
    echo -e "${YELLOW}Testing: GET /api/projects${NC}"
    curl -v "${BOARD_API_URL}/api/projects?workspaceId=${WORKSPACE_ID}" \
      -H "Authorization: Bearer ${TEST_TOKEN}"
    ;;

  create-project)
    echo -e "${YELLOW}Testing: POST /api/projects${NC}"
    curl -v -X POST "${BOARD_API_URL}/api/projects" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer ${TEST_TOKEN}" \
      -d "{
        \"workspaceId\": \"${WORKSPACE_ID}\",
        \"projectName\": \"Quick Test Project $(date +%s)\",
        \"projectDescription\": \"Created by quick-test.sh\",
        \"projectColor\": \"#3B82F6\"
      }"
    ;;

  boards)
    echo -e "${YELLOW}Testing: GET /api/projects/{id}/boards${NC}"
    # Get first project
    PROJECTS=$(curl -s "${BOARD_API_URL}/api/projects?workspaceId=${WORKSPACE_ID}" \
      -H "Authorization: Bearer ${TEST_TOKEN}")
    PROJECT_ID=$(echo "$PROJECTS" | grep -o '"projectId":"[^"]*' | head -1 | cut -d'"' -f4)
    
    if [ -z "$PROJECT_ID" ]; then
      echo -e "${RED}❌ No projects found. Create one first.${NC}"
      exit 1
    fi
    
    echo -e "${GREEN}Testing with Project ID: ${PROJECT_ID}${NC}"
    curl -v "${BOARD_API_URL}/api/projects/${PROJECT_ID}/boards" \
      -H "Authorization: Bearer ${TEST_TOKEN}"
    ;;

  user-check)
    echo -e "${YELLOW}Testing: User Service Connectivity${NC}"
    echo -e "${BLUE}1. Direct user-service call:${NC}"
    curl -v "${USER_API_URL}/api/users/${USER_ID}"
    echo ""
    echo -e "${BLUE}2. From board-service (triggers internal call):${NC}"
    curl -v "${BOARD_API_URL}/api/projects?workspaceId=${WORKSPACE_ID}" \
      -H "Authorization: Bearer ${TEST_TOKEN}"
    ;;

  *)
    echo -e "${RED}Unknown test type: ${TEST_TYPE}${NC}"
    echo ""
    echo -e "${YELLOW}Available tests:${NC}"
    echo "  projects         - GET /api/projects"
    echo "  create-project   - POST /api/projects"
    echo "  boards           - GET /api/projects/{id}/boards"
    echo "  user-check       - Test user-service connectivity"
    exit 1
    ;;
esac

echo ""
echo -e "${GREEN}✅ Test completed${NC}"
