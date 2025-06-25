# Multi-stage build for the container installer
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the container installer binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o container-installer ./container/

# Final container image
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add \
    curl \
    wget \
    git \
    docker-cli \
    docker-compose \
    bash \
    sudo \
    ca-certificates \
    openssl \
    nodejs \
    npm

# Create app user (non-root)
RUN adduser -D -s /bin/bash app

# Set working directory
WORKDIR /app

# Copy installer binary
COPY --from=builder /app/container-installer /usr/local/bin/container-installer

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/configs /app/templates \
    && chown -R app:app /app

# Copy templates and configs
COPY container/templates/ /app/templates/

# Set executable permissions
RUN chmod +x /usr/local/bin/container-installer

# Switch to app user for security
USER app

# Set default command
ENTRYPOINT ["/usr/local/bin/container-installer"]
CMD ["install"]
