# Multi-stage build for smaller final image
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy static assets and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/gopher_six.json .

# Change ownership to appuser
RUN chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8000/api/health || exit 1

# Set environment variables
ENV PORT=8000
ENV HOST=0.0.0.0

# Run the application
CMD ["./main"]
