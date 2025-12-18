# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build all binaries with optimizations
# -ldflags="-s -w" strips debug info for smaller binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=2.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -a -installsuffix cgo -o server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -a -installsuffix cgo -o scraper ./cmd/scraper

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -a -installsuffix cgo -o scheduler ./cmd/scheduler

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -a -installsuffix cgo -o geocode ./cmd/geocode

# Final stage - minimal image
FROM scratch

# Copy CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binaries
COPY --from=builder /app/server /server
COPY --from=builder /app/scraper /scraper
COPY --from=builder /app/scheduler /scheduler
COPY --from=builder /app/geocode /geocode

# Copy web files
COPY --from=builder /app/web /web

# Expose port
EXPOSE 3000

# Note: Health check removed from Dockerfile as 'scratch' image has no curl/wget
# Use Kubernetes liveness/readiness probes or external monitoring instead
# For Docker health check, use docker-compose with a sidecar or external probe

# Run as non-root user for security
USER 1000:1000

# Default command
ENTRYPOINT ["/server"]

