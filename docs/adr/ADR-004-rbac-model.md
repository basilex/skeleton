# ADR-004: RBAC Model

## Status: Accepted

## Context

An access control system is needed:
- Different access levels for different users
- Flexibility in assigning permissions
- Simple integration with HTTP middleware

## Decision

Role-Based Access Control (RBAC):
- User → UserRole → Role → RolePermission → Permission
- Permission format: `resource:action`
- Wildcard support: `*:*`, `resource:*`
- Middleware chain: Authenticate → RequirePermission

## Consequences

### Positive
- Standard, understandable approach
- Easy to add new permissions
- Wildcard reduces rule duplication

### Negative
- Doesn't support attribute-based rules (ABAC)
- For fine-grained access, extension is needed