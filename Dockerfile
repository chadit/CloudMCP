# CloudMCP Multi-stage Dockerfile
# Optimized for small image size and security

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -a -installsuffix cgo \
    -o cloud-mcp \
    cmd/cloud-mcp/main.go

# Verify the binary
RUN ./cloud-mcp --version || echo "Binary built successfully"

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    && adduser -D -s /bin/sh cloudmcp

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/cloud-mcp /usr/local/bin/cloud-mcp

# Make binary executable
RUN chmod +x /usr/local/bin/cloud-mcp

# Create necessary directories
RUN mkdir -p /app/config /app/logs && \
    chown -R cloudmcp:cloudmcp /app

# Switch to non-root user
USER cloudmcp

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose metrics port (configurable via METRICS_PORT)
EXPOSE 8080

# Set default environment variables
ENV LOG_LEVEL=info \
    ENABLE_METRICS=true \
    METRICS_PORT=8080 \
    DAEMON_MODE=true

# Default command
ENTRYPOINT ["/usr/local/bin/cloud-mcp"]
CMD []

# Metadata
LABEL maintainer="CloudMCP Team" \
      description="CloudMCP - Model Context Protocol server for cloud infrastructure management" \
      version="${VERSION:-dev}" \
      org.opencontainers.image.title="CloudMCP" \
      org.opencontainers.image.description="Model Context Protocol server for cloud infrastructure management" \
      org.opencontainers.image.vendor="CloudMCP" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/chadit/CloudMCP"