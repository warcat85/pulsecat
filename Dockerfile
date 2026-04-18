# Multi-stage Dockerfile for PulseCat
# Build stage
FROM golang:1.24.3-alpine AS builder

# Use Yandex mirror for Alpine packages
RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|http://mirror.yandex.ru/mirrors/alpine|g' /etc/apk/repositories

# Install build dependencies
RUN apk add --no-cache git tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION:-dev} \
    -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
    -X main.CommitHash=${COMMIT_HASH:-unknown}" \
    -o /app/pulsecat ./cmd/pulsecat

# Runtime stage - using scratch (minimal) with just the binary
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone info from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder
COPY --from=builder /app/pulsecat /pulsecat

# Expose gRPC port
EXPOSE 25225

# Health check disabled in scratch (no shell)
# For production, use external health checks or a different base image

# Set entrypoint
ENTRYPOINT ["/pulsecat"]

# Default command (can be overridden)
CMD ["--config", "configs/config.yaml"]

# Notes for running this container:
# 1. For full system monitoring capabilities, run with:
#    --privileged \
#    -v /proc:/host/proc:ro \
#    -v /sys:/host/sys:ro \
#    -v /var/run/docker.sock:/var/run/docker.sock:ro (optional)
#    --network host (for network monitoring)
#
# 2. Example docker run command:
#    docker run -d \
#      --name pulsecat \
#      --privileged \
#      -v /proc:/host/proc:ro \
#      -v /sys:/host/sys:ro \
#      -p 25225:25225 \
#      pulsecat:latest
#
# 3. Since we're using scratch, the container has no shell or utilities.
#    This is more secure but may require host mounts for system monitoring.