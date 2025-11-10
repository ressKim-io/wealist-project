# Board Service (Go/Gin)

Kanban board API service for weAlist project management platform. This is a migration from Python FastAPI to Go Gin framework.

## Overview

**Board Service** is a microservice that provides Kanban board functionality for the weAlist project management platform. It combines features from Jira, Slack, and Notion to deliver a comprehensive project management experience.

## Tech Stack

- **Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **ORM**: [GORM v2](https://gorm.io) - Go ORM library with PostgreSQL driver
- **Database**: PostgreSQL 16
- **Cache**: Redis 7 (DB 1)
- **Migration**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Logging**: [uber-go/zap](https://github.com/uber-go/zap) - Structured JSON logging
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)
- **JWT**: [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - HS512 algorithm

## Architecture

- **Microservices**: Integrates with User Service (Spring Boot, port 8080)
- **Authentication**: JWT token validation (shared secret with User Service)
- **Database**: PostgreSQL with UUID primary keys, no foreign keys (sharding-ready)
- **Soft Deletes**: Unified `is_deleted` boolean flag in `BaseModel` (inherited by all entities)
  - Exception: `Comment` table uses GORM's `DeletedAt` for historical reasons
- **BaseModel**: All entities inherit from BaseModel with:
  - `ID` (UUID, auto-generated)
  - `CreatedAt` (timestamp, auto-managed)
  - `UpdatedAt` (timestamp, auto-managed)
  - `IsDeleted` (boolean, default false)

## Project Structure

```
board-service/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management (Viper)
│   ├── database/         # PostgreSQL connection (GORM)
│   ├── cache/            # Redis client
│   ├── middleware/       # HTTP middlewares (auth, logger, recovery, cors, request_id)
│   ├── handler/          # HTTP request handlers
│   ├── service/          # Business logic
│   ├── repository/       # Data access layer
│   ├── domain/           # Domain models (entities)
│   ├── dto/              # Data transfer objects
│   ├── apperrors/        # Custom error types
│   └── client/           # External service clients (User Service)
├── pkg/
│   ├── logger/           # Zap logger initialization
│   └── jwt/              # JWT validation utilities
├── migrations/           # Database migration files
├── scripts/              # Utility scripts
├── docker/               # Dockerfile and .dockerignore
├── Makefile              # Build and development commands
├── go.mod                # Go module dependencies
└── README.md             # This file
```

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 17
- Redis 7
- Docker and Docker Compose (for containerized deployment)

## Getting Started

### 1. Install Dependencies

```bash
go mod download
go mod tidy
```

### 2. Configure Environment

Copy `.env.example` to `.env` and update the values:

```bash
cp .env.example .env
```

Key environment variables:
- `ENV`: Environment mode (`dev` or `prod`)
- `USE_AUTO_MIGRATE`: Enable GORM AutoMigrate (`true` for dev, `false` for prod)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `SECRET_KEY`: JWT secret (must match User Service)
- `USER_SERVICE_URL`: User Service base URL

### 3. Database Setup

#### Development (AutoMigrate)

For rapid development, use GORM AutoMigrate:

```bash
# Set USE_AUTO_MIGRATE=true in .env
echo "USE_AUTO_MIGRATE=true" >> .env

# Run server (AutoMigrate runs automatically)
go run cmd/api/main.go

# Or use the init script
./scripts/db/init_dev.sh
```

#### Production (Manual Migrations)

For production, use manual SQL migrations:

```bash
# Apply all pending migrations
./scripts/db/apply_migrations.sh prod

# Rollback specific migration
./scripts/db/rollback.sh prod 20250106120000

# Dump current schema
./scripts/db/dump_schema.sh prod
```

### 4. Run the Application

**Local development:**
```bash
go run cmd/api/main.go
```

**Build and run binary:**
```bash
go build -ldflags="-s -w" -o board-api cmd/api/main.go
./board-api
```

**Docker:**
```bash
docker build -f docker/Dockerfile -t board-service:latest .
docker run -p 8000:8000 --env-file .env board-service:latest
```

### 5. Access Swagger UI

Swagger is available in development mode only:

```bash
# Visit http://localhost:8000/swagger/index.html
open http://localhost:8000/swagger/index.html

# Regenerate Swagger docs (if needed)
swag init -g cmd/api/main.go -o docs
```

## Available Commands

### Development Commands

```bash
# Run service
go run cmd/api/main.go

# Build binary
go build -ldflags="-s -w" -o board-api cmd/api/main.go

# Run tests
go test -v -race -cover ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html

# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run
golangci-lint run --fix

# Tidy dependencies
go mod download
go mod tidy
```

### Database Commands

```bash
# Initialize development database
./scripts/db/init_dev.sh

# Apply migrations (production)
./scripts/db/apply_migrations.sh prod

# Rollback migration
./scripts/db/rollback.sh prod 20250106120000

# Dump schema
./scripts/db/dump_schema.sh dev
```

### Swagger Commands

```bash
# Generate Swagger documentation
swag init -g cmd/api/main.go -o docs

# View Swagger UI (dev mode only)
open http://localhost:8000/swagger/index.html
```

### Docker Commands

```bash
# Build image
docker build -f docker/Dockerfile -t board-service:latest .

# Run container
docker run -p 8000:8000 --env-file .env board-service:latest
```

## API Endpoints

### Health Check (No Authentication)

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-06T10:00:00Z",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

### API Routes (Authentication Required)

All `/api/*` endpoints require JWT authentication via `Authorization: Bearer <token>` header.

#### Projects
- `POST /api/projects` - Create project
- `GET /api/projects/search` - Search projects
- `GET /api/projects/:project_id` - Get project details
- `PUT /api/projects/:project_id` - Update project
- `DELETE /api/projects/:project_id` - Delete project (soft)
- `POST /api/projects/join-requests` - Create join request
- `GET /api/projects/:project_id/join-requests` - List join requests
- `PUT /api/projects/join-requests/:join_request_id` - Approve/reject join request
- `GET /api/projects/:project_id/members` - List project members
- `PUT /api/projects/:project_id/members/:member_id/role` - Update member role
- `DELETE /api/projects/:project_id/members/:member_id` - Remove member

#### Custom Fields
- `POST /api/custom-fields/roles` - Create custom role
- `GET /api/custom-fields/projects/:project_id/roles` - List custom roles
- `GET /api/custom-fields/roles/:role_id` - Get custom role
- `PUT /api/custom-fields/roles/:role_id` - Update custom role
- `DELETE /api/custom-fields/roles/:role_id` - Delete custom role
- `PUT /api/custom-fields/projects/:project_id/roles/order` - Update role order
- Similar endpoints for `/stages` and `/importance`

#### Boards
- `POST /api/boards` - Create board
- `GET /api/boards` - List boards (filters: project_id, stage_id, importance_id, role_id, assignee_id, author_id)
- `GET /api/boards/:board_id` - Get board details
- `PUT /api/boards/:board_id` - Update board
- `DELETE /api/boards/:board_id` - Delete board (soft)

#### User Order (Drag-and-Drop)
- `GET /api/projects/:project_id/orders/role-board` - Get role-based board view
- `GET /api/projects/:project_id/orders/stage-board` - Get stage-based board view
- `PUT /api/projects/:project_id/orders/role-columns` - Update role column order
- `PUT /api/projects/:project_id/orders/stage-columns` - Update stage column order
- `PUT /api/projects/:project_id/orders/role-boards/:role_id` - Update board order in role column
- `PUT /api/projects/:project_id/orders/stage-boards/:stage_id` - Update board order in stage column

#### Comments
- `POST /api/comments` - Create comment
- `GET /api/comments?board_id=<uuid>` - List comments by board
- `PUT /api/comments/:id` - Update comment
- `DELETE /api/comments/:id` - Delete comment

## Authentication

This service validates JWT tokens issued by the User Service. The tokens use HS512 algorithm and must include:

```json
{
  "sub": "user_id",
  "exp": 1234567890
}
```

Token validation errors:
- `401 MISSING_TOKEN`: No Authorization header
- `401 INVALID_TOKEN`: Malformed or invalid token
- `401 TOKEN_EXPIRED`: Token has expired

## Development

### Code Structure Guidelines

1. **No Foreign Keys**: Use application-level relationships, add comments like `comment='References users.id (no FK)'`
2. **UUID Primary Keys**: All entities use `uuid.UUID` type
3. **Error Handling**: All errors use `apperrors.AppError`
4. **Response Format**: Use `dto.Success()`, `dto.Error()`, `dto.Paginated()`
5. **Logging**: Use structured logging with zap

### Database Migration Strategy

**Development**: GORM AutoMigrate for rapid prototyping
**Production**: Manual SQL migrations for full control

#### Creating a New Migration

1. Modify models in `internal/domain/`
2. Test with AutoMigrate: `ENV=dev USE_AUTO_MIGRATE=true go run cmd/api/main.go`
3. Dump schema: `./scripts/db/dump_schema.sh dev`
4. Create migration files (naming: `YYYYMMDDHHMMSS_description.{up|down}.sql`):
   ```sql
   -- migrations/20250110120000_add_attachments.up.sql
   CREATE TABLE IF NOT EXISTS attachments (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       board_id UUID NOT NULL,
       file_name VARCHAR(255) NOT NULL,
       is_deleted BOOLEAN DEFAULT FALSE,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   CREATE INDEX idx_attachments_board_id ON attachments(board_id);
   CREATE INDEX idx_attachments_is_deleted ON attachments(is_deleted);

   INSERT INTO schema_versions (version, description)
   VALUES ('20250110120000', 'Add attachments table');
   ```
5. Create rollback file:
   ```sql
   -- migrations/20250110120000_add_attachments.down.sql
   DROP TABLE IF EXISTS attachments CASCADE;
   DELETE FROM schema_versions WHERE version = '20250110120000';
   ```
6. Test migration: `./scripts/db/apply_migrations.sh dev`
7. Test rollback: `./scripts/db/rollback.sh dev 20250110120000`
8. Update `autoMigrateAll()` in `internal/database/postgres.go`

#### Migration Guidelines

- **No Foreign Keys**: Use comments like `-- References boards.id (no FK for sharding)`
- **UUID Primary Keys**: Always use `UUID DEFAULT gen_random_uuid()`
- **Soft Deletes**: Add `is_deleted BOOLEAN DEFAULT FALSE` and index
- **Idempotent SQL**: Use `IF NOT EXISTS` / `IF EXISTS`
- **Track Versions**: All migrations must insert into `schema_versions` table

### Running Tests

#### Unit Tests
```bash
# Run all tests
go test -v -race -cover ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html

# Open coverage report
open coverage.html
```

#### Integration Tests

The service includes a comprehensive integration test script that validates all major functionality:

```bash
# Run from project root
cd ../scripts/board_test_script
./test_board_service.sh
```

The test script validates:
- ✅ Health Check
- ✅ Token Generation (from User Service)
- ✅ Workspace CRUD operations
- ✅ Project CRUD operations
- ✅ Custom Role creation
- ✅ Board creation and management
- ✅ Stage-based and Role-based board views
- ✅ Board ordering (drag-and-drop functionality)
- ✅ Project member queries
- ✅ Board updates

All tests must pass before deploying to production.

## Docker Deployment

The service uses multi-stage Docker builds for optimized image size.

**Build:**
```bash
docker build -f docker/Dockerfile -t board-service:latest .
```

**Run:**
```bash
docker run -p 8000:8000 \
  -e DATABASE_URL="postgresql://..." \
  -e REDIS_URL="redis://..." \
  -e SECRET_KEY="..." \
  board-service:latest
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment (dev/prod) | `dev` |
| `SERVER_PORT` | HTTP server port | `8000` |
| `DATABASE_URL` | PostgreSQL connection URL | Required |
| `REDIS_URL` | Redis connection URL | Required |
| `SECRET_KEY` | JWT secret key (HS512) | Required |
| `USER_SERVICE_URL` | User Service base URL | Required |
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) | `http://localhost:3000` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |

## Logging

The service uses structured JSON logging in production and console logging in development.

**Log Levels:**
- `debug`: Detailed information for debugging
- `info`: General informational messages
- `warn`: Warning messages
- `error`: Error messages

**Log Fields:**
- `timestamp`: ISO 8601 timestamp
- `level`: Log level
- `msg`: Log message
- `request_id`: Unique request identifier
- `user_id`: Authenticated user ID (if available)
- `error`: Error details (if applicable)

## Features

### Core Functionality
- ✅ **Workspace Management**: Create, search, and manage workspaces
- ✅ **Project Management**: Create projects within workspaces with member management
- ✅ **Custom Fields**: Project-specific roles, stages, and importance levels
- ✅ **Board Management**: Create and organize boards with custom attributes
- ✅ **Drag-and-Drop**: User-specific board ordering with Redis caching
- ✅ **Board Views**: Role-based and Stage-based board views
- ✅ **Comments**: Add comments to boards
- ✅ **Join Requests**: Workspace and project join request workflow
- ✅ **Member Management**: Role-based access control for workspaces and projects
- ✅ **Soft Delete**: All entities support soft deletion for data recovery

### Technical Features
- ✅ **JWT Authentication**: Secure token-based authentication
- ✅ **Redis Caching**: User order caching for performance
- ✅ **Structured Logging**: JSON logs with request tracing
- ✅ **Swagger Documentation**: Interactive API documentation (dev mode)
- ✅ **Health Checks**: Database and Redis connectivity monitoring
- ✅ **CORS Support**: Configurable cross-origin resource sharing
- ✅ **Migration System**: Two-track migration strategy (AutoMigrate + Manual SQL)

## Recent Updates

### v1.2.0 - Fractional Indexing for Board Ordering (2025-01-11)

**Major Changes:**
- **Migrated from integer-based ordering to fractional indexing** for O(1) board position updates
- Replaced `display_order` (int) with `position` (string) in `user_board_order` table
- **Performance improvement**: Moving a board now updates only 1 row instead of N rows
- **New MoveBoard API**: Uses `before_position` and `after_position` instead of numeric index
- Implemented fractional indexing utility (`internal/util/position.go`) using base62 encoding
- Created comprehensive frontend API guide: `docs/FRONTEND_API_GUIDE.md`

**Technical Details:**
- Position strings are lexicographically sortable (e.g., "a0", "a0V", "a1")
- Can always insert between any two positions without reordering other boards
- Migration script automatically converts existing integer orders to string positions
- Eliminated `recalculateBoardOrders` function (no longer needed)

**Benefits:**
- O(1) insert/move operations (was O(N) with cascading updates)
- Eliminates database bottleneck for large board lists
- Same approach used by Linear, Jira, Notion, and Figma
- Simpler frontend implementation (only need before/after positions)

**Documentation:**
- Frontend implementation guide: `docs/FRONTEND_API_GUIDE.md`
- Migration files: `migrations/20250111120000_convert_to_fractional_indexing.{up|down}.sql`
- Utility functions: `internal/util/position.go`

### v1.1.0 - API Field Name Standardization (2025-11-09)

**Major Changes:**
- **Standardized all ID field names to snake_case** for consistency with database schema
- Updated all DTO request/response structs:
  - `projectId` → `project_id`
  - `workspaceId` → `workspace_id`
  - `boardId` → `board_id`
  - `stageId` → `stage_id`
  - `roleId`/`roleIds` → `role_id`/`role_ids`
  - `importanceId` → `importance_id`
  - `assigneeId` → `assignee_id`
  - `userId` → `user_id`
- Response ID fields made explicit:
  - Generic `id` → specific `board_id`, `project_id`, `stage_id`, `role_id`, `importance_id`, `comment_id`
- Updated frontend TypeScript interfaces to match backend schema
- Regenerated Swagger documentation with new field names

**Benefits:**
- Clear and explicit field naming
- Direct alignment with PostgreSQL column names
- Improved API contract clarity
- Easier debugging and maintenance

### v1.0.1 - Soft Delete Unification (2025-01-06)

**Major Changes:**
- Unified all domain models to inherit from `BaseModel` with `is_deleted` field
- Renamed API entity from "Kanban" to "Board" throughout codebase
- Updated endpoints: `/api/kanbans` → `/api/boards`
- Updated user order endpoints: `/stage-kanbans` → `/stage-boards`, `/role-kanbans` → `/role-boards`

**Exception:** `Comment` table continues to use `gorm.DeletedAt` for historical consistency.

## Contributing

1. Create a feature branch from `main`
2. Make your changes
3. Run tests: `go test -v ./...`
4. Run linter: `golangci-lint run`
5. Build to verify: `go build cmd/api/main.go`
6. Run integration tests: `cd ../scripts/board_test_script && ./test_board_service.sh`
7. Submit a pull request

## License

This project is part of the weAlist platform.
