# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

ConnX is a high-performance HTTP reverse proxy and load balancer written in Go that uses the gnet framework for efficient network operations. The project provides round-robin load balancing, health checking, and real-time backend monitoring.

## Common Commands

### Development Commands
- `make build` - Build the proxy binary to `bin/proxy`
- `make test` - Run all tests using `go test ./...`
- `make run` - Run the proxy directly with `go run cmd/proxy/main.go`
- `make clean` - Remove the bin/ directory
- `go run cmd/proxy/main.go -config=path/to/config.yaml` - Run with custom config

### Testing
- `go test ./...` - Run all tests
- `go test -v ./internal/proxy` - Run tests for specific package with verbose output
- `go test -race ./...` - Run tests with race detection

### Code Quality
- Check for golangci-lint configuration for linting standards
- The project uses standard Go formatting with `go fmt`

## Architecture

### Core Components

The application follows a clean architecture pattern with these key packages:

1. **cmd/proxy/main.go** - Entry point that loads configuration and starts the server
2. **internal/server** - Main gnet-based HTTP server that handles incoming connections and `/metrics` endpoint
3. **internal/proxy** - Backend management and server pool operations
4. **internal/health** - Health checking system for backend servers with metrics integration
5. **internal/loadbalancer** - Load balancing algorithms (round-robin)
6. **internal/config** - Configuration loading and validation
7. **internal/metrics** - Prometheus-style metrics collection and exposition
8. **pkg/** - Reusable utilities (HTTP parsing, logging)

### Key Architecture Patterns

- **Event-driven networking**: Uses gnet.BuiltinEventEngine for high-performance connection handling
- **Server pool pattern**: Centralized management of backend servers with health status tracking
- **Configuration-driven**: YAML-based configuration with sensible defaults
- **Graceful error handling**: Proper error propagation and backend failure handling

### Request Flow

1. gnet server receives connection in `server.OnTraffic()`
2. HTTP request is parsed from raw bytes
3. Check if request is for `/metrics` endpoint and handle accordingly
4. Load balancer selects next available backend via `pool.GetNextPeer()`
5. Request is forwarded to backend with proper header copying
6. Response is dumped and written back to client
7. Metrics are recorded (request counts, response times, errors)
8. Errors increment backend error counters and metrics

### Configuration Structure

Configuration is loaded from YAML with the following structure:
- `server.port` / `server.host` - Server binding configuration
- `backends[]` - List of backend URLs
- `healthCheck.interval` / `healthCheck.timeout` - Health check parameters

Default values are applied in `config.Validate()` method:
- Port: 8080
- Host: "0.0.0.0" 
- Health check interval: 30 seconds
- Health check timeout: 5 seconds

### Health Checking System

- Runs in background goroutine with configurable intervals
- Performs HTTP GET requests to backend URLs
- Updates backend status in server pool and metrics
- Failed backends are marked unavailable and excluded from load balancing

### Metrics System

- **Endpoint**: `/metrics` - Prometheus-style metrics exposition
- **Metrics collected**:
  - `connx_requests_total` - Total HTTP requests by method and status
  - `connx_errors_total` - Total error count
  - `connx_response_time_seconds` - Response time statistics (avg, p95, p99)
  - `connx_backend_up` - Backend health status (1=up, 0=down)
  - `connx_active_connections` - Current active connections
  - `connx_uptime_seconds` - Server uptime
  - `connx_requests_per_second` - Current RPS
- **Integration**: Metrics are collected in real-time during request processing
- **Memory management**: Response time history limited to last 1000 requests

## Dependencies

- **gnet v2.7.2** - High-performance networking framework
- **yaml.v3** - YAML configuration parsing
- **Standard library** - HTTP utilities, logging, time operations

## Development Notes

- The server uses multicore processing via `gnet.WithMulticore(true)`
- Port reusing is enabled for better performance
- Error counting is implemented per backend for monitoring
- Raw HTTP request/response handling for maximum performance
- Comments are now in English throughout the codebase

## Potential Enhancements

### Performance & Monitoring
- ✅ **Metrics endpoint** (`/metrics`) - Prometheus-style metrics for requests/sec, response times, error rates
- **Connection pooling** - Reuse HTTP connections to backends instead of creating new ones per request
- **Request/response compression** - Add gzip support for bandwidth optimization
- **Circuit breaker pattern** - Temporarily stop sending requests to consistently failing backends

### Load Balancing & Routing
- **Multiple load balancing algorithms** - Weighted round-robin, least connections, IP hash
- **Path-based routing** - Route requests to different backend pools based on URL paths
- **Header-based routing** - Route based on custom headers (e.g., API version, tenant ID)
- **Sticky sessions** - Session affinity using cookies or IP-based routing

### Security & Reliability
- **Rate limiting** - Per-IP or per-route request throttling
- **TLS/HTTPS support** - SSL termination and backend HTTPS connections
- **Authentication middleware** - JWT validation, API key checking
- **Request/response transformation** - Header manipulation, request/response body modification

### Observability & Operations
- **Structured logging** - JSON logs with correlation IDs, request tracing
- **Health check dashboard** - Web UI showing backend status and statistics
- **Graceful shutdown** - Proper connection draining on SIGTERM
- **Hot configuration reload** - Update backends/config without restart

### Advanced Features
- **WebSocket proxying** - Support for WebSocket connections
- **HTTP/2 support** - Modern protocol support for better performance
- **Dynamic backend discovery** - Integration with service discovery (Consul, etcd)
- **Request caching** - Cache responses for improved performance