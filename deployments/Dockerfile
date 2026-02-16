# Multi-stage build for Poly Oracle

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o poly-oracle ./cmd/poly-oracle

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/poly-oracle .

# Create data directory
RUN mkdir -p /app/data

# Copy example config
COPY configs/config.yaml.example /app/configs/config.yaml.example

# Expose no ports (not needed for this service)

# Run the binary
ENTRYPOINT ["./poly-oracle"]
CMD ["--config", "/app/configs/config.yaml"]
