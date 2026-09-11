# Goal

**API Version**: `goal.plausible.m.crossplane.io/v1beta1`

The `Goal` resource represents a conversion goal in Plausible Analytics.

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `forProvider.siteDomain` | string | no* | Direct domain reference |
| `forProvider.siteDomainRef` | object | no* | Reference to a Site resource |
| `forProvider.goalType` | string | yes | Goal type (event or page) |
| `forProvider.eventName` | string | no** | Event name (for event goals) |
| `forProvider.pagePath` | string | no** | Page path (for page goals) |

*Either siteDomain or siteDomainRef is required
**Required based on goalType (event requires eventName, page requires pagePath)

## Example

Event Goal:
```yaml
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: signup-goal
  namespace: production
spec:
  forProvider:
    siteDomainRef:
      name: my-website
    goalType: event
    eventName: "Signup"
  providerConfigRef:
    name: default
```

Page Goal:
```yaml
apiVersion: goal.plausible.m.crossplane.io/v1beta1
kind: Goal
metadata:
  name: conversion-goal
  namespace: production
spec:
  forProvider:
    siteDomain: "example.com"
    goalType: page
    pagePath: "/purchase/complete"
  providerConfigRef:
    name: default
```

## Behavior

- **Create**: Creates a new conversion goal
- **Update**: Goals are immutable after creation
- **Delete**: Deletes the goal

## Status Fields

- `status.atProvider.id` — Unique goal ID
- `status.atProvider.goalType` — Goal type (event or page)
- `status.atProvider.eventName` — Event name
- `status.atProvider.pagePath` — Page path
- `status.atProvider.createdAt` — Creation timestamp

## Resource Relationships

Goals reference Sites via `spec.siteDomainRef`:
```yaml
spec:
  forProvider:
    siteDomainRef:
      name: my-website
```

## Important Notes

1. **Goal Types**: Only "event" and "page" types are supported
2. **Event Names**: Must match exactly what your website sends
3. **Page Paths**: Should include the leading slash
4. **Immutability**: Goals cannot be updated after creation
5. **Uniqueness**: The combination of site + goal type + event/page must be unique
