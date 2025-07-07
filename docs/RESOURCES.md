# Plausible Provider Resources

This document provides detailed information about all resources supported by the Plausible provider.

## Table of Contents
- [Site Resource](#site-resource)
- [Goal Resource](#goal-resource)
- [Resource Relationships](#resource-relationships)
- [Common Patterns](#common-patterns)

## Site Resource

The `Site` resource represents a website in Plausible Analytics.

### API Version
- Group: `site.plausible.crossplane.io`
- Version: `v1alpha1`
- Kind: `Site`

### Specification

```yaml
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: my-website
spec:
  forProvider:
    # Required: The domain name for the site
    domain: example.com
    
    # Optional: Team ID to associate the site with
    # If not specified, uses the default team
    teamID: "team-123"
    
    # Optional: Timezone for the site
    # Must be a valid IANA timezone (e.g., "America/New_York", "Europe/London")
    # Defaults to UTC if not specified
    timezone: "America/New_York"
    
    # Optional: New domain for updating an existing site
    # Only used during updates, leave empty during creation
    newDomain: "new-example.com"
  
  # Reference to the ProviderConfig
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique ID assigned by Plausible
    id: "site-abc123"
    
    # Current domain of the site
    domain: "example.com"
    
    # Team ID the site belongs to
    teamID: "team-123"
    
    # Timestamps
    createdAt: "2023-01-01T00:00:00Z"
    updatedAt: "2023-01-02T00:00:00Z"
  
  # Crossplane resource conditions
  conditions:
  - type: Ready
    status: "True"
    reason: Available
    lastTransitionTime: "2023-01-01T00:00:00Z"
```

### Examples

#### Basic Site
```yaml
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: company-blog
spec:
  forProvider:
    domain: blog.company.com
  providerConfigRef:
    name: default
```

#### Site with Team and Timezone
```yaml
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: regional-site
spec:
  forProvider:
    domain: uk.company.com
    teamID: "uk-team"
    timezone: "Europe/London"
  providerConfigRef:
    name: default
```

#### Updating a Site Domain
```yaml
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: company-site
spec:
  forProvider:
    domain: old-domain.com
    newDomain: new-domain.com  # This triggers a domain update
  providerConfigRef:
    name: default
```

### Important Notes

1. **Domain Uniqueness**: Domains must be unique within your Plausible account
2. **Domain Updates**: Use the `newDomain` field to update a site's domain
3. **Timezone**: Once set, timezone cannot be changed via the API
4. **Team Association**: Team ID cannot be changed after creation

## Goal Resource

The `Goal` resource represents a conversion goal in Plausible Analytics.

### API Version
- Group: `goal.plausible.crossplane.io`
- Version: `v1alpha1`
- Kind: `Goal`

### Specification

```yaml
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: my-goal
spec:
  forProvider:
    # Site association (use one of the following methods)
    
    # Method 1: Direct domain reference
    siteDomain: "example.com"
    
    # Method 2: Reference to a Site resource
    siteDomainRef:
      name: my-website
    
    # Method 3: Selector (not yet implemented)
    # siteDomainSelector:
    #   matchLabels:
    #     environment: production
    
    # Required: Type of goal
    # Must be either "event" or "page"
    goalType: event
    
    # Required for event goals: The event name to track
    eventName: "Signup"
    
    # Required for page goals: The page path to track
    pagePath: "/thank-you"
  
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique ID assigned by Plausible
    id: "goal-xyz789"
    
    # Goal type (event or page)
    goalType: "event"
    
    # Event name (for event goals)
    eventName: "Signup"
    
    # Page path (for page goals)
    pagePath: ""
    
    # Creation timestamp
    createdAt: "2023-01-01T00:00:00Z"
  
  conditions:
  - type: Ready
    status: "True"
    reason: Available
```

### Examples

#### Event Goal
```yaml
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: signup-goal
spec:
  forProvider:
    siteDomainRef:
      name: my-website
    goalType: event
    eventName: "Signup"
  providerConfigRef:
    name: default
```

#### Page Goal
```yaml
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: conversion-goal
spec:
  forProvider:
    siteDomain: "example.com"
    goalType: page
    pagePath: "/purchase/complete"
  providerConfigRef:
    name: default
```

#### Multiple Goals for One Site
```yaml
# Newsletter signup
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: newsletter-goal
spec:
  forProvider:
    siteDomainRef:
      name: company-site
    goalType: event
    eventName: "Newsletter Signup"
  providerConfigRef:
    name: default
---
# Contact form submission
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: contact-goal
spec:
  forProvider:
    siteDomainRef:
      name: company-site
    goalType: event
    eventName: "Contact Form"
  providerConfigRef:
    name: default
---
# Thank you page visit
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: thankyou-goal
spec:
  forProvider:
    siteDomainRef:
      name: company-site
    goalType: page
    pagePath: "/thank-you"
  providerConfigRef:
    name: default
```

### Important Notes

1. **Goal Types**: Only "event" and "page" types are supported
2. **Event Names**: Must match exactly what your website sends
3. **Page Paths**: Should include the leading slash
4. **Immutability**: Goals cannot be updated after creation
5. **Uniqueness**: The combination of site + goal type + event/page must be unique

## Resource Relationships

### Site → Goals Relationship

Goals depend on Sites. The relationship can be established in two ways:

1. **Direct Domain Reference**
   ```yaml
   siteDomain: "example.com"
   ```

2. **Resource Reference**
   ```yaml
   siteDomainRef:
     name: my-website
   ```

Using resource references is recommended as it:
- Creates an explicit dependency
- Ensures the site exists before creating the goal
- Automatically handles site deletion order

### Deletion Behavior

- Deleting a Site does NOT automatically delete its Goals
- Goals must be explicitly deleted before removing a Site
- Use Kubernetes finalizers or owner references for cascading deletion

## Common Patterns

### Complete Site Setup

```yaml
# 1. Create the site
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: ecommerce-site
  labels:
    app: ecommerce
    environment: production
spec:
  forProvider:
    domain: shop.example.com
    timezone: "America/New_York"
  providerConfigRef:
    name: default
---
# 2. Create conversion goals
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: ecommerce-purchase
  labels:
    app: ecommerce
    goal-type: conversion
spec:
  forProvider:
    siteDomainRef:
      name: ecommerce-site
    goalType: event
    eventName: "Purchase"
  providerConfigRef:
    name: default
---
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: ecommerce-add-to-cart
  labels:
    app: ecommerce
    goal-type: engagement
spec:
  forProvider:
    siteDomainRef:
      name: ecommerce-site
    goalType: event
    eventName: "Add to Cart"
  providerConfigRef:
    name: default
```

### Multi-Environment Setup

```yaml
# Development site
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: app-dev
  labels:
    environment: development
spec:
  forProvider:
    domain: dev.app.example.com
  providerConfigRef:
    name: default
---
# Production site
apiVersion: site.plausible.crossplane.io/v1alpha1
kind: Site
metadata:
  name: app-prod
  labels:
    environment: production
spec:
  forProvider:
    domain: app.example.com
    timezone: "America/New_York"
  providerConfigRef:
    name: default
---
# Shared goal configuration for both environments
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: signup-goal-dev
spec:
  forProvider:
    siteDomainRef:
      name: app-dev
    goalType: event
    eventName: "User Signup"
  providerConfigRef:
    name: default
---
apiVersion: goal.plausible.crossplane.io/v1alpha1
kind: Goal
metadata:
  name: signup-goal-prod
spec:
  forProvider:
    siteDomainRef:
      name: app-prod
    goalType: event
    eventName: "User Signup"
  providerConfigRef:
    name: default
```