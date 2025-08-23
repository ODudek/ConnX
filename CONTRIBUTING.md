# Contributing to ConnX

Thank you for considering contributing to ConnX! We welcome contributions from everyone.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bug fix
4. Make your changes
5. Run tests and ensure they pass
6. Commit your changes with a clear commit message
7. Push to your fork and submit a pull request

## Development Setup

### Prerequisites

- Go 1.23 or later
- Make (optional, but recommended)

### Building

```bash
# Clone the repository
git clone https://github.com/ODudek/ConnX.git
cd ConnX

# Install dependencies
go mod download

# Build the project
make build
# or
go build -o bin/proxy cmd/proxy/main.go
```

### Running Tests

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

### Code Quality

We use `golangci-lint` for code quality checks. Install it and run:

```bash
golangci-lint run
```

## Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Write clear, descriptive commit messages
- Add tests for new functionality
- Document exported functions and types
- Keep functions small and focused

## Pull Request Guidelines

### Before Submitting

- Ensure all tests pass
- Run `golangci-lint` and fix any issues
- Update documentation if necessary
- Add tests for new features
- Ensure your branch is up-to-date with main

### Pull Request Process

1. **Title**: Use a clear, descriptive title
2. **Description**: Explain what your PR does and why
3. **Testing**: Describe how you tested your changes
4. **Documentation**: Update relevant documentation
5. **Breaking Changes**: Clearly mark any breaking changes

### PR Template

```
## Summary
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
```

## Reporting Issues

### Bug Reports

Please include:
- ConnX version
- Go version
- Operating system
- Configuration used
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

### Feature Requests

Please include:
- Clear description of the feature
- Use case and motivation
- Possible implementation approach
- Alternatives considered

## Architecture Guidelines

### Adding New Features

1. **Metrics**: All new features should include relevant metrics
2. **Configuration**: Use YAML configuration with sensible defaults
3. **Testing**: Include unit tests and integration tests where applicable
4. **Documentation**: Update CLAUDE.md and README.md
5. **Error Handling**: Implement proper error handling and logging

### Package Organization

- `cmd/` - Application entry points
- `internal/` - Private application code
- `pkg/` - Public library code (if any)
- `configs/` - Configuration files
- `docs/` - Additional documentation

## Communication

- Use GitHub Issues for bug reports and feature requests
- Use GitHub Discussions for questions and general discussion
- Be respectful and constructive in all interactions

## License

By contributing to ConnX, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in the project's acknowledgments. Significant contributions may be highlighted in release notes.

Thank you for your contributions! 🚀