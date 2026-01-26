# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY server/go.mod server/go.sum ./server/
WORKDIR /app/server
RUN go mod download

# Copy source code
WORKDIR /app
COPY server/ ./server/
COPY client/ ./client/

# Build the binary
WORKDIR /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/zetl .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 zetl && \
    adduser -u 1000 -G zetl -s /bin/sh -D zetl

# Copy binary to server/ subdirectory (Go code uses relative paths like ../client/)
COPY --from=builder /app/zetl ./server/zetl
COPY --from=builder /app/client ./client

# Set ownership
RUN chown -R zetl:zetl /app

USER zetl

WORKDIR /app/server

EXPOSE 8080

CMD ["./zetl"]
