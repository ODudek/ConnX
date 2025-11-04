# ConnX

ConnX is a high-performance, lightweight HTTP reverse proxy and load balancer written in Go. It leverages the gnet framework for efficient network operations and provides features like health checking and round-robin load balancing.

## Features

### Core Features
- **High-performance HTTP reverse proxy** - Built on gnet for efficient network operations
- **Multiple load balancing algorithms**:
  - Round-robin (default)
  - Weighted round-robin (distribute traffic by backend weight)
  - Least connections (route to backend with fewest active connections)
- **Circuit breaker pattern** - Automatic failure detection and recovery
- **Rate limiting** - Global or per-IP request throttling with token bucket algorithm
- **Graceful shutdown** - Proper connection draining on SIGTERM/SIGINT
- **Active health checking** - Periodic backend health verification
- **Real-time metrics** - Prometheus-compatible `/metrics` endpoint
- **Backward compatible configuration** - Supports both legacy and new config formats

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/ODudek/ConnX.git
cd ConnX

# Start with Docker Compose (includes example backends)
docker-compose up -d

# Test the setup
curl http://localhost:8080
curl http://localhost:8080/metrics
```

### Using Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/ODudek/ConnX/releases):

```bash
# Linux
wget https://github.com/ODudek/ConnX/releases/latest/download/connx-linux-amd64
chmod +x connx-linux-amd64
./connx-linux-amd64 -config=config.yaml

# macOS
wget https://github.com/ODudek/ConnX/releases/latest/download/connx-darwin-amd64
chmod +x connx-darwin-amd64
./connx-darwin-amd64 -config=config.yaml
```

### Building from Source

```bash
git clone https://github.com/ODudek/ConnX.git
cd ConnX
make build
./bin/proxy -config=configs/config.yaml
```

## Configuration

Create a `config.yaml` file with the following structure:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  shutdownTimeout: 30  # seconds for graceful shutdown

backends:
  - url: "http://backend1:8080"
    weight: 1
  - url: "http://backend2:8080"
    weight: 2  # Gets 2x more traffic
  - url: "http://backend3:8080"
    weight: 1

loadBalancing:
  algorithm: "round-robin"  # Options: round-robin, weighted-round-robin, least-connections

healthCheck:
  interval: 30  # seconds
  timeout: 5    # seconds

circuitBreaker:
  enabled: true
  maxFailures: 5      # Open circuit after N failures
  timeout: 60         # Seconds before trying half-open state
  resetTimeout: 30    # Seconds in half-open state
  halfOpenSuccess: 2  # Successes needed to close circuit

rateLimit:
  enabled: false      # Enable/disable rate limiting
  requestsPerSec: 100 # Requests per second allowed
  burst: 200          # Maximum burst size
  perIP: false        # true for per-IP, false for global
```

### Legacy Configuration Support

The old configuration format is still supported for backward compatibility:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

backends:
  - "http://backend1:8080"
  - "http://backend2:8080"

healthCheck:
  interval: 30
  timeout: 5
```

## Monitoring & Metrics

ConnX exposes Prometheus-compatible metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

### Available Metrics

- `connx_requests_total` - Total HTTP requests by method and status
- `connx_errors_total` - Total error count
- `connx_response_time_seconds` - Response time statistics (avg, p95, p99)
- `connx_backend_up` - Backend health status (1=up, 0=down)
- `connx_active_connections` - Current active connections
- `connx_uptime_seconds` - Server uptime
- `connx_requests_per_second` - Current requests per second

## Usage

### Command Line Options

```bash
# Run with default config
./connx

# Run with custom config
./connx -config=/path/to/config.yaml

# Docker
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/configs/config.yaml odudek/connx:latest
```

### Example Configurations

See `configs/examples/` for various configuration examples:

- `basic.yaml` - Minimal setup
- `production.yaml` - Production-ready configuration
- `development.yaml` - Development environment
- `monitoring.yaml` - With Prometheus integration
- `with-rate-limiting.yaml` - Rate limiting enabled
- `weighted-round-robin.yaml` - Weighted load balancing
- `least-connections.yaml` - Least connections algorithm
- `full-features.yaml` - All features enabled

### Default Values

If not specified in the config file, the following default values are used:
- Server port: 8080
- Server host: "0.0.0.0"
- Shutdown timeout: 30 seconds
- Health check interval: 30 seconds
- Health check timeout: 5 seconds
- Load balancing algorithm: round-robin
- Backend weight: 1
- Circuit breaker: enabled (5 failures, 60s timeout)
- Rate limiting: disabled

## CI/CD

This project uses GitHub Actions for Continuous Integration and Deployment:

### CI Pipeline
- Runs tests
- Performs linting using golangci-lint
- Builds the application
- Runs on every push to main and pull requests

### CD Pipeline
- Triggered by tags starting with 'v'
- Creates releases with binaries for multiple platforms
- Builds and pushes Docker images to Docker Hub
- Generates release notes automatically

### Docker
Latest images are available on Docker Hub:
```bash
docker pull odudek/connx:latest
```

## Architecture

ConnX consists of several key components:

- **Proxy Server**: Handles incoming connections and forwards requests to backends
- **Server Pool**: Manages the collection of backend servers with active connection tracking
- **Health Checker**: Monitors backend server health with configurable intervals
- **Load Balancer**: Distributes requests using configurable algorithms (round-robin, weighted, least-connections)
- **Circuit Breaker**: Prevents cascade failures by temporarily disabling unhealthy backends
- **Rate Limiter**: Token bucket-based rate limiting (global or per-IP)

## Performance

ConnX is built with performance in mind:
- Uses gnet for efficient network operations
- Tracks active connections for optimal load distribution
- Minimizes memory allocations
- Supports multicore processing
- Circuit breaker prevents wasted requests to failing backends
- Token bucket rate limiting for efficient request throttling

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Quick Contributing Guide

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`make test`)
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run linting
golangci-lint run

# Build
make build
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Report Issues](https://github.com/ODudek/ConnX/issues)
- 💡 [Feature Requests](https://github.com/ODudek/ConnX/issues/new?template=feature_request.md)
- 💬 [Discussions](https://github.com/ODudek/ConnX/discussions)

## Badges

[![CI](https://github.com/ODudek/ConnX/workflows/CI/badge.svg)](https://github.com/ODudek/ConnX/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ODudek/ConnX)](https://goreportcard.com/report/github.com/ODudek/ConnX)
[![Docker Pulls](https://img.shields.io/docker/pulls/odudek/connx)](https://hub.docker.com/r/odudek/connx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://golang.org/doc/install)

## Acknowledgments

- [gnet](https://github.com/panjf2000/gnet) - High-performance, lightweight, non-blocking, event-driven networking framework
- All [contributors](https://github.com/ODudek/ConnX/contributors) who help make ConnX better

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=ODudek/ConnX&type=Date)](https://star-history.com/#ODudek/ConnX&Date)
