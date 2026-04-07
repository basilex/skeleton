package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "users:read", false},
		{"wildcard", "*:*", false},
		{"resource wildcard", "users:*", false},
		{"empty", "", true},
		{"no colon", "usersread", true},
		{"empty resource", ":read", true},
		{"empty action", "users:", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPermission(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, Permission(tt.input), p)
			}
		})
	}
}

func TestPermissionMatches(t *testing.T) {
	tests := []struct {
		name     string
		perm     string
		other    string
		expected bool
	}{
		{"exact match", "users:read", "users:read", true},
		{"no match", "users:read", "users:write", false},
		{"wildcard all", "*:*", "users:read", true},
		{"wildcard all reverse", "users:read", "*:*", true},
		{"resource wildcard", "users:*", "users:read", true},
		{"resource wildcard no match", "users:*", "roles:read", false},
		{"different resource", "users:read", "roles:read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := NewPermission(tt.perm)
			o, _ := NewPermission(tt.other)
			require.Equal(t, tt.expected, p.Matches(o))
		})
	}
}

func TestPermissionResourceAndAction(t *testing.T) {
	p, _ := NewPermission("users:read")
	require.Equal(t, "users", p.Resource())
	require.Equal(t, "read", p.Action())

	w, _ := NewPermission("*:*")
	require.Equal(t, "*", w.Resource())
	require.Equal(t, "*", w.Action())
}
