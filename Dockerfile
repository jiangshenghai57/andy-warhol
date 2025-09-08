# Use official Go image as build environment
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk add --no-cache git ca-certificates tzdata

# Create a non-root user
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Copy local modules (since you're using replace directives)
COPY config/ ./config/
COPY amortization/ ./amortization/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0: Build static binary
# GOOS=linux: Target Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o andy-warhol .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/andy-warhol", "--health-check"] || exit 1

# Run the binary
CMD ["/app/andy-warhol"]