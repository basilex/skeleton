package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

func setupTestKeys(t *testing.T) (privatePath, publicPath string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := &privateKey.PublicKey

	dir := t.TempDir()
	privatePath = filepath.Join(dir, "private.pem")
	publicPath = filepath.Join(dir, "public.pem")

	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})
	require.NoError(t, os.WriteFile(privatePath, privatePEM, 0600))

	publicBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})
	require.NoError(t, os.WriteFile(publicPath, publicPEM, 0644))

	return privatePath, publicPath
}

func TestJWTService_GenerateAndValidate(t *testing.T) {
	privatePath, publicPath := setupTestKeys(t)

	svc, err := NewJWTService(privatePath, publicPath, 15, 7)
	require.NoError(t, err)

	userID := domain.NewUserID()
	perm, _ := domain.NewPermission("users:read")
	role, _ := domain.NewRole("admin", "", []domain.Permission{perm})

	token, err := svc.GenerateAccessToken(userID, []domain.Role{*role})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, userID, claims.UserID)
	require.Equal(t, []string{"admin"}, claims.Roles)
	require.Equal(t, []string{"users:read"}, claims.Permissions)
}

func TestJWTService_InvalidToken(t *testing.T) {
	privatePath, publicPath := setupTestKeys(t)

	svc, err := NewJWTService(privatePath, publicPath, 15, 7)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken("invalid.token.here")
	require.Error(t, err)
}

func TestJWTService_WrongKey(t *testing.T) {
	privatePath1, publicPath1 := setupTestKeys(t)
	privatePath2, publicPath2 := setupTestKeys(t)

	svc1, err := NewJWTService(privatePath1, publicPath1, 15, 7)
	require.NoError(t, err)

	svc2, err := NewJWTService(privatePath2, publicPath2, 15, 7)
	require.NoError(t, err)

	userID := domain.NewUserID()
	role, _ := domain.NewRole("admin", "", nil)

	token, err := svc1.GenerateAccessToken(userID, []domain.Role{*role})
	require.NoError(t, err)

	_, err = svc2.ValidateAccessToken(token)
	require.Error(t, err)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	privatePath, publicPath := setupTestKeys(t)

	svc, err := NewJWTService(privatePath, publicPath, -1, 7)
	require.NoError(t, err)

	userID := domain.NewUserID()
	role, _ := domain.NewRole("admin", "", nil)

	token, err := svc.GenerateAccessToken(userID, []domain.Role{*role})
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(token)
	require.Error(t, err)
}

func TestJWTService_MultiplePermissions(t *testing.T) {
	privatePath, publicPath := setupTestKeys(t)

	svc, err := NewJWTService(privatePath, publicPath, 15, 7)
	require.NoError(t, err)

	userID := domain.NewUserID()
	p1, _ := domain.NewPermission("users:read")
	p2, _ := domain.NewPermission("users:write")
	p3, _ := domain.NewPermission("roles:manage")
	role1, _ := domain.NewRole("admin", "", []domain.Permission{p1, p2})
	role2, _ := domain.NewRole("manager", "", []domain.Permission{p3})

	token, err := svc.GenerateAccessToken(userID, []domain.Role{*role1, *role2})
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.Len(t, claims.Permissions, 3)
	require.Contains(t, claims.Permissions, "users:read")
	require.Contains(t, claims.Permissions, "users:write")
	require.Contains(t, claims.Permissions, "roles:manage")
}

func TestJWTService_RefreshToken(t *testing.T) {
	privatePath, publicPath := setupTestKeys(t)

	svc, err := NewJWTService(privatePath, publicPath, 15, 7)
	require.NoError(t, err)

	userID := domain.NewUserID()
	refreshToken, err := svc.GenerateRefreshToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	parsedUserID, err := svc.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	require.Equal(t, userID, parsedUserID)
}
