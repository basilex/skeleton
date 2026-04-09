package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewUserPreferences(t *testing.T) {
	userID := NewUserID()

	prefs, err := NewUserPreferences(userID)
	require.NoError(t, err)
	require.NotNil(t, prefs)
	require.NotEmpty(t, prefs.ID())
	require.Equal(t, userID, prefs.UserID())
	require.Equal(t, ThemeAuto, prefs.Theme())
	require.Equal(t, LanguageEn, prefs.Language())
	require.Equal(t, DateFormatYMD, prefs.DateFormat())
	require.Equal(t, "UTC", prefs.Timezone())
	require.True(t, prefs.Notifications().EmailEnabled())
	require.True(t, prefs.Notifications().PushEnabled())
	require.True(t, prefs.Notifications().InAppEnabled())
	require.False(t, prefs.Notifications().SMSEnabled())
}

func TestUserPreferences_SetTheme(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	tests := []struct {
		theme   Theme
		wantErr bool
	}{
		{ThemeLight, false},
		{ThemeDark, false},
		{ThemeAuto, false},
		{Theme("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.theme), func(t *testing.T) {
			err := prefs.SetTheme(tt.theme)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.theme, prefs.Theme())
			}
		})
	}
}

func TestUserPreferences_SetLanguage(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	tests := []struct {
		lang    Language
		wantErr bool
	}{
		{LanguageEn, false},
		{LanguageUk, false},
		{LanguageDe, false},
		{LanguageFr, false},
		{LanguageEs, false},
		{Language("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			err := prefs.SetLanguage(tt.lang)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.lang, prefs.Language())
			}
		})
	}
}

func TestUserPreferences_SetDateFormat(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	tests := []struct {
		format  DateFormat
		wantErr bool
	}{
		{DateFormatDMY, false},
		{DateFormatMDY, false},
		{DateFormatYMD, false},
		{DateFormat("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			err := prefs.SetDateFormat(tt.format)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.format, prefs.DateFormat())
			}
		})
	}
}

func TestUserPreferences_SetTimezone(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	t.Run("valid timezone", func(t *testing.T) {
		err := prefs.SetTimezone("Europe/Kyiv")
		require.NoError(t, err)
		require.Equal(t, "Europe/Kyiv", prefs.Timezone())
	})

	t.Run("empty timezone", func(t *testing.T) {
		err := prefs.SetTimezone("")
		require.Error(t, err)
	})
}

func TestUserPreferences_NotificationSettings(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	t.Run("set email notifications", func(t *testing.T) {
		prefs.SetEmailNotifications(false)
		require.False(t, prefs.Notifications().EmailEnabled())

		prefs.SetEmailNotifications(true)
		require.True(t, prefs.Notifications().EmailEnabled())
	})

	t.Run("set sms notifications", func(t *testing.T) {
		prefs.SetSMSNotifications(true)
		require.True(t, prefs.Notifications().SMSEnabled())

		prefs.SetSMSNotifications(false)
		require.False(t, prefs.Notifications().SMSEnabled())
	})

	t.Run("set push notifications", func(t *testing.T) {
		prefs.SetPushNotifications(false)
		require.False(t, prefs.Notifications().PushEnabled())

		prefs.SetPushNotifications(true)
		require.True(t, prefs.Notifications().PushEnabled())
	})

	t.Run("set marketing emails", func(t *testing.T) {
		prefs.SetMarketingEmails(true)
		require.True(t, prefs.Notifications().MarketingEmails())

		prefs.SetMarketingEmails(false)
		require.False(t, prefs.Notifications().MarketingEmails())
	})

	t.Run("set weekly digest", func(t *testing.T) {
		prefs.SetWeeklyDigest(false)
		require.False(t, prefs.Notifications().WeeklyDigest())

		prefs.SetWeeklyDigest(true)
		require.True(t, prefs.Notifications().WeeklyDigest())
	})
}

func TestUserPreferences_QuietHours(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	t.Run("set quiet hours", func(t *testing.T) {
		err := prefs.SetQuietHours(22, 6)
		require.NoError(t, err)
		require.NotNil(t, prefs.Notifications().QuietHoursStart())
		require.NotNil(t, prefs.Notifications().QuietHoursEnd())
		require.Equal(t, 22, *prefs.Notifications().QuietHoursStart())
		require.Equal(t, 6, *prefs.Notifications().QuietHoursEnd())
	})

	t.Run("invalid start hour", func(t *testing.T) {
		err := prefs.SetQuietHours(24, 6)
		require.Error(t, err)
	})

	t.Run("invalid end hour", func(t *testing.T) {
		err := prefs.SetQuietHours(22, 25)
		require.Error(t, err)
	})

	t.Run("clear quiet hours", func(t *testing.T) {
		prefs.SetQuietHours(22, 6)
		prefs.ClearQuietHours()
		require.Nil(t, prefs.Notifications().QuietHoursStart())
		require.Nil(t, prefs.Notifications().QuietHoursEnd())
	})
}

func TestUserPreferences_SetNotificationsTimezone(t *testing.T) {
	userID := NewUserID()
	prefs, _ := NewUserPreferences(userID)

	t.Run("set valid timezone", func(t *testing.T) {
		err := prefs.SetNotificationsTimezone("America/New_York")
		require.NoError(t, err)
		require.Equal(t, "America/New_York", prefs.Notifications().Timezone())
	})

	t.Run("set empty timezone", func(t *testing.T) {
		err := prefs.SetNotificationsTimezone("")
		require.Error(t, err)
	})
}

func TestReconstituteUserPreferences(t *testing.T) {
	id := NewPreferencesID()
	userID := NewUserID()
	notifications := NewNotificationSettings()
	notifications.emailEnabled = false
	notifications.smsEnabled = true
	now := time.Now().UTC()

	prefs := ReconstituteUserPreferences(
		id,
		userID,
		ThemeDark,
		LanguageUk,
		DateFormatDMY,
		"Europe/Kyiv",
		notifications,
		now,
		now,
	)

	require.Equal(t, id, prefs.ID())
	require.Equal(t, userID, prefs.UserID())
	require.Equal(t, ThemeDark, prefs.Theme())
	require.Equal(t, LanguageUk, prefs.Language())
	require.Equal(t, DateFormatDMY, prefs.DateFormat())
	require.Equal(t, "Europe/Kyiv", prefs.Timezone())
	require.False(t, prefs.Notifications().EmailEnabled())
	require.True(t, prefs.Notifications().SMSEnabled())
}

func TestParseTheme(t *testing.T) {
	tests := []struct {
		input   string
		want    Theme
		wantErr bool
	}{
		{"light", ThemeLight, false},
		{"dark", ThemeDark, false},
		{"auto", ThemeAuto, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTheme(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input   string
		want    Language
		wantErr bool
	}{
		{"en", LanguageEn, false},
		{"uk", LanguageUk, false},
		{"de", LanguageDe, false},
		{"fr", LanguageFr, false},
		{"es", LanguageEs, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLanguage(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseDateFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    DateFormat
		wantErr bool
	}{
		{"dd/mm/yyyy", DateFormatDMY, false},
		{"mm/dd/yyyy", DateFormatMDY, false},
		{"yyyy-mm-dd", DateFormatYMD, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateFormat(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
