package persistence_test

import (
	"context"
	"testing"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/persistence"
	"github.com/basilex/skeleton/pkg/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func createTestUserForPrefs(t *testing.T, pool *pgxpool.Pool) identityDomain.UserID {
	userRepo := persistence.NewUserRepository(pool)
	ctx := context.Background()

	email, _ := identityDomain.NewEmail("test-prefs@example.com")
	passwordHash, _ := identityDomain.NewPasswordHash("password123")
	user, err := identityDomain.NewUser(email, passwordHash)
	require.NoError(t, err)

	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	return user.ID()
}

func TestPreferencesRepository_Save(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	t.Run("save new preferences", func(t *testing.T) {
		prefs, err := identityDomain.NewUserPreferences(userID)
		require.NoError(t, err)

		err = repo.Save(ctx, prefs)
		require.NoError(t, err)

		found, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userID, found.UserID())
		require.Equal(t, identityDomain.ThemeAuto, found.Theme())
		require.Equal(t, identityDomain.LanguageEn, found.Language())
	})

	t.Run("update existing preferences", func(t *testing.T) {
		prefs, _ := identityDomain.NewUserPreferences(userID)
		_ = repo.Save(ctx, prefs)

		_ = prefs.SetTheme(identityDomain.ThemeDark)
		_ = prefs.SetLanguage(identityDomain.LanguageUk)
		err := repo.Save(ctx, prefs)
		require.NoError(t, err)

		found, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, identityDomain.ThemeDark, found.Theme())
		require.Equal(t, identityDomain.LanguageUk, found.Language())
	})
}

func TestPreferencesRepository_FindByUserID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	t.Run("find existing preferences", func(t *testing.T) {
		prefs, _ := identityDomain.NewUserPreferences(userID)
		_ = repo.Save(ctx, prefs)

		found, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userID, found.UserID())
	})

	t.Run("find non-existing preferences", func(t *testing.T) {
		otherUserID := identityDomain.NewUserID()
		_, err := repo.FindByUserID(ctx, otherUserID)
		require.Error(t, err)
	})
}

func TestPreferencesRepository_FindByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	prefs, _ := identityDomain.NewUserPreferences(userID)
	_ = repo.Save(ctx, prefs)

	t.Run("find by id", func(t *testing.T) {
		found, err := repo.FindByID(ctx, prefs.ID())
		require.NoError(t, err)
		require.Equal(t, prefs.ID(), found.ID())
		require.Equal(t, userID, found.UserID())
	})

	t.Run("find non-existing id", func(t *testing.T) {
		_, err := repo.FindByID(ctx, identityDomain.NewPreferencesID())
		require.Error(t, err)
	})
}

func TestPreferencesRepository_Delete(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	prefs, _ := identityDomain.NewUserPreferences(userID)
	_ = repo.Save(ctx, prefs)

	t.Run("delete preferences", func(t *testing.T) {
		err := repo.Delete(ctx, prefs.ID())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, prefs.ID())
		require.Error(t, err)
	})
}

func TestPreferencesRepository_DeleteByUserID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	prefs, _ := identityDomain.NewUserPreferences(userID)
	_ = repo.Save(ctx, prefs)

	t.Run("delete by user id", func(t *testing.T) {
		err := repo.DeleteByUserID(ctx, userID)
		require.NoError(t, err)

		_, err = repo.FindByUserID(ctx, userID)
		require.Error(t, err)
	})
}

func TestPreferencesRepository_PreferencesWithQuietHours(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	t.Run("save and retrieve quiet hours", func(t *testing.T) {
		prefs, _ := identityDomain.NewUserPreferences(userID)

		// Set quiet hours
		err := prefs.SetQuietHours(22, 6)
		require.NoError(t, err)

		// Set theme and language
		_ = prefs.SetTheme(identityDomain.ThemeDark)
		_ = prefs.SetLanguage(identityDomain.LanguageUk)

		err = repo.Save(ctx, prefs)
		require.NoError(t, err)

		found, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)

		notifications := found.Notifications()
		require.NotNil(t, notifications.QuietHoursStart())
		require.NotNil(t, notifications.QuietHoursEnd())
		require.Equal(t, 22, *notifications.QuietHoursStart())
		require.Equal(t, 6, *notifications.QuietHoursEnd())
		require.Equal(t, identityDomain.ThemeDark, found.Theme())
		require.Equal(t, identityDomain.LanguageUk, found.Language())
	})
}

func TestPreferencesRepository_UpdateNotificationSettings(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := persistence.NewPreferencesRepository(pool)
	ctx := context.Background()
	userID := createTestUserForPrefs(t, pool)

	t.Run("update notification settings", func(t *testing.T) {
		prefs, _ := identityDomain.NewUserPreferences(userID)

		// Update notification settings
		prefs.SetEmailNotifications(false)
		prefs.SetSMSNotifications(true)
		prefs.SetPushNotifications(false)
		prefs.SetMarketingEmails(true)
		prefs.SetWeeklyDigest(false)

		err := repo.Save(ctx, prefs)
		require.NoError(t, err)

		found, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)

		notifications := found.Notifications()
		require.False(t, notifications.EmailEnabled())
		require.True(t, notifications.SMSEnabled())
		require.False(t, notifications.PushEnabled())
		require.True(t, notifications.MarketingEmails())
		require.False(t, notifications.WeeklyDigest())
	})
}
