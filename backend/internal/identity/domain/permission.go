// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"fmt"
	"strings"
)

// Permission represents an authorization permission in "resource:action" format.
// Examples: "users:read", "users:*", "*:*" (super admin).
type Permission string

// NewPermission creates a Permission from a string in "resource:action" format.
// Wildcards are supported: "users:*" grants all actions on users resource,
// "*:*" grants all permissions (super admin).
func NewPermission(name string) (Permission, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("permission is required")
	}
	if name == "*:*" {
		return Permission(name), nil
	}
	parts := strings.SplitN(name, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("permission: must be in format resource:action")
	}
	if parts[1] == "*" {
		return Permission(name), nil
	}
	return Permission(name), nil
}

// String returns the string representation of the permission.
func (p Permission) String() string {
	return string(p)
}

// Matches checks if this permission grants access to the requested permission.
// Supports wildcard matching: "users:*" matches "users:read", "*:*" matches everything.
func (p Permission) Matches(other Permission) bool {
	if p == "*:*" || other == "*:*" {
		return true
	}
	if p == other {
		return true
	}
	pParts := strings.SplitN(p.String(), ":", 2)
	oParts := strings.SplitN(other.String(), ":", 2)
	if len(pParts) != 2 || len(oParts) != 2 {
		return false
	}
	if pParts[0] == oParts[0] && pParts[1] == "*" {
		return true
	}
	if pParts[0] == "*" && pParts[1] == oParts[1] {
		return true
	}
	return false
}

// Resource extracts the resource part from the permission string.
func (p Permission) Resource() string {
	parts := strings.SplitN(p.String(), ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// Action extracts the action part from the permission string.
func (p Permission) Action() string {
	parts := strings.SplitN(p.String(), ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
