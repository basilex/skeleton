package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")

	user, err := NewUser(email, hash)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, email, user.Email())
	require.True(t, user.IsActive())
	require.Empty(t, user.Roles())
	require.False(t, user.CreatedAt().IsZero())
}

func TestUserAssignRole(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	roleID := RoleID("role-1")
	err := user.AssignRole(roleID)
	require.NoError(t, err)
	require.Len(t, user.Roles(), 1)
	require.Equal(t, roleID, user.Roles()[0])
}

func TestUserAssignDuplicateRole(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	roleID := RoleID("role-1")
	err := user.AssignRole(roleID)
	require.NoError(t, err)

	err = user.AssignRole(roleID)
	require.ErrorIs(t, err, ErrRoleAlreadyAssigned)
}

func TestUserRevokeRole(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	roleID := RoleID("role-1")
	_ = user.AssignRole(roleID)

	err := user.RevokeRole(roleID)
	require.NoError(t, err)
	require.Empty(t, user.Roles())
}

func TestUserRevokeRoleNotAssigned(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	err := user.RevokeRole(RoleID("role-1"))
	require.ErrorIs(t, err, ErrRoleNotAssigned)
}

func TestUserDeactivate(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	user.Deactivate()
	require.False(t, user.IsActive())
}

func TestUserDeactivateCannotAssignRole(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	user.Deactivate()

	err := user.AssignRole(RoleID("role-1"))
	require.ErrorIs(t, err, ErrUserInactive)
}

func TestUserDeactivateCannotRevokeRole(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	roleID := RoleID("role-1")
	_ = user.AssignRole(roleID)
	user.Deactivate()

	err := user.RevokeRole(roleID)
	require.ErrorIs(t, err, ErrUserInactive)
}

func TestUserCheckPassword(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	require.True(t, user.CheckPassword("Password1234!"))
	require.False(t, user.CheckPassword("WrongPassword"))
}

func TestUserPullEvents(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	events := user.PullEvents()
	require.Len(t, events, 1)
	_, ok := events[0].(UserRegistered)
	require.True(t, ok)

	events = user.PullEvents()
	require.Empty(t, events)
}

func TestUserHasPermission(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	perm, _ := NewPermission("users:read")
	role, _ := NewRole("admin", "", []Permission{perm})

	_ = user.AssignRole(role.ID())

	require.True(t, user.HasPermission(perm, []Role{*role}))

	otherPerm, _ := NewPermission("roles:manage")
	require.False(t, user.HasPermission(otherPerm, []Role{*role}))
}

func TestUserHasWildcardPermission(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	wildcard, _ := NewPermission("*:*")
	role, _ := NewRole("super_admin", "", []Permission{wildcard})

	_ = user.AssignRole(role.ID())

	require.True(t, user.HasPermission(Permission("users:read"), []Role{*role}))
	require.True(t, user.HasPermission(Permission("anything:whatever"), []Role{*role}))
}

func TestUserSetRoles(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	hash, _ := NewPasswordHash("Password1234!")
	user, _ := NewUser(email, hash)

	roles := []RoleID{RoleID("r1"), RoleID("r2")}
	user.SetRoles(roles)

	require.Len(t, user.Roles(), 2)
	require.Equal(t, roles, user.Roles())
}

func TestReconstituteUser(t *testing.T) {
	id := UserID("test-id")
	email, _ := NewEmail("test@example.com")
	hash := PasswordHash("hashed")
	now := time.Now().UTC()

	user, err := ReconstituteUser(id, email, hash, []RoleID{RoleID("r1")}, true, now, now)
	require.NoError(t, err)
	require.Equal(t, id, user.ID())
	require.Equal(t, email, user.Email())
	require.True(t, user.IsActive())
	require.Len(t, user.Roles(), 1)
	require.Empty(t, user.PullEvents())
}
