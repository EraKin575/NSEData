# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools (if needed for CGO dependencies)
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod ./
RUN go mod tidy

# Copy rest of the project
COPY . .

# Build binary
RUN go build -o main ./cmd/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Copy only the binary from builder stage
COPY --from=builder /app/main .

# Use a non-root user for safety (optional)
RUN adduser -D appuser
USER appuser

CMD ["./main"]
