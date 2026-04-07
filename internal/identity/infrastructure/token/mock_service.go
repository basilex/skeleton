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
	return fmt.Sprintf("access-%s-%d", string(userID), time.Now().Unix()), nil
}

func (s *MockTokenService) GenerateRefreshToken() (string, error) {
	return "refresh-" + uuid.NewV7().String(), nil
}

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

func (s *MockTokenService) ValidateRefreshToken(token string) (domain.UserID, error) {
	if len(token) < 8 || token[:8] != "refresh-" {
		return "", fmt.Errorf("invalid refresh token format")
	}
	return domain.UserID("mock-user-id"), nil
}
