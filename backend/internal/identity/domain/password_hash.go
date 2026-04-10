// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHash represents a bcrypt hashed password.
type PasswordHash string

// NewPasswordHash creates a new bcrypt hash from a plain text password.
// The password must be at least 8 characters long.
func NewPasswordHash(plainPassword string) (PasswordHash, error) {
	if plainPassword == "" {
		return "", fmt.Errorf("password is required")
	}
	if len(plainPassword) < 8 {
		return "", fmt.Errorf("password: must be at least 8 characters")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return PasswordHash(hashed), nil
}

// String returns the string representation of the password hash.
func (h PasswordHash) String() string {
	return string(h)
}

// Matches checks if the provided plain text password matches the stored hash.
func (h PasswordHash) Matches(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(plainPassword))
	return err == nil
}
