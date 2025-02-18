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

## Installation

```bash
go get github.com/ODudek/ConnX
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

## Usage

Run the proxy server:

```bash
connx -config=/path/to/config.yaml
```

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

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [gnet](https://github.com/panjf2000/gnet) - A high-performance, lightweight, non-blocking, event-driven networking framework
