package domain

import (
	"fmt"
	"strings"
)

type Permission string

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

func (p Permission) String() string {
	return string(p)
}

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

func (p Permission) Resource() string {
	parts := strings.SplitN(p.String(), ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func (p Permission) Action() string {
	parts := strings.SplitN(p.String(), ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
