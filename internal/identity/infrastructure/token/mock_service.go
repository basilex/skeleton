// Package token provides token generation and validation implementations.
// This package contains infrastructure implementations for JWT token handling
// and mock token services for testing purposes.
package token

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

// MockTokenService is a mock implementation of the token service interface.
// It generates predictable tokens for testing purposes and does not perform
// any cryptographic operations. Use only in test environments.
type MockTokenService struct{}

// GenerateAccessToken generates a mock access token containing the user ID and permissions.
// The token format is "access-{userID}-{timestamp}" for predictable testing.
func (s *MockTokenService) GenerateAccessToken(userID domain.UserID, roles []domain.Role) (string, error) {
	permissions := make([]string, 0)
	for _, role := range roles {
		for _, p := range role.Permissions() {
			permissions = append(permissions, p.String())
		}
	}
	return fmt.Sprintf("access-%s-%d", string(userID), time.Now().Unix()), nil
}

// GenerateRefreshToken generates a mock refresh token using a UUID.
// Returns a token in the format "refresh-{uuid}".
func (s *MockTokenService) GenerateRefreshToken() (string, error) {
	return "refresh-" + uuid.NewV7().String(), nil
}

// ValidateAccessToken validates a mock access token and returns the claims.
// Accepts any token starting with "access-" prefix and returns mock claims.
func (s *MockTokenService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	if len(tokenString) < 7 || tokenString[:7] != "access-" {
		return nil, fmt.Errorf("invalid access token format")
	}
	return &domain.TokenClaims{
		UserID:      domain.UserID("mock-user-id"),
		Roles:       []string{"admin"},
		Permissions: []string{"*:*"},
	}, nil
}

// ValidateRefreshToken validates a mock refresh token and returns the associated user ID.
// Accepts any token starting with "refresh-" prefix and returns a mock user ID.
func (s *MockTokenService) ValidateRefreshToken(token string) (domain.UserID, error) {
	if len(token) < 8 || token[:8] != "refresh-" {
		return "", fmt.Errorf("invalid refresh token format")
	}
	return domain.UserID("mock-user-id"), nil
}
