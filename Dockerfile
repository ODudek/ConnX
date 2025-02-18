FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o connx ./cmd/proxy

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/connx .
COPY configs/config.yaml /app/configs/config.yaml
EXPOSE 8080
CMD ["./connx"]
