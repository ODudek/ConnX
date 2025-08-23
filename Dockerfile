FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o connx ./cmd/proxy

FROM alpine:latest

# Install ca-certificates for HTTPS backends
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/connx .

# Copy default configuration
COPY configs/config.yaml /app/configs/config.yaml

# Create non-root user
RUN addgroup -g 1001 -S connx && \
    adduser -u 1001 -S connx -G connx

# Change ownership of app directory
RUN chown -R connx:connx /app

# Switch to non-root user
USER connx

EXPOSE 8080

ENTRYPOINT [ "./connx" ]
