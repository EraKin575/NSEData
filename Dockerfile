# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools and timezone data
RUN apk add --no-cache git tzdata

# Copy go mod and sum first for better caching
COPY go.mod ./
RUN go mod tidy

# Copy rest of the project
COPY . .

# Build static binary
RUN CGO_ENABLED=0 go build -o main ./cmd/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Copy binary
COPY --from=builder /app/main .

# Optional: copy timezone data if your app uses it
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Use non-root user
RUN adduser -D appuser
USER appuser

CMD ["./main"]
