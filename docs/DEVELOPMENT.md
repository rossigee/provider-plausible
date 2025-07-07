# Development Guide

This guide covers development setup, building, testing, and contributing to the Plausible provider.

## Table of Contents
- [Development Environment](#development-environment)
- [Building the Provider](#building-the-provider)
- [Running Locally](#running-locally)
- [Testing](#testing)
- [Adding New Resources](#adding-new-resources)
- [Debugging](#debugging)
- [Release Process](#release-process)

## Development Environment

### Prerequisites

- Go 1.21+
- Docker
- Kind or another Kubernetes cluster
- Kubectl
- Crossplane CLI (optional but recommended)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/crossplane-contrib/provider-plausible
   cd provider-plausible
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install code generation tools**
   ```bash
   go install -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
   go install -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet@master
   ```

4. **Setup a local Kubernetes cluster**
   ```bash
   kind create cluster --name crossplane-dev
   ```

5. **Install Crossplane**
   ```bash
   kubectl create namespace crossplane-system
   
   helm repo add crossplane-stable https://charts.crossplane.io/stable
   helm repo update
   
   helm install crossplane \
     crossplane-stable/crossplane \
     --namespace crossplane-system \
     --set args='{"--debug","--enable-management-policies"}'
   ```

## Building the Provider

### Generate Code

Always regenerate code after modifying API types:

```bash
make generate
```

This generates:
- CRD YAML files in `package/crds/`
- DeepCopy methods (`zz_generated.deepcopy.go`)
- Managed resource methods (`zz_generated.managed.go`)

### Build Binary

```bash
make build
```

The binary will be at `_output/bin/provider`

### Build Docker Image

```bash
# Build for current platform
make docker-build

# Build for multiple platforms
make docker-build PLATFORMS="linux_amd64 linux_arm64"
```

### Build Provider Package

```bash
make xpkg-build
```

The package will be at `_output/xpkg/`

## Running Locally

### Option 1: Out-of-Cluster

1. **Export kubeconfig**
   ```bash
   export KUBECONFIG=~/.kube/config
   ```

2. **Create ProviderConfig**
   ```bash
   kubectl apply -f examples/provider/secret.yaml
   kubectl apply -f examples/provider/config.yaml
   ```

3. **Run the provider**
   ```bash
   make run
   ```

### Option 2: In-Cluster with Kind

1. **Build and load image**
   ```bash
   make docker-build
   kind load docker-image crossplane/provider-plausible:latest --name crossplane-dev
   ```

2. **Install provider**
   ```yaml
   apiVersion: pkg.crossplane.io/v1
   kind: Provider
   metadata:
     name: provider-plausible
   spec:
     package: crossplane/provider-plausible:latest
     packagePullPolicy: Never  # Use local image
   ```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/clients -v

# Run with race detection
go test -race ./...
```

### Integration Tests

Create a test environment:

```bash
# 1. Apply test credentials
kubectl create secret generic plausible-test-creds \
  --from-literal=credentials='{"apiKey":"test-key"}' \
  -n crossplane-system

# 2. Apply test resources
kubectl apply -f examples/provider/config.yaml
kubectl apply -f examples/site/site.yaml

# 3. Check resource status
kubectl describe site.site.plausible.crossplane.io example-site
```

### E2E Tests

```bash
# Run e2e tests (requires real Plausible API access)
export PLAUSIBLE_API_KEY="your-test-key"
go test -tags=e2e ./test/e2e/...
```

## Adding New Resources

### 1. Define API Types

Create new types in `apis/<group>/v1alpha1/types.go`:

```go
// SharedLinkParameters are the configurable fields of a SharedLink
type SharedLinkParameters struct {
    SiteDomain string `json:"siteDomain"`
    Name       string `json:"name"`
}

// SharedLinkObservation are the observable fields of a SharedLink
type SharedLinkObservation struct {
    ID  string `json:"id,omitempty"`
    URL string `json:"url,omitempty"`
}

// SharedLinkSpec defines the desired state of a SharedLink
type SharedLinkSpec struct {
    xpv1.ResourceSpec `json:",inline"`
    ForProvider       SharedLinkParameters `json:"forProvider"`
}

// SharedLinkStatus represents the observed state of a SharedLink
type SharedLinkStatus struct {
    xpv1.ResourceStatus `json:",inline"`
    AtProvider          SharedLinkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// SharedLink is a managed resource
type SharedLink struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   SharedLinkSpec   `json:"spec"`
    Status SharedLinkStatus `json:"status,omitempty"`
}
```

### 2. Add Controller

Create controller in `internal/controller/sharedlink/sharedlink.go`:

```go
func Setup(mgr ctrl.Manager, o controller.Options) error {
    name := managed.ControllerName(v1alpha1.SharedLinkGroupKind)
    
    r := managed.NewReconciler(mgr,
        resource.ManagedKind(v1alpha1.SharedLinkGroupVersionKind),
        managed.WithExternalConnecter(&connector{
            kube:         mgr.GetClient(),
            usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
            newServiceFn: clients.NewClient,
        }),
        // ... other options
    )
    
    return ctrl.NewControllerManagedBy(mgr).
        Named(name).
        For(&v1alpha1.SharedLink{}).
        Complete(r)
}
```

### 3. Implement External Client

Add methods to handle CRUD operations:

```go
func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
    // Check if resource exists
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
    // Create the resource
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
    // Update the resource
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
    // Delete the resource
}
```

### 4. Register Controller

Add to `internal/controller/controller.go`:

```go
func Setup(mgr ctrl.Manager, o controller.Options) error {
    for _, setup := range []func(ctrl.Manager, controller.Options) error{
        config.Setup,
        site.Setup,
        goal.Setup,
        sharedlink.Setup,  // Add new controller
    } {
        if err := setup(mgr, o); err != nil {
            return err
        }
    }
    return nil
}
```

### 5. Generate and Test

```bash
make generate
go test ./internal/controller/sharedlink
make run
```

## Debugging

### Enable Debug Logging

```bash
# When running locally
make run ARGS="--debug"

# In-cluster
kubectl edit deployment/provider-plausible-*
# Add --debug to container args
```

### Common Issues

1. **CRD Installation Issues**
   ```bash
   # Reinstall CRDs
   kubectl delete crd sites.site.plausible.crossplane.io
   make install-crds
   ```

2. **RBAC Issues**
   ```bash
   # Check provider service account permissions
   kubectl describe clusterrole provider-plausible-*
   ```

3. **API Client Issues**
   ```bash
   # Test API connectivity
   curl -H "Authorization: Bearer YOUR_KEY" https://plausible.io/api/v1/sites
   ```

### Debugging Tools

```bash
# Watch provider logs
kubectl logs -f -n crossplane-system deployment/provider-plausible-*

# Describe problematic resources
kubectl describe site.site.plausible.crossplane.io my-site

# Check events
kubectl get events --field-selector involvedObject.name=my-site

# Enable verbose API logging
export PLAUSIBLE_DEBUG=true
make run
```

## Release Process

### 1. Update Version

```bash
# Update version in:
# - Makefile
# - package/crossplane.yaml
```

### 2. Generate Artifacts

```bash
make generate
make build
make xpkg-build
```

### 3. Run Tests

```bash
make test
make e2e-test  # Requires test environment
```

### 4. Tag Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 5. Build and Push Images

```bash
make docker-build docker-push REGISTRY=crossplane
make xpkg-build xpkg-push REGISTRY=crossplane
```

### 6. Create GitHub Release

Include:
- Changelog
- Breaking changes
- Migration guide (if applicable)

## Code Style

### Go Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports`
- Add comments for exported types and functions
- Keep functions focused and small

### Commit Messages

Follow conventional commits:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `chore:` Build process, dependencies

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Run `make reviewable` (formats, tests, builds)
5. Submit PR with clear description
6. Address review feedback

## Useful Commands

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Update dependencies
go mod tidy

# Verify module
go mod verify

# Clean build artifacts
make clean

# Full pre-commit check
make reviewable
```