# =========================
# 1. BUILD STAGE
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build api server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o api ./cmd/api

# =========================
# 2. RUNTIME STAGE
# =========================
FROM alpine:3.20

WORKDIR /app

# Add CA certs for HTTPS/JWT/OAuth calls
RUN apk add --no-cache ca-certificates curl

# Copy binary only
COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]
