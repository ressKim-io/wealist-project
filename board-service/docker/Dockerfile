# Builder stage
FROM golang:1.25-bookworm AS builder 

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# 💡 [필수] 종속성 정리: go.mod에 명시된 버전에 맞춰 모듈을 정리하고, 누락된 의존성을 확보합니다.
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
# 🚨 [최종 수정] -mod=readonly 대신 -mod=mod를 사용하여 빌드 시 종속성 그래프를 다시 확인하도록 강제합니다.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -ldflags="-w -s" -o board-api ./cmd/api

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
# 런타임은 계속 가벼운 Alpine을 사용할 수 있습니다.
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /app/board-api .

# Copy configuration files
COPY --from=builder --chown=appuser:appuser /app/configs ./configs

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# Run the application
CMD ["./board-api"]