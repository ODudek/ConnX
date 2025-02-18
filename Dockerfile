FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o proxy cmd/proxy/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/proxy .
COPY configs/config.yaml configs/
EXPOSE 8080
CMD ["./proxy"]
