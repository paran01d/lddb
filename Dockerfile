# Build stage  
FROM golang:1.23-alpine AS builder

# Install required packages for compilation
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Tidy up the module dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests to lddb.com
RUN apk --no-cache add ca-certificates sqlite

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy web assets
COPY --from=builder /app/web ./web

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

# Declare volume for data persistence
VOLUME ["/app/data"]

# Switch to non-root user
USER appuser

# Expose port (will be dynamically determined by the app)
EXPOSE 8080-8099

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/collection || exit 1

# Set environment variables
ENV GIN_MODE=release

# Run the application
CMD ["./main"]