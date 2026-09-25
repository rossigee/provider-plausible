# Getting Started

Guide to getting started with the provider.

## Installation

Install the provider:

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-plausible:v0.4.9
```

## Prerequisites

- Kubernetes cluster with Crossplane v2.5+ installed

## Quick Start

1. Create a ProviderConfig:

```yaml
apiVersion: plausible.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: plausible-credentials
      namespace: crossplane-system
```
