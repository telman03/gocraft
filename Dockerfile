# Build stage
FROM golang:1.23.12-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application and place it in bin directory as expected by deployment
RUN mkdir -p bin && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ai-backend-generator ./cmd/gocraft

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S gocraft && \
    adduser -S gocraft -u 1001 -G gocraft

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/ai-backend-generator ./bin/ai-backend-generator

# Copy templates and static files
COPY --from=builder /app/internal/templates ./internal/templates

# Create output directory
RUN mkdir -p output && chown -R gocraft:gocraft /app

# Switch to non-root user
USER gocraft

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1

# Run the application
CMD ["bin/ai-backend-generator"]