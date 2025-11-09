# Stage 1: Builder
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Sync go.mod and go.sum first (for Go version compatibility)
RUN go mod tidy

# Download dependencies (after tidy to ensure consistency)
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o board-api ./cmd/api

# golang-migrate CLI 다운로드
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Stage 2: Runtime
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl postgresql-client

# Create non-root user
RUN adduser -D -u 1001 appuser

# Copy binary from builder
COPY --from=builder /build/board-api .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy scripts and migrations
COPY --from=builder /build/scripts/init.sh /scripts/
COPY --from=builder /build/scripts/wait-for-it.sh /scripts/
COPY --from=builder /build/migrations /app/migrations

# Make scripts executable
RUN chmod +x /scripts/init.sh /scripts/wait-for-it.sh

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Use init script as entrypoint
ENTRYPOINT ["/scripts/init.sh"]
