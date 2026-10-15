# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.9] - 2026-09-24

### Changed

- Refreshed the Crossplane APIs fork dependency to `v2.5.0-rc.0`.
- Hardened tag-only, xpkg-only release publishing for exact `vMAJOR.MINOR.PATCH` tags at the current `origin/master` commit.
- Publishes `linux_amd64` and `linux_arm64` xpkg artifacts, aliases the version as `latest`, and verifies matching digests and both platform manifests before creating the GitHub Release.

## [v0.4.1] - 2026-09-18

### Fixes

- Register `customproperty.plausible.m.crossplane.io/v1beta1`,
  `guest.plausible.m.crossplane.io/v1beta1`,
  `sharedlink.plausible.m.crossplane.io/v1beta1`, and
  `team.plausible.m.crossplane.io/v1beta1` in the scheme so that
  `mgr.Add(NewMRStateRecorder(...))` does not fail at startup with
  "no kind is registered for the type v1beta1.TeamList in scheme".

  Without this fix, `provider-plausible` CrashLoopBackOff at every
  startup (820+ restarts), preventing the metrics endpoint :8080 from
  ever opening.

## [Unreleased]

## [v0.4.2] - 2026-09-18

### Changed

- Replaced custom release workflow with the standardized
  `release-template.yml` (`make publish ...`) so that the build
  submodule's `xpkg.mk` / `imagelight.mk` directly produce both
  container image and `.xpkg` for the same digest. The previous
  workflow's `docker buildx build --file ./cluster/images/*/Dockerfile .`
  failed with "/bin/linux_amd64/provider: not found".

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