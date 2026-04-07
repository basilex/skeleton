package token

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

type MockTokenService struct{}

func (s *MockTokenService) GenerateAccessToken(userID domain.UserID, roles []domain.Role) (string, error) {
	permissions := make([]string, 0)
	for _, role := range roles {
		for _, p := range role.Permissions() {
			permissions = append(permissions, p.String())
		}
	}
	return fmt.Sprintf("mock-access-%s-%d", string(userID), time.Now().Unix()), nil
}

func (s *MockTokenService) GenerateRefreshToken() (string, error) {
	return uuid.NewV7().String(), nil
}

func (s *MockTokenService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	if len(tokenString) < 12 || tokenString[:12] != "mock-access-" {
		return nil, fmt.Errorf("invalid mock token")
	}
	return &domain.TokenClaims{
		UserID:      domain.UserID("mock-user-id"),
		Roles:       []string{"admin"},
		Permissions: []string{"*:*"},
	}, nil
}

func (s *MockTokenService) ValidateRefreshToken(token string) (domain.UserID, error) {
	return domain.UserID("mock-user-id"), nil
}
