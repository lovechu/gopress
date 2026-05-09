# Multi-stage Dockerfile for GoPress
# Build stage
FROM golang:1.21-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gopress ./cmd/server

# Production stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata bash

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN adduser -D -g '' gopress

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/gopress .
COPY --from=builder /app/docker-entrypoint.sh ./
COPY --from=builder /app/themes ./themes
COPY --from=builder /app/migrations ./migrations

# Make entrypoint script executable
RUN chmod +x docker-entrypoint.sh

# Create storage directory
RUN mkdir -p ./storage/uploads && chown -R gopress:gopress ./storage

# Switch to non-root user
USER gopress

# Expose the application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Use entrypoint script
ENTRYPOINT ["./docker-entrypoint.sh"]

# Run the application
CMD ["./gopress"]
