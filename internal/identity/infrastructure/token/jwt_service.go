// Package token provides token generation and validation implementations.
// This package contains infrastructure implementations for JWT token handling
// and mock token services for testing purposes.
package token

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService implements token generation and validation using RSA-signed JWTs.
// It uses RS256 algorithm with separate public and private keys for enhanced security.
type JWTService struct {
	privateKey       *rsa.PrivateKey
	publicKey        *rsa.PublicKey
	accessTTLMinutes int
	refreshTTLDays   int
}

// NewJWTService creates a new JWT token service with the provided key files.
// The privateKeyPath and publicKeyPath should point to PEM-encoded RSA key files.
// accessTTLMinutes specifies the access token expiration time in minutes.
// refreshTTLDays specifies the refresh token expiration time in days (currently unused).
func NewJWTService(privateKeyPath, publicKeyPath string, accessTTLMinutes, refreshTTLDays int) (*JWTService, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &JWTService{
		privateKey:       privateKey,
		publicKey:        publicKey,
		accessTTLMinutes: accessTTLMinutes,
		refreshTTLDays:   refreshTTLDays,
	}, nil
}

// claims represents the custom JWT claims structure containing user identity
// and authorization information.
type claims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// GenerateAccessToken generates a new RSA-signed JWT access token for the specified user.
// The token includes the user ID, role names, and all permissions aggregated from roles.
func (s *JWTService) GenerateAccessToken(userID domain.UserID, roles []domain.Role) (string, error) {
	permissions := make([]string, 0)
	for _, role := range roles {
		for _, p := range role.Permissions() {
			permissions = append(permissions, p.String())
		}
	}
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name()
	}

	now := time.Now().UTC()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.accessTTLMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:      userID.String(),
		Roles:       roleNames,
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken generates a new refresh token using UUID v7.
// This token is used to obtain new access tokens without re-authentication.
func (s *JWTService) GenerateRefreshToken() (string, error) {
	return uuid.NewV7().String(), nil
}

// ValidateAccessToken validates an RSA-signed JWT access token and extracts claims.
// Returns domain.TokenClaims containing user ID, roles, and permissions.
func (s *JWTService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	userID, err := domain.ParseUserID(c.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id from token: %w", err)
	}

	return &domain.TokenClaims{
		UserID:      userID,
		Roles:       c.Roles,
		Permissions: c.Permissions,
	}, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user ID.
// Current implementation returns empty user ID and no error (placeholder implementation).
func (s *JWTService) ValidateRefreshToken(_ string) (domain.UserID, error) {
	return domain.UserID{}, nil
}
