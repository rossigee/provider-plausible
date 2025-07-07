# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-07-07

### Added
- Initial release of provider-plausible
- Site resource management with full CRUD operations
- Goal resource support for event and page goals
- Multi-tenant support via team assignments
- Domain migration capabilities
- Comprehensive unit tests
- GitHub Actions CI/CD pipeline
- Complete documentation and examples

### Features
- **Site Management**: Create, read, update, and delete Plausible Analytics sites
- **Goal Tracking**: Manage conversion goals (event-based and page-based)
- **Domain Migration**: Support for updating site domains while preserving analytics data
- **Team Support**: Multi-tenant deployment with team-based site organization
- **Cross-references**: Reference sites from goals using Kubernetes selectors
- **Status Reporting**: Rich status information with conditions and observations

### API Resources
- `Site` (v1alpha1): Manage Plausible Analytics sites
- `Goal` (v1alpha1): Manage conversion tracking goals
- `ProviderConfig` (v1beta1): Configure provider authentication and settings

### Technical Details
- Built on Crossplane provider framework
- Uses Plausible Sites API for programmatic site management
- Support for custom Plausible instances via baseURL configuration
- Comprehensive error handling with 404 detection
- Observability through OpenTelemetry (inherited from Crossplane)

### Documentation
- Complete README with installation and usage examples
- API reference documentation
- Troubleshooting guide
- Contributing guidelines
- Development setup instructions

### Testing
- Unit tests for all client operations
- Controller behavior tests
- Mock implementations for testing
- CI pipeline with automated testing

### Build & Deployment
- Docker containerization
- Crossplane package (xpkg) format
- Multi-registry support (Harbor, Docker Hub)
- Automated releases via GitHub Actions

## [0.1.4] - 2025-07-07

### Added
- Basic provider structure and implementation
- Initial Site resource support
- Plausible API client implementation

### Notes
- This version was used for initial development and testing
- Contains working implementation but lacks comprehensive testing and documentation