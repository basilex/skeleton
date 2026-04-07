// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// emailRegex defines the validation pattern for email addresses.
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// Email represents a validated email address value object.
type Email string

// NewEmail creates an Email from a string after validation.
// The email is converted to lowercase and trimmed of whitespace.
func NewEmail(email string) (Email, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" {
		return "", fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(e) {
		return "", fmt.Errorf("email: must be a valid email address")
	}
	return Email(e), nil
}

// String returns the string representation of the email.
func (e Email) String() string {
	return string(e)
}
