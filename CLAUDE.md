# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

ConnX is a high-performance HTTP reverse proxy and load balancer written in Go that uses the gnet framework for efficient network operations. The project provides multiple load balancing algorithms, circuit breaker pattern, rate limiting, health checking, graceful shutdown, and real-time backend monitoring.

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

1. **cmd/proxy/main.go** - Entry point that loads configuration, starts the server, and handles graceful shutdown
2. **internal/server** - Main gnet-based HTTP server that handles incoming connections and `/metrics` endpoint
3. **internal/proxy** - Backend management and server pool operations with circuit breaker integration
4. **internal/health** - Health checking system for backend servers with metrics integration
5. **internal/loadbalancer** - Load balancing algorithms (round-robin, weighted round-robin, least connections)
6. **internal/circuitbreaker** - Circuit breaker pattern implementation for automatic failure handling
7. **internal/ratelimit** - Token bucket rate limiting (global and per-IP)
8. **internal/config** - Configuration loading and validation with backward compatibility
9. **internal/metrics** - Prometheus-style metrics collection and exposition
10. **pkg/** - Reusable utilities (HTTP parsing, logging)

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
- `server.port` / `server.host` / `server.shutdownTimeout` - Server configuration
- `backends[]` - List of backend configurations with URL and weight
- `loadBalancing.algorithm` - Load balancing algorithm selection
- `healthCheck.interval` / `healthCheck.timeout` - Health check parameters
- `circuitBreaker.*` - Circuit breaker configuration
- `rateLimit.*` - Rate limiting configuration

Default values are applied in `config.Validate()` method:
- Port: 8080
- Host: "0.0.0.0"
- Shutdown timeout: 30 seconds
- Health check interval: 30 seconds
- Health check timeout: 5 seconds
- Load balancing: round-robin
- Circuit breaker: enabled with 5 max failures, 60s timeout
- Rate limit: disabled

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

## Implemented Features

### Performance & Monitoring
- ✅ **Metrics endpoint** (`/metrics`) - Prometheus-style metrics for requests/sec, response times, error rates
- ✅ **Circuit breaker pattern** - Automatically opens circuit after failures, prevents cascade failures
- ✅ **Active connection tracking** - Tracks active connections per backend for least-connections algorithm

### Load Balancing & Routing
- ✅ **Multiple load balancing algorithms** - Round-robin, weighted round-robin, least connections
- ✅ **Weighted backends** - Distribute traffic based on backend capacity/priority

### Security & Reliability
- ✅ **Rate limiting** - Token bucket rate limiting with global and per-IP modes
- ✅ **Circuit breaker** - Automatic failure detection and recovery with half-open state
- ✅ **Graceful shutdown** - Proper connection draining on SIGTERM/SIGINT with configurable timeout

### Configuration & Operations
- ✅ **Backward compatibility** - Legacy config format automatically converted to new format
- ✅ **Comprehensive configuration** - All features configurable via YAML

## Potential Future Enhancements

### Performance & Monitoring
- **Connection pooling** - Reuse HTTP connections to backends instead of creating new ones per request
- **Request/response compression** - Add gzip support for bandwidth optimization

### Load Balancing & Routing
- **Path-based routing** - Route requests to different backend pools based on URL paths
- **Header-based routing** - Route based on custom headers (e.g., API version, tenant ID)
- **Sticky sessions** - Session affinity using cookies or IP-based routing
- **IP hash algorithm** - Consistent hashing for request distribution

### Security & Reliability
- **TLS/HTTPS support** - SSL termination and backend HTTPS connections
- **Authentication middleware** - JWT validation, API key checking
- **Request/response transformation** - Header manipulation, request/response body modification

### Observability & Operations
- **Structured logging** - JSON logs with correlation IDs, request tracing
- **Health check dashboard** - Web UI showing backend status and statistics
- **Hot configuration reload** - Update backends/config without restart

### Advanced Features
- **WebSocket proxying** - Support for WebSocket connections
- **HTTP/2 support** - Modern protocol support for better performance
- **Dynamic backend discovery** - Integration with service discovery (Consul, etcd)
- **Request caching** - Cache responses for improved performance