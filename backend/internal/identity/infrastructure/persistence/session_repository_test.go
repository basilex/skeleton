package persistence_test

import (
	"context"
	"net"
	"testing"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/persistence"
	"github.com/basilex/skeleton/pkg/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, pool *pgxpool.Pool) identityDomain.UserID {
	userRepo := persistence.NewUserRepository(pool)
	ctx := context.Background()

	email, _ := identityDomain.NewEmail("test@example.com")
	passwordHash, _ := identityDomain.NewPasswordHash("password123")
	user, err := identityDomain.NewUser(email, passwordHash)
	require.NoError(t, err)

	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	return user.ID()
}

func TestSessionRepository_Save(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")
	duration := time.Hour * 24

	session, err := identityDomain.NewSession(userID, device, ip, duration)
	require.NoError(t, err)

	t.Run("save new session", func(t *testing.T) {
		err := repo.Save(ctx, session)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, session.ID())
		require.NoError(t, err)
		require.Equal(t, session.ID(), found.ID())
		require.Equal(t, userID, found.UserID())
		require.Equal(t, identityDomain.SessionStatusActive, found.Status())
	})

	t.Run("update existing session", func(t *testing.T) {
		err := session.Refresh(time.Hour * 2)
		require.NoError(t, err)

		err = repo.Save(ctx, session)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, session.ID())
		require.NoError(t, err)
		require.True(t, found.ExpiresAt().After(time.Now().Add(time.Hour)))
	})
}

func TestSessionRepository_FindByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()

	t.Run("find existing session", func(t *testing.T) {
		userID := createTestUser(t, pool)
		device := identityDomain.NewDeviceInfo("Mozilla/5.0", "mobile", "iOS", "Safari", "iPhone")
		ip := net.ParseIP("10.0.0.1")
		session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
		_ = repo.Save(ctx, session)

		found, err := repo.FindByID(ctx, session.ID())
		require.NoError(t, err)
		require.Equal(t, session.ID(), found.ID())
		require.Equal(t, userID, found.UserID())
	})

	t.Run("find non-existing session", func(t *testing.T) {
		_, err := repo.FindByID(ctx, identityDomain.NewSessionID())
		require.Error(t, err)
	})
}

func TestSessionRepository_FindByUserID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	t.Run("find sessions by user", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook")
			ip := net.ParseIP("192.168.1.1")
			session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
			_ = repo.Save(ctx, session)
		}

		sessions, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, sessions, 3)
	})

	t.Run("find sessions for non-existing user", func(t *testing.T) {
		otherUserID := identityDomain.NewUserID()
		sessions, err := repo.FindByUserID(ctx, otherUserID)
		require.NoError(t, err)
		require.Len(t, sessions, 0)
	})
}

func TestSessionRepository_FindActiveByUserID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	t.Run("find active sessions", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook")
			ip := net.ParseIP("192.168.1.1")
			session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
			_ = repo.Save(ctx, session)
		}

		// Create and immediately expire a session
		expiredSession, _ := identityDomain.NewSession(userID,
			identityDomain.NewDeviceInfo("Mozilla/5.0", "mobile", "iOS", "Safari", "iPhone"),
			net.ParseIP("10.0.0.1"),
			time.Hour,
		)
		_ = repo.Save(ctx, expiredSession)
		expiredSession.Expire()
		_ = repo.Save(ctx, expiredSession)

		sessions, err := repo.FindActiveByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, sessions, 2)
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook")
	ip := net.ParseIP("192.168.1.1")
	session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
	_ = repo.Save(ctx, session)

	t.Run("delete session", func(t *testing.T) {
		err := repo.Delete(ctx, session.ID())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, session.ID())
		require.Error(t, err)
	})
}

func TestSessionRepository_DeleteByUserID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	for i := 0; i < 3; i++ {
		device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook")
		ip := net.ParseIP("192.168.1.1")
		session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
		_ = repo.Save(ctx, session)
	}

	t.Run("delete all sessions for user", func(t *testing.T) {
		err := repo.DeleteByUserID(ctx, userID)
		require.NoError(t, err)

		sessions, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, sessions, 0)
	})
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewSessionRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	for i := 0; i < 2; i++ {
		device := identityDomain.NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook")
		ip := net.ParseIP("192.168.1.1")
		session, _ := identityDomain.NewSession(userID, device, ip, time.Hour)
		_ = repo.Save(ctx, session)
	}

	expiredSession, _ := identityDomain.NewSession(userID,
		identityDomain.NewDeviceInfo("Mozilla/5.0", "mobile", "iOS", "Safari", "iPhone"),
		net.ParseIP("10.0.0.1"),
		time.Hour,
	)
	_ = repo.Save(ctx, expiredSession)
	expiredSession.Expire()
	_ = repo.Save(ctx, expiredSession)

	t.Run("delete expired sessions", func(t *testing.T) {
		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)

		sessions, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, sessions, 2)
	})
}
