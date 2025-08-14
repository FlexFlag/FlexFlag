# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the edge server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o edge-server ./cmd/edge-server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S flexflag && \
    adduser -u 1001 -S flexflag -G flexflag

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/edge-server .

# Change ownership to non-root user
RUN chown flexflag:flexflag /app/edge-server

# Switch to non-root user
USER flexflag

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8081/health || exit 1

# Set environment variables with defaults
ENV FLEXFLAG_EDGE_PORT=8081
ENV FLEXFLAG_EDGE_ENVIRONMENT=production
ENV FLEXFLAG_EDGE_MAX_FLAGS=10000
ENV FLEXFLAG_EDGE_MAX_API_KEYS=1000
ENV FLEXFLAG_EDGE_CACHE_TTL=1h
ENV FLEXFLAG_EDGE_CLEANUP_INTERVAL=5m
ENV FLEXFLAG_EDGE_ENABLE_METRICS=true

# Run the edge server
CMD ["./edge-server"]