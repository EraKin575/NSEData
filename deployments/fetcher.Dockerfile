# Stage 1: Build
FROM golang:1.24.2-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git tzdata

# Dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 go build -o fetcher ./cmd/fetcher/main.go

# Stage 2: Run
FROM alpine:latest
WORKDIR /app

# Copy binary
COPY --from=builder /app/fetcher .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Use non-root user
RUN adduser -D appuser
USER appuser

CMD ["./fetcher"]
