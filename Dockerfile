# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download 2>&1

# Copy entire source
COPY . .

# Verify go.mod and go.sum
RUN go mod verify 2>&1

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -installsuffix cgo -o auth-service ./cmd/api && \
    echo "Build complete, checking for binary..." && \
    ls -lah auth-service

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/auth-service .

# Copy migrations
COPY --from=builder /app/database/migrations ./database/migrations

# Copy .env if it exists (optional in docker)
COPY --from=builder /app/.env .

EXPOSE 8000

CMD ["./auth-service"]
