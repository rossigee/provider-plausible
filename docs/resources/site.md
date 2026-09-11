# Site

**API Version**: `site.plausible.m.crossplane.io/v1beta1`

The `Site` resource represents a website in Plausible Analytics.

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `forProvider.domain` | string | yes | Domain name for the site |
| `forProvider.teamID` | string | no | Team ID to associate the site with |
| `forProvider.timezone` | string | no | IANA timezone (e.g., America/New_York) |
| `forProvider.newDomain` | string | no | New domain for updates |

## Example

```yaml
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: my-website
  namespace: production
spec:
  forProvider:
    domain: example.com
    teamID: "team-123"
    timezone: "America/New_York"
  providerConfigRef:
    name: default
```

## Behavior

- **Create**: Creates a new site in Plausible Analytics
- **Update**: Updates site domain (via newDomain), timezone
- **Delete**: Deletes the site from Plausible

## Status Fields

- `status.atProvider.id` — Unique site ID
- `status.atProvider.domain` — Current domain
- `status.atProvider.teamID` — Team ID
- `status.atProvider.createdAt` — Creation timestamp
- `status.atProvider.updatedAt` — Last update timestamp

## Important Notes

1. **Domain Uniqueness**: Domains must be unique within your Plausible account
2. **Domain Updates**: Use the `newDomain` field to update a site's domain
3. **Timezone**: Once set, timezone cannot be changed via the API
4. **Team Association**: Team ID cannot be changed after creation
