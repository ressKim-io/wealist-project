#!/bin/bash

# EC2 Database Reset Script for Board Service
# This script drops and recreates the board service database on EC2
# Use this when switching from SQL migrations to GORM auto-migration

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== EC2 Board Service Database Reset ===${NC}"
echo "📅 Started at: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# =============================================================================
# Load environment variables from Parameter Store
# =============================================================================
echo "📥 Loading database credentials from Parameter Store..."

PARAMETER_PREFIX="/wealist/dev"
AWS_REGION="ap-northeast-2"

load_param() {
    local param_name="$1"
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    
    echo "  ⏳ Loading: ${param_name}"
    
    value=$(aws ssm get-parameter \
        --name "${full_param_path}" \
        --query 'Parameter.Value' \
        --output text \
        --region ${AWS_REGION} 2>&1)
    
    if [ $? -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
        echo -e "${RED}  ❌ Failed to load parameter: ${param_name}${NC}" >&2
        exit 1
    fi
    
    echo "  ✅ Loaded: ${param_name}"
    echo "$value"
}

load_secret() {
    local param_name="$1"
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    
    echo "  ⏳ Loading secret: ${param_name}"
    
    value=$(aws ssm get-parameter \
        --name "${full_param_path}" \
        --with-decryption \
        --query 'Parameter.Value' \
        --output text \
        --region ${AWS_REGION} 2>&1)
    
    if [ $? -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
        echo -e "${RED}  ❌ Failed to load secret: ${param_name}${NC}" >&2
        exit 1
    fi
    
    echo "  ✅ Loaded: ${param_name} (encrypted)"
    echo "$value"
}

# Load database credentials
POSTGRES_SUPERUSER=$(load_param "db/postgres_superuser")
POSTGRES_SUPERUSER_PASSWORD=$(load_secret "db/postgres_superuser_password")
BOARD_DB_NAME=$(load_param "db/board_db_name")
BOARD_DB_USER=$(load_param "db/board_db_user")
BOARD_DB_PASSWORD=$(load_secret "db/board_db_password")

echo ""
echo "✅ Credentials loaded successfully"
echo "📋 Configuration:"
echo "   - Database: ${BOARD_DB_NAME}"
echo "   - User: ${BOARD_DB_USER}"
echo "   - Superuser: ${POSTGRES_SUPERUSER}"
echo ""

# =============================================================================
# Confirm before proceeding
# =============================================================================
echo -e "${RED}⚠️  WARNING: This will DELETE ALL DATA in the database!${NC}"
echo -e "${YELLOW}Database: ${BOARD_DB_NAME}${NC}"
echo ""
read -p "Are you sure you want to continue? Type 'yes' to proceed: " confirm

if [ "$confirm" != "yes" ]; then
    echo -e "${YELLOW}Aborted by user.${NC}"
    exit 0
fi

echo ""
echo -e "${YELLOW}Proceeding with database reset...${NC}"
echo ""

# =============================================================================
# Database connection settings
# =============================================================================
DB_HOST="localhost"
DB_PORT="5432"

# Check if PostgreSQL container is running
echo "🔍 Checking PostgreSQL container..."
if ! docker ps --format '{{.Names}}' | grep -q wealist-postgres; then
    echo -e "${RED}❌ PostgreSQL container is not running!${NC}"
    echo "Please start the PostgreSQL container first:"
    echo "  docker-compose -f /home/ubuntu/wealist/docker/compose/docker-compose.ec2-dev.yml up -d postgres"
    exit 1
fi
echo "✅ PostgreSQL container is running"
echo ""

# Export password for psql
export PGPASSWORD="$POSTGRES_SUPERUSER_PASSWORD"

# =============================================================================
# Terminate existing connections
# =============================================================================
echo "🔌 Terminating existing connections to database '${BOARD_DB_NAME}'..."
docker exec wealist-postgres psql -U "$POSTGRES_SUPERUSER" -d postgres -c "
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '${BOARD_DB_NAME}'
  AND pid <> pg_backend_pid();
" 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Existing connections terminated"
else
    echo -e "${YELLOW}⚠️  Could not terminate connections (database may not exist yet)${NC}"
fi
echo ""

# =============================================================================
# Drop existing database
# =============================================================================
echo "🗑️  Dropping existing database '${BOARD_DB_NAME}'..."
docker exec wealist-postgres psql -U "$POSTGRES_SUPERUSER" -d postgres -c "DROP DATABASE IF EXISTS ${BOARD_DB_NAME};" 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Database dropped successfully"
else
    echo -e "${RED}❌ Failed to drop database${NC}"
    exit 1
fi
echo ""

# =============================================================================
# Create new database
# =============================================================================
echo "📦 Creating new database '${BOARD_DB_NAME}'..."
docker exec wealist-postgres psql -U "$POSTGRES_SUPERUSER" -d postgres -c "CREATE DATABASE ${BOARD_DB_NAME};" 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Database created successfully"
else
    echo -e "${RED}❌ Failed to create database${NC}"
    exit 1
fi
echo ""

# =============================================================================
# Grant privileges to board user
# =============================================================================
echo "🔐 Granting privileges to user '${BOARD_DB_USER}'..."
docker exec wealist-postgres psql -U "$POSTGRES_SUPERUSER" -d postgres -c "
GRANT ALL PRIVILEGES ON DATABASE ${BOARD_DB_NAME} TO ${BOARD_DB_USER};
" 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Privileges granted successfully"
else
    echo -e "${YELLOW}⚠️  Could not grant privileges (user may not exist yet)${NC}"
fi
echo ""

# =============================================================================
# Verify database exists
# =============================================================================
echo "🔍 Verifying database..."
docker exec wealist-postgres psql -U "$POSTGRES_SUPERUSER" -d postgres -c "\l" | grep "$BOARD_DB_NAME" > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Database verified"
else
    echo -e "${RED}❌ Database verification failed${NC}"
    exit 1
fi
echo ""

# Unset password
unset PGPASSWORD

echo ""
echo -e "${GREEN}=== Database Reset Complete ===${NC}"
echo "📅 Finished at: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Restart the board-service container:"
echo "   docker-compose -f /home/ubuntu/wealist/docker/compose/docker-compose.ec2-dev.yml restart board-service"
echo ""
echo "2. Check the logs to verify GORM auto-migration:"
echo "   docker logs -f wealist-board-service"
echo ""
echo "3. Verify health check:"
echo "   curl http://localhost:8000/health"
echo ""
