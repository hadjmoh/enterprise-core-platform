# Dockerfile for Cloud Ingestion Service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o cloud-ingester \
    ./cmd/cloud-ingester

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 cloud-ingester && \
    adduser -D -u 1000 -G cloud-ingester cloud-ingester

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cloud-ingester .

# Change ownership
RUN chown -R cloud-ingester:cloud-ingester /app

# Switch to non-root user
USER cloud-ingester

# Expose metrics port (if implemented)
EXPOSE 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep cloud-ingester || exit 1

# Run the service
ENTRYPOINT ["./cloud-ingester"]
CMD ["--asset-interval=5m", "--event-interval=1m"]
