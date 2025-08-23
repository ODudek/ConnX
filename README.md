# ConnX

ConnX is a high-performance, lightweight HTTP reverse proxy and load balancer written in Go. It leverages the gnet framework for efficient network operations and provides features like health checking and round-robin load balancing.

## Features

- High-performance HTTP reverse proxy
- Round-robin load balancing
- Active health checking of backend servers
- Configuration via YAML file
- Real-time backend server status monitoring
- Request and error statistics tracking
- Customizable timeouts and intervals

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

backends:
  - "http://backend1:8080"
  - "http://backend2:8080"
  - "http://backend3:8080"

healthCheck:
  interval: 30  # seconds
  timeout: 5    # seconds
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

### Default Values

If not specified in the config file, the following default values are used:
- Server port: 8080
- Server host: "0.0.0.0"
- Health check interval: 30 seconds
- Health check timeout: 5 seconds

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
- **Server Pool**: Manages the collection of backend servers
- **Health Checker**: Monitors backend server health
- **Load Balancer**: Distributes requests across healthy backends

## Performance

ConnX is built with performance in mind:
- Uses gnet for efficient network operations
- Implements connection pooling
- Minimizes memory allocations
- Supports multicore processing

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
