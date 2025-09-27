# Crossplane Provider for Plausible Analytics (v2 Native)

**✅ BUILD STATUS: WORKING** - v2 native provider with namespaced resources (v1.2.0)

[![Build Status](https://github.com/crossplane-contrib/provider-plausible/workflows/CI/badge.svg)](https://github.com/crossplane-contrib/provider-plausible/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/crossplane-contrib/provider-plausible)](https://goreportcard.com/report/github.com/crossplane-contrib/provider-plausible)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A [Crossplane v2 native](https://crossplane.io/) provider for managing [Plausible Analytics](https://plausible.io/) resources programmatically through Kubernetes with namespace isolation.

## Container Registry
- **Primary**: `ghcr.io/rossigee/provider-plausible:v1.2.0`
- **Harbor**: Available via environment configuration
- **Upbound**: Available via environment configuration

## Overview

The Plausible provider enables platform teams to manage Plausible Analytics sites and goals as Kubernetes resources. This allows for:
- Declarative configuration of analytics infrastructure
- GitOps workflows for analytics management
- Integration with existing Kubernetes tooling
- Consistent lifecycle management across environments

## Features

- **🚀 v2 Native**: Namespaced resources for better multi-tenancy and isolation
- **Site Management**: Create, update, and delete Plausible sites
- **Goal Tracking**: Manage conversion goals with event and page-based tracking
- **Multi-tenant**: Support for team-based site management with namespace isolation
- **Cross-references**: Reference sites from goals using Kubernetes selectors
- **Observability**: Built-in status reporting and condition management

## Prerequisites

- Kubernetes cluster with Crossplane v1.20+ installed
- Plausible Analytics account with Sites API access
- API key with `sites:provision:*` scope

## Installation

### Quick Start

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-plausible:v1.2.0
```

### Declarative Installation

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-plausible
spec:
  package: ghcr.io/rossigee/provider-plausible:v1.2.0
```

## Configuration

### 1. Create API Key Secret

Create a Plausible API key with `sites:provision:*` scope in your Plausible dashboard, then create a Kubernetes secret:

```bash
kubectl create secret generic plausible-credentials \
  --from-literal=credentials='{"apiKey":"YOUR_API_KEY_HERE"}' \
  -n crossplane-system
```

### 2. Configure Provider

```yaml
apiVersion: plausible.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  # baseURL: "https://plausible.yourdomain.com"  # Optional: defaults to plausible.io
  credentials:
    source: Secret
    secretRef:
      name: plausible-credentials
      namespace: crossplane-system
      key: credentials
```

## Usage Examples

### Basic Site Creation

```yaml
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: company-website
  namespace: production
spec:
  forProvider:
    domain: company.example.com
    timezone: "America/New_York"
  providerConfigRef:
    name: default
```

### Site with Team Assignment

```yaml
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: team-site
  namespace: team-marketing
spec:
  forProvider:
    domain: team.example.com
    timezone: "UTC"
    teamID: "team-123"
  providerConfigRef:
    name: default
```

### Updating Site Domain

```yaml
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: rebranded-site
  namespace: production
spec:
  forProvider:
    domain: old-domain.com
    newDomain: new-domain.com  # Will migrate analytics data
    timezone: "Europe/London"
  providerConfigRef:
    name: default
```

### Event Goal Creation

```yaml
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: signup-conversion
  namespace: production
spec:
  forProvider:
    siteDomainRef:
      name: company-website
    goalType: event
    eventName: "Sign Up"
  providerConfigRef:
    name: default
```

### Page Goal Creation

```yaml
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: thank-you-page
  namespace: production
spec:
  forProvider:
    siteDomainRef:
      name: company-website
    goalType: page
    pagePath: "/thank-you"
  providerConfigRef:
    name: default
```

### Cross-Reference Example

```yaml
# Create multiple sites and goals that reference them
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: marketing-site
  namespace: marketing
  labels:
    team: marketing
spec:
  forProvider:
    domain: marketing.example.com
    timezone: "America/Los_Angeles"
---
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: marketing-signup
  namespace: marketing
spec:
  forProvider:
    siteDomainRef:
      name: marketing-site
    goalType: event
    eventName: "Newsletter Signup"
---
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: marketing-download
  namespace: marketing
spec:
  forProvider:
    siteDomainRef:
      name: marketing-site
    goalType: page
    pagePath: "/download"
```

## Resource Reference

### Site Resource

#### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain` | string | Yes | Website domain (e.g., "example.com") |
| `timezone` | string | No | Site timezone (default: "UTC") |
| `newDomain` | string | No | New domain for migration |
| `teamID` | string | No | Team ID for multi-tenant setups |

#### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `siteID` | string | Plausible site identifier |
| `conditions` | []Condition | Resource status conditions |

### Goal Resource

#### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `siteDomainRef.name` | string | Yes | Reference to Site resource |
| `goalType` | string | Yes | "event" or "page" |
| `eventName` | string | Conditional | Required for event goals |
| `pagePath` | string | Conditional | Required for page goals |

#### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `goalID` | string | Plausible goal identifier |
| `conditions` | []Condition | Resource status conditions |

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- crossplane CLI

### Building from Source

```bash
# Clone the repository
git clone https://github.com/crossplane-contrib/provider-plausible.git
cd provider-plausible

# Generate code and manifests
make generate

# Run tests
make test

# Build binary
make build

# Build Docker image
make docker-build

# Package for Crossplane
make build-package
```

### Running Locally

```bash
# Start the provider locally (requires kubeconfig)
make run
```

### Testing

```bash
# Run unit tests
make test

# Run integration tests (requires Plausible instance)
make test-integration

# Run linting
make lint
```

## Troubleshooting

### Common Issues

#### 401 Unauthorized
- Verify API key has `sites:provision:*` scope
- Check API key is correctly formatted in secret
- Ensure Plausible instance has Sites API enabled

#### 406 Not Acceptable
- Verify your Plausible account has Sites API access
- Check API key has the correct `sites:provision:*` scope

#### Connection Errors
- Verify `baseURL` in ProviderConfig
- Check network connectivity to Plausible instance
- Verify TLS certificates if using custom domain

### Debug Mode

Enable debug logging by setting the provider's log level:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-plausible
spec:
  package: ghcr.io/rossigee/provider-plausible:v1.2.0
  runtimeConfigRef:
    name: debug-config
---
apiVersion: pkg.crossplane.io/v1beta1
kind: DeploymentRuntimeConfig
metadata:
  name: debug-config
spec:
  deploymentTemplate:
    spec:
      template:
        spec:
          containers:
          - name: package-runtime
            args:
            - --debug
```

### Getting Help

- [GitHub Issues](https://github.com/crossplane-contrib/provider-plausible/issues)
- [Crossplane Slack](https://slack.crossplane.io/) - #providers channel
- [Plausible Documentation](https://plausible.io/docs)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test lint`
6. Submit a pull request

## Security

For security concerns, please email security@crossplane.io rather than opening a public issue.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Crossplane](https://crossplane.io/) team for the excellent provider framework
- [Plausible Analytics](https://plausible.io/) for the privacy-focused analytics platform