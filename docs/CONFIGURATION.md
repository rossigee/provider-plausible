# Provider Configuration Guide

This guide walks through all configuration options for the Plausible provider.

## Table of Contents
- [Prerequisites](#prerequisites)
- [API Key Setup](#api-key-setup)
- [Provider Installation](#provider-installation)
- [ProviderConfig Setup](#providerconfig-setup)
- [Self-Hosted Plausible](#self-hosted-plausible)
- [Troubleshooting](#troubleshooting)

## Prerequisites

1. A Kubernetes cluster with Crossplane v1.20+ installed
2. A Plausible Analytics account (either cloud or self-hosted)
3. A Plausible API key with appropriate permissions

## API Key Setup

### Obtaining an API Key

1. Log into your Plausible Analytics account
2. Navigate to Account Settings → API Keys
3. Click "Create API Key"
4. Give your key a descriptive name (e.g., "crossplane-provider")
5. **Important**: Contact Plausible support to enable Site Provisioning API access for your key

### Required Permissions

Your API key needs the following permissions:
- Sites API: Read, Write, Delete
- Goals API: Read, Write, Delete

## Provider Installation

### Option 1: Using Crossplane CLI

```bash
kubectl crossplane install provider crossplane/provider-plausible:latest
```

### Option 2: Using Kubernetes Manifest

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-plausible
spec:
  package: crossplane/provider-plausible:latest
  # Optional: specify a specific version
  # package: crossplane/provider-plausible:v0.1.0
```

Apply the manifest:
```bash
kubectl apply -f provider-plausible.yaml
```

### Verify Installation

Check that the provider is healthy:
```bash
kubectl get providers.pkg.crossplane.io
```

## ProviderConfig Setup

### 1. Create the Credentials Secret

The provider expects credentials in JSON format:

```bash
# Create the secret directly
kubectl create secret generic plausible-credentials \
  --from-literal=credentials='{"apiKey":"YOUR_PLAUSIBLE_API_KEY"}' \
  -n crossplane-system

# Or create from a file
echo '{"apiKey":"YOUR_PLAUSIBLE_API_KEY"}' > credentials.json
kubectl create secret generic plausible-credentials \
  --from-file=credentials=credentials.json \
  -n crossplane-system
rm credentials.json  # Clean up
```

### 2. Create the ProviderConfig

Create a file named `provider-config.yaml`:

```yaml
apiVersion: plausible.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: plausible-credentials
      namespace: crossplane-system
      key: credentials
  # For Plausible Cloud, baseURL is optional
  # For self-hosted, specify your instance URL
  # baseURL: https://plausible.io
```

Apply the configuration:
```bash
kubectl apply -f provider-config.yaml
```

### 3. Verify ProviderConfig

```bash
kubectl get providerconfigs.plausible.crossplane.io
kubectl describe providerconfig.plausible.crossplane.io default
```

## Self-Hosted Plausible

If you're using a self-hosted Plausible instance:

```yaml
apiVersion: plausible.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: self-hosted
spec:
  credentials:
    source: Secret
    secretRef:
      name: plausible-credentials
      namespace: crossplane-system
      key: credentials
  # Required for self-hosted instances
  baseURL: https://analytics.example.com
```

### Multiple ProviderConfigs

You can create multiple ProviderConfigs for different Plausible instances:

```yaml
# Cloud instance
apiVersion: plausible.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: plausible-cloud
spec:
  credentials:
    source: Secret
    secretRef:
      name: cloud-credentials
      namespace: crossplane-system
      key: credentials
---
# Self-hosted instance
apiVersion: plausible.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: plausible-selfhosted
spec:
  credentials:
    source: Secret
    secretRef:
      name: selfhosted-credentials
      namespace: crossplane-system
      key: credentials
  baseURL: https://analytics.internal.company.com
```

Then reference the specific config in your resources:

```yaml
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: my-site
spec:
  providerConfigRef:
    name: plausible-selfhosted  # Use specific config
  forProvider:
    domain: example.com
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   ```
   cannot get credentials: cannot extract credentials
   ```
   - Verify the secret exists and contains valid JSON
   - Check the secret namespace matches the ProviderConfig

2. **API Permission Errors**
   ```
   API request failed with status 403: Forbidden
   ```
   - Ensure your API key has Site Provisioning API access
   - Contact Plausible support if needed

3. **Connection Errors**
   ```
   failed to send http request
   ```
   - For self-hosted: verify the baseURL is correct
   - Check network connectivity from the provider pod

### Debug Commands

```bash
# Check provider logs
kubectl logs -n crossplane-system deployment/provider-plausible-*

# Verify secret contents (be careful with sensitive data)
kubectl get secret plausible-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d

# Check provider config status
kubectl describe providerconfig.plausible.crossplane.io default

# List all Plausible resources
kubectl get sites.site.plausible.crossplane.io
kubectl get goals.goal.plausible.crossplane.io
```