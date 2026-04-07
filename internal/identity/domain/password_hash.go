package domain

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHash string

func NewPasswordHash(plainPassword string) (PasswordHash, error) {
	if plainPassword == "" {
		return "", fmt.Errorf("password is required")
	}
	if len(plainPassword) < 8 {
		return "", fmt.Errorf("password: must be at least 8 characters")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return PasswordHash(hashed), nil
}

func (h PasswordHash) String() string {
	return string(h)
}

func (h PasswordHash) Matches(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(plainPassword))
	return err == nil
}
