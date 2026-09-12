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

### Other Resources

| Resource | API Group | Description |
|----------|-----------|-------------|
| Guest | `guest.plausible.m.crossplane.io/v1beta1` | Site guest access |
| SharedLink | `sharedlink.plausible.m.crossplane.io/v1beta1` | Shared dashboard links |
| Team | `team.plausible.m.crossplane.io/v1beta1` | Teams |
| CustomProperty | `customproperty.plausible.m.crossplane.io/v1beta1` | Custom event properties |
| ProviderConfig | `plausible.m.crossplane.io/v1beta1` | Credentials (cluster-scoped) |

## Important Notes

- Deleting a Site does NOT automatically delete its Goals
- Goals must be explicitly deleted before removing a Site

## API Coverage Gaps

Plausible API surface not yet modeled: funnels, segments beyond custom properties, imports/exports, user preferences, consolidated team invitations, and stats API query configuration.
