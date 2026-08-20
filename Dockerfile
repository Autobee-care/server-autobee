# ─── Builder Stage ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o bin/api ./cmd/api

# ─── Runtime Stage ────────────────────────────────────────────────────────────
FROM alpine:3.20

# Security: run as non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/api ./api

# Use non-root user
USER appuser

EXPOSE 8080

ENTRYPOINT ["./api"]
