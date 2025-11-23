# Configuration Guide

This document describes all configuration options for the Project Board Management API.

## Table of Contents

- [Overview](#overview)
- [Configuration Files](#configuration-files)
- [Environment Variables](#environment-variables)
- [Configuration Reference](#configuration-reference)
- [Examples](#examples)

## Overview

The application supports two methods of configuration:

1. **YAML Configuration File** (`configs/config.yaml`)
2. **Environment Variables** (`.env` file or system environment)

Environment variables take precedence over YAML configuration, allowing you to override specific settings without modifying the config file.

## Configuration Files

### config.yaml

The main configuration file located at `configs/config.yaml`. This file contains all application settings in YAML format.

**Setup:**
```bash
cp configs/config.yaml.example configs/config.yaml
# Edit config.yaml with your settings
```

### .env

Optional environment variables file. Useful for local development and Docker deployments.

**Setup:**
```bash
cp .env.example .env
# Edit .env with your settings
```

## Environment Variables

All configuration values can be overridden using environment variables. This is particularly useful for:

- Docker/Kubernetes deployments
- CI/CD pipelines
- Different environments (dev, staging, production)
- Keeping secrets out of version control

### Dual Format Support

The board-service supports two environment variable formats for maximum compatibility:

**Original Format (wealist-project compatible):**
- `DATABASE_URL`: Full PostgreSQL connection string
- `SECRET_KEY`: JWT signing key
- `USER_SERVICE_URL`: User service endpoint
- `ENV`: Environment mode (dev/prod)
- `CORS_ORIGINS`: Allowed CORS origins

**Current Format:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: Individual database settings
- `JWT_SECRET`: JWT signing key
- `USER_API_BASE_URL`: User service endpoint
- `SERVER_MODE`: Server mode (debug/release)
- `CORS_ALLOWED_ORIGINS`: Allowed CORS origins

Both formats are fully supported. When both formats are provided, the original format takes precedence to ensure compatibility with existing wealist-project deployments.

### Priority

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values in code
2. `config.yaml` file
3. Environment variables (original format takes precedence over current format)

## Configuration Reference

### Server Configuration

Controls HTTP server behavior.

| YAML Path | Environment Variable | Alias | Type | Default | Description |
|-----------|---------------------|-------|------|---------|-------------|
| `server.port` | `SERVER_PORT` | - | string | `"8000"` | Port to listen on |
| `server.mode` | `SERVER_MODE` | `ENV` | string | `"debug"` | Server mode: `debug` or `release` |
| `server.read_timeout` | `SERVER_READ_TIMEOUT` | - | duration | `10s` | HTTP read timeout |
| `server.write_timeout` | `SERVER_WRITE_TIMEOUT` | - | duration | `10s` | HTTP write timeout |
| `server.shutdown_timeout` | `SERVER_SHUTDOWN_TIMEOUT` | - | duration | `30s` | Graceful shutdown timeout |

**Server Modes:**
- `debug`: Enables detailed logging, stack traces in errors, and development features
- `release`: Production-optimized mode with minimal logging and no debug information

**ENV Alias Mapping:**
- `ENV=dev` maps to `SERVER_MODE=debug`
- `ENV=prod` maps to `SERVER_MODE=release`

### Database Configuration

PostgreSQL database connection settings.

#### DATABASE_URL Format (Original Format)

The board-service supports PostgreSQL connection strings via the `DATABASE_URL` environment variable:

**Format:**
```
postgresql://user:password@host:port/dbname?sslmode=disable
```

**Example:**
```bash
DATABASE_URL=postgresql://board_service:board_pass12345@postgres:5432/wealist_board_db?sslmode=disable
```

When `DATABASE_URL` is provided, it will be parsed and used for the database connection. Individual `DB_*` variables can still override specific components if provided.

#### Individual Variables (Current Format)

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `database.host` | `DB_HOST` | string | `"localhost"` | PostgreSQL host |
| `database.port` | `DB_PORT` | string | `"5432"` | PostgreSQL port |
| `database.user` | `DB_USER` | string | `"postgres"` | Database user |
| `database.password` | `DB_PASSWORD` | string | - | Database password |
| `database.dbname` | `DB_NAME` | string | `"project_board"` | Database name |
| `database.max_open_conns` | `DB_MAX_OPEN_CONNS` | int | `25` | Maximum open connections |
| `database.max_idle_conns` | `DB_MAX_IDLE_CONNS` | int | `5` | Maximum idle connections |
| `database.conn_max_lifetime` | `DB_CONN_MAX_LIFETIME` | duration | `5m` | Connection max lifetime |

**Connection Pool Guidelines:**
- `max_open_conns`: Set based on your database server capacity (typically 25-100)
- `max_idle_conns`: Should be less than `max_open_conns` (typically 5-10)
- `conn_max_lifetime`: Prevents stale connections (typically 5m-30m)

**Precedence Rules:**
1. If `DATABASE_URL` is provided, it is parsed first
2. Individual `DB_*` variables override the parsed values if provided
3. This allows flexible configuration in different environments

### Logger Configuration

Structured logging with Zap.

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `logger.level` | `LOG_LEVEL` | string | `"info"` | Log level |
| `logger.output_path` | `LOG_OUTPUT_PATH` | string | `"stdout"` | Output destination |

**Log Levels:**
- `debug`: Most verbose, includes all logs (development)
- `info`: General information about application flow (production)
- `warn`: Warning messages
- `error`: Error messages only

**Output Paths:**
- `stdout`: Standard output (default, good for Docker)
- `stderr`: Standard error
- `/path/to/file.log`: File path for persistent logging

### JWT Configuration

JSON Web Token settings for authentication.

| YAML Path | Environment Variable | Alias | Type | Default | Description |
|-----------|---------------------|-------|------|---------|-------------|
| `jwt.secret` | `JWT_SECRET` | `SECRET_KEY` | string | - | Secret key for signing tokens |
| `jwt.expire_time` | `JWT_EXPIRE_TIME` | - | duration | `24h` | Token expiration time |

**Alias Support:**
- `SECRET_KEY` is supported as an alias for `JWT_SECRET` (original format)
- Both variable names work identically
- `SECRET_KEY` takes precedence if both are provided

**Security Notes:**
- **NEVER** commit JWT secrets to version control
- Use a strong, random secret (minimum 32 characters)
- Generate with: `openssl rand -base64 32`
- Rotate secrets periodically in production

### User Service Configuration

External user service connection settings.

| YAML Path | Environment Variable | Alias | Type | Default | Description |
|-----------|---------------------|-------|------|---------|-------------|
| `user_api.base_url` | `USER_API_BASE_URL` | `USER_SERVICE_URL` | string | - | User service endpoint URL |

**Alias Support:**
- `USER_SERVICE_URL` is supported as an alias for `USER_API_BASE_URL` (original format)
- Both variable names work identically
- `USER_SERVICE_URL` takes precedence if both are provided

### CORS Configuration

Cross-Origin Resource Sharing settings.

| YAML Path | Environment Variable | Alias | Type | Default | Description |
|-----------|---------------------|-------|------|---------|-------------|
| `cors.allowed_origins` | `CORS_ALLOWED_ORIGINS` | `CORS_ORIGINS` | string | - | Comma-separated list of allowed origins |

**Alias Support:**
- `CORS_ORIGINS` is supported as an alias for `CORS_ALLOWED_ORIGINS` (original format)
- Both variable names work identically
- `CORS_ORIGINS` takes precedence if both are provided

**Example:**
```bash
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

## Environment Variable Alias Mappings

The board-service supports dual environment variable formats for compatibility with the original wealist-project environment. This section documents all alias mappings and precedence rules.

### Complete Alias Reference

| Original Format (Priority) | Current Format | Configuration Path | Description |
|---------------------------|----------------|-------------------|-------------|
| `DATABASE_URL` | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | `database.*` | PostgreSQL connection |
| `SECRET_KEY` | `JWT_SECRET` | `jwt.secret` | JWT signing key |
| `USER_SERVICE_URL` | `USER_API_BASE_URL` | `user_api.base_url` | User service endpoint |
| `ENV` | `SERVER_MODE` | `server.mode` | Server mode (dev/prod vs debug/release) |
| `CORS_ORIGINS` | `CORS_ALLOWED_ORIGINS` | `cors.allowed_origins` | CORS allowed origins |

### Precedence Rules

When both formats are provided, the **original format takes precedence**:

1. **DATABASE_URL** is checked first
   - If found, it is parsed to extract connection components
   - Individual `DB_*` variables can still override specific components
   - Example: `DATABASE_URL` sets host, but `DB_PORT` can override the port

2. **SECRET_KEY** takes precedence over `JWT_SECRET`
   - If `SECRET_KEY` is set, it is used
   - If not, `JWT_SECRET` is checked
   - Only one needs to be provided

3. **USER_SERVICE_URL** takes precedence over `USER_API_BASE_URL`
   - If `USER_SERVICE_URL` is set, it is used
   - If not, `USER_API_BASE_URL` is checked

4. **ENV** takes precedence over `SERVER_MODE`
   - `ENV=dev` maps to `SERVER_MODE=debug`
   - `ENV=prod` maps to `SERVER_MODE=release`
   - If `ENV` is not set, `SERVER_MODE` is used directly

5. **CORS_ORIGINS** takes precedence over `CORS_ALLOWED_ORIGINS`
   - If `CORS_ORIGINS` is set, it is used
   - If not, `CORS_ALLOWED_ORIGINS` is checked

### DATABASE_URL Format Details

The `DATABASE_URL` environment variable uses the standard PostgreSQL connection string format:

**Format:**
```
postgresql://[user]:[password]@[host]:[port]/[database]?[parameters]
```

**Components:**
- `user`: Database username
- `password`: Database password (URL-encoded if contains special characters)
- `host`: Database hostname or IP address
- `port`: Database port (typically 5432)
- `database`: Database name
- `parameters`: Optional query parameters (e.g., `sslmode=disable`)

**Examples:**

Basic connection:
```bash
DATABASE_URL=postgresql://postgres:password@localhost:5432/project_board?sslmode=disable
```

Docker environment:
```bash
DATABASE_URL=postgresql://board_service:board_pass12345@postgres:5432/wealist_board_db?sslmode=disable
```

With special characters in password (URL-encoded):
```bash
DATABASE_URL=postgresql://user:p%40ssw%23rd@localhost:5432/dbname?sslmode=disable
```

**Parsing Behavior:**
- The URL is parsed to extract individual components
- Extracted values populate `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Individual `DB_*` variables can override parsed values if provided after
- Invalid URLs will result in a clear error message with the expected format

### Migration Guide for Existing Deployments

If you're migrating from an existing deployment using the current format, you have two options:

#### Option 1: Keep Current Format (No Changes Required)

Your existing configuration will continue to work without any changes:

```bash
# Existing .env file - still works!
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=project_board
JWT_SECRET=your-secret-key
USER_API_BASE_URL=http://user-service:8080
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

#### Option 2: Migrate to Original Format

To align with the wealist-project environment, update your `.env` file:

**Before (Current Format):**
```bash
SERVER_PORT=8080
SERVER_MODE=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=project_board
JWT_SECRET=your-secret-key
USER_API_BASE_URL=http://user-service:8080
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

**After (Original Format):**
```bash
SERVER_PORT=8000
ENV=dev
DATABASE_URL=postgresql://postgres:password@localhost:5432/project_board?sslmode=disable
SECRET_KEY=your-secret-key
USER_SERVICE_URL=http://user-service:8080
CORS_ORIGINS=http://localhost:3000
```

**Key Changes:**
1. Port changed from `8080` to `8000`
2. `SERVER_MODE=debug` → `ENV=dev`
3. Individual `DB_*` variables → `DATABASE_URL`
4. `JWT_SECRET` → `SECRET_KEY`
5. `USER_API_BASE_URL` → `USER_SERVICE_URL`
6. `CORS_ALLOWED_ORIGINS` → `CORS_ORIGINS`

#### Option 3: Hybrid Approach

You can mix both formats. Original format takes precedence:

```bash
# Use original format for most settings
ENV=dev
DATABASE_URL=postgresql://postgres:password@postgres:5432/project_board?sslmode=disable
SECRET_KEY=your-secret-key

# Override specific database settings with current format
DB_HOST=localhost  # Overrides host from DATABASE_URL

# Use current format for other settings
USER_API_BASE_URL=http://user-service:8080
```

### Validation and Error Handling

The configuration loader validates all settings on startup:

**DATABASE_URL Parsing Errors:**
```
Failed to parse DATABASE_URL: invalid format
Expected: postgresql://user:password@host:port/dbname?sslmode=disable
Falling back to individual DB_* environment variables
```

**Missing Required Variables:**
```
Configuration validation failed: jwt secret is required
Please set either SECRET_KEY or JWT_SECRET environment variable
```

**Helpful Error Messages:**
- All errors indicate which variable is missing or invalid
- Errors suggest both possible variable names (original and current format)
- Clear format examples are provided for DATABASE_URL

## Examples

### Development Environment (Original Format)

**.env:**
```bash
# Board Service Configuration
ENV=dev
SERVER_PORT=8000
LOG_LEVEL=debug

# Database Configuration (using DATABASE_URL)
DATABASE_URL=postgresql://postgres:password@localhost:5432/project_board_dev?sslmode=disable

# JWT Configuration
SECRET_KEY=dev-secret-key-not-for-production

# User Service Configuration
USER_SERVICE_URL=http://localhost:8080

# CORS Configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Development Environment (Current Format)

**config.yaml:**
```yaml
server:
  port: "8000"
  mode: "debug"
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 30s

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "project_board_dev"
  max_open_conns: 10
  max_idle_conns: 2
  conn_max_lifetime: 5m

logger:
  level: "debug"
  output_path: "stdout"

jwt:
  secret: "dev-secret-key-not-for-production"
  expire_time: 24h
```

### Production Environment (Original Format)

**.env:**
```bash
# Board Service Configuration
ENV=prod
SERVER_PORT=8000
LOG_LEVEL=info

# Database Configuration
DATABASE_URL=postgresql://app_user:strong-db-password@postgres.production.internal:5432/project_board?sslmode=require

# JWT Configuration
SECRET_KEY=<strong-random-secret-min-32-chars>

# User Service Configuration
USER_SERVICE_URL=http://user-service:8080

# CORS Configuration
CORS_ORIGINS=https://app.example.com,https://www.example.com
```

### Production Environment (Current Format)

**config.yaml:**
```yaml
server:
  port: "8000"
  mode: "release"
  read_timeout: 15s
  write_timeout: 15s
  shutdown_timeout: 30s

database:
  host: "postgres.production.internal"
  port: "5432"
  user: "app_user"
  dbname: "project_board"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 10m

logger:
  level: "info"
  output_path: "/var/log/project-board/app.log"

jwt:
  expire_time: 8h
```

**Environment Variables (for secrets):**
```bash
DB_PASSWORD=<strong-database-password>
JWT_SECRET=<strong-random-secret>
```

### Docker Deployment (Original Format)

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  board-service:
    image: project-board-api:latest
    container_name: board-service
    environment:
      - SERVER_PORT=8000
      - ENV=prod
      - DATABASE_URL=postgresql://postgres:${DB_PASSWORD}@postgres:5432/project_board?sslmode=disable
      - SECRET_KEY=${JWT_SECRET}
      - USER_SERVICE_URL=http://user-service:8080
      - CORS_ORIGINS=http://localhost:3000
      - LOG_LEVEL=info
    ports:
      - "8000:8000"
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 3s
      retries: 3
  
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=project_board
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**.env (for docker-compose):**
```bash
DB_PASSWORD=secure-password-here
JWT_SECRET=secure-jwt-secret-here
```

### Docker Deployment (Current Format)

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  api:
    image: project-board-api:latest
    environment:
      - SERVER_PORT=8000
      - SERVER_MODE=release
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=project_board
      - LOG_LEVEL=info
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8000:8000"
    depends_on:
      - postgres
  
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=project_board
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**.env (for docker-compose):**
```bash
DB_PASSWORD=secure-password-here
JWT_SECRET=secure-jwt-secret-here
```

### Kubernetes Deployment

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: project-board-config
data:
  SERVER_PORT: "8080"
  SERVER_MODE: "release"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_USER: "postgres"
  DB_NAME: "project_board"
  LOG_LEVEL: "info"
  LOG_OUTPUT_PATH: "stdout"
```

**Secret:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: project-board-secrets
type: Opaque
stringData:
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-secret>
```

**Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: project-board-api
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: project-board-api:latest
        envFrom:
        - configMapRef:
            name: project-board-config
        - secretRef:
            name: project-board-secrets
        ports:
        - containerPort: 8080
```

## Duration Format

Duration values use Go's duration format:

- `s` - seconds (e.g., `30s`)
- `m` - minutes (e.g., `5m`)
- `h` - hours (e.g., `24h`)
- `d` - days (e.g., `7d`) - Note: Use hours for days (168h = 7d)

**Examples:**
- `10s` - 10 seconds
- `5m` - 5 minutes
- `24h` - 24 hours
- `168h` - 7 days

## Validation

The application validates configuration on startup. Required fields:

- `server.port`
- `database.host`
- `database.port`
- `database.user`
- `database.dbname`
- `jwt.secret`

If validation fails, the application will exit with an error message indicating which field is missing or invalid.

## Best Practices

### Security

1. **Never commit secrets** to version control
2. Use environment variables for sensitive data
3. Rotate JWT secrets regularly
4. Use strong, random passwords for database
5. Restrict database user permissions

### Performance

1. Tune connection pool based on load
2. Set appropriate timeouts
3. Use `release` mode in production
4. Monitor connection pool metrics

### Logging

1. Use `info` level in production
2. Use `debug` level only for troubleshooting
3. Rotate log files to prevent disk space issues
4. Use structured logging for better analysis

### Deployment

1. Use separate configs for each environment
2. Override secrets with environment variables
3. Test configuration changes in staging first
4. Document any custom configuration

## Troubleshooting

### Application won't start

**Error: "database host is required"**
- Ensure `DB_HOST` is set or `database.host` is in config.yaml

**Error: "jwt secret is required"**
- Set `JWT_SECRET` environment variable or `jwt.secret` in config.yaml

### Database connection issues

**Error: "connection refused"**
- Check `DB_HOST` and `DB_PORT` are correct
- Ensure PostgreSQL is running
- Verify network connectivity

**Error: "authentication failed"**
- Verify `DB_USER` and `DB_PASSWORD` are correct
- Check PostgreSQL user permissions

### Performance issues

**Too many database connections**
- Reduce `max_open_conns`
- Increase `conn_max_lifetime`

**Slow response times**
- Increase `read_timeout` and `write_timeout`
- Check database query performance
- Review connection pool settings

## Support

For additional help:
- Check application logs for detailed error messages
- Review the [README.md](../README.md) for setup instructions
- Consult the [API documentation](./API.md) for endpoint details
