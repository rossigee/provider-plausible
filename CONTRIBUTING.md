# Contributing to provider-plausible

Thank you for your interest in contributing to the Crossplane Plausible provider! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to the [Crossplane Code of Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker
- kubectl
- A Kubernetes cluster (local or remote)
- Access to a Plausible Analytics instance with Sites API enabled

### Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/provider-plausible.git
   cd provider-plausible
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Generate code and manifests:
   ```bash
   make generate
   ```

5. Run tests:
   ```bash
   make test
   ```

## Development Workflow

### Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes
3. Add or update tests
4. Run tests and linting:
   ```bash
   make test lint
   ```

5. Generate updated manifests if needed:
   ```bash
   make generate
   ```

6. Commit your changes with a clear commit message
7. Push to your fork and create a pull request

### Code Generation

This project uses code generation for:
- CRD manifests
- Deep copy methods
- Webhook configurations

After modifying API types, always run:
```bash
make generate
```

### Testing

#### Unit Tests

Run unit tests with:
```bash
make test
```

Unit tests should:
- Test individual functions and methods
- Use mocks for external dependencies
- Have good coverage of edge cases
- Be fast and deterministic

#### Integration Tests

Integration tests require a running Kubernetes cluster and Plausible instance:
```bash
make test-integration
```

#### Manual Testing

1. Build and load the provider image:
   ```bash
   make docker-build docker-push
   ```

2. Install in your cluster:
   ```bash
   kubectl crossplane install provider your-registry/provider-plausible:dev
   ```

3. Create test resources and verify behavior

### Linting and Formatting

We use several tools to maintain code quality:

```bash
# Run all linting
make lint

# Format code
make fmt

# Vet code
make vet
```

## Pull Request Guidelines

### Before Submitting

- [ ] Tests pass locally
- [ ] Code is properly formatted
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] No merge commits (use rebase)

### PR Description

Include:
- Clear description of changes
- Motivation for the change
- Testing done
- Any breaking changes
- Related issues (if any)

### Review Process

1. All PRs require review from a maintainer
2. Address feedback promptly
3. Keep PRs focused and reasonably sized
4. Be responsive to questions and suggestions

## Architecture

### Provider Structure

```
provider-plausible/
├── apis/                 # API definitions
│   ├── site/            # Site resource API
│   └── goal/            # Goal resource API
├── internal/
│   ├── clients/         # Plausible API client
│   └── controller/      # Resource controllers
├── package/             # Crossplane package configuration
└── examples/            # Usage examples
```

### Key Components

- **API Types**: Define Kubernetes resources (CRDs)
- **Controllers**: Implement resource lifecycle management
- **Clients**: Handle Plausible API communication
- **Webhooks**: Validation and defaulting logic

### Crossplane Patterns

This provider follows standard Crossplane patterns:
- Managed resources with observe/create/update/delete lifecycle
- External name annotation for resource identification
- Condition-based status reporting
- Reference resolution for cross-resource dependencies

## API Guidelines

### Resource Design

- Follow Kubernetes API conventions
- Use clear, descriptive field names
- Provide comprehensive validation
- Include helpful documentation

### Backwards Compatibility

- Don't break existing APIs without version bumps
- Deprecate fields before removal
- Provide migration paths for breaking changes

## Documentation

### Code Documentation

- Public APIs must have godoc comments
- Complex logic should be well-commented
- Include examples in documentation

### User Documentation

- Update README.md for new features
- Add examples for new resources
- Update troubleshooting guides

### API Documentation

- Describe all fields in API types
- Include validation constraints
- Provide usage examples

## Testing Guidelines

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name string
        args args
        want want
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Best Practices

- Use table-driven tests
- Test both success and error paths
- Mock external dependencies
- Use meaningful test names
- Assert on specific error messages

## Release Process

### Versioning

We follow semantic versioning:
- Major: Breaking changes
- Minor: New features, backwards compatible
- Patch: Bug fixes

### Release Checklist

1. Update CHANGELOG.md
2. Update version in package configuration
3. Create and push git tag
4. Build and push container image
5. Create GitHub release
6. Update documentation

## Getting Help

- [GitHub Issues](https://github.com/crossplane-contrib/provider-plausible/issues) - Bug reports and feature requests
- [Crossplane Slack](https://slack.crossplane.io/) - #providers channel for questions
- [Crossplane Documentation](https://crossplane.io/docs/) - General Crossplane information

## Common Tasks

### Adding a New Resource

1. Define API types in `apis/`
2. Implement controller in `internal/controller/`
3. Add client methods in `internal/clients/`
4. Write comprehensive tests
5. Add examples and documentation
6. Update package configuration

### Adding a New Field

1. Update API types
2. Run `make generate`
3. Update controller logic if needed
4. Add tests for new functionality
5. Update examples and documentation

### Debugging

Enable debug logging:
```bash
make run DEBUG=true
```

Use delve for debugging:
```bash
dlv debug ./cmd/provider -- --debug
```

## Security

### Reporting Vulnerabilities

Please report security vulnerabilities to security@crossplane.io rather than opening public issues.

### Security Considerations

- Never log sensitive data (API keys, tokens)
- Validate all inputs
- Use secure defaults
- Follow principle of least privilege

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.