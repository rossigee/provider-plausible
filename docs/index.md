# Provider Plausible Analytics Documentation

A Crossplane v2 provider for managing Plausible Analytics resources with complete namespace isolation for multi-tenancy.

## Quick Links

- [Configuration](configuration.md) — Authentication and connection setup
- [Development](development.md) — Building, testing, and contributing

## Resource Documentation

### Analytics Resources

| Resource | API Group | Description |
|----------|-----------|-------------|
| [Site](resources/site.md) | `site.plausible.m.crossplane.io/v1beta1` | Website in Plausible Analytics |
| [Goal](resources/goal.md) | `goal.plausible.m.crossplane.io/v1beta1` | Conversion goals |

## Resource Relationships

- **Goal** references **Site** via `spec.siteDomainRef` or `spec.siteDomain`

## Important Notes

- Deleting a Site does NOT automatically delete its Goals
- Goals must be explicitly deleted before removing a Site
