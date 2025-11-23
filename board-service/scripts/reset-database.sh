#!/bin/bash

# Database Reset Script for Board Service
# This script drops and recreates the board service database
# Use this when switching from SQL migrations to GORM auto-migration

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-project_board}"
DB_PASSWORD="${DB_PASSWORD:-password}"

echo -e "${YELLOW}=== Board Service Database Reset ===${NC}"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "User: $DB_USER"
echo "Database: $DB_NAME"
echo ""

# Confirm before proceeding
read -p "This will DELETE ALL DATA in the database. Are you sure? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo -e "${RED}Aborted.${NC}"
    exit 1
fi

echo -e "${YELLOW}Connecting to PostgreSQL...${NC}"

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Drop existing database
echo -e "${YELLOW}Dropping existing database '$DB_NAME'...${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database dropped successfully${NC}"
else
    echo -e "${RED}✗ Failed to drop database${NC}"
    exit 1
fi

# Create new database
echo -e "${YELLOW}Creating new database '$DB_NAME'...${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;" 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database created successfully${NC}"
else
    echo -e "${RED}✗ Failed to create database${NC}"
    exit 1
fi

# Verify database exists
echo -e "${YELLOW}Verifying database...${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "\l" | grep "$DB_NAME" > /dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database verified${NC}"
else
    echo -e "${RED}✗ Database verification failed${NC}"
    exit 1
fi

# Unset password
unset PGPASSWORD

echo ""
echo -e "${GREEN}=== Database Reset Complete ===${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Start the board-service application"
echo "2. GORM auto-migration will create all tables automatically"
echo ""
