// Package domain provides domain entities and repository interfaces for the identity module.
package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// PreferencesID is a unique identifier for user preferences.
type PreferencesID string

// NewPreferencesID generates a new unique PreferencesID.
func NewPreferencesID() PreferencesID {
	return PreferencesID(uuid.NewV7().String())
}

// ParsePreferencesID validates and converts a string to PreferencesID.
func ParsePreferencesID(s string) (PreferencesID, error) {
	if s == "" {
		return "", fmt.Errorf("preferences id cannot be empty")
	}
	return PreferencesID(s), nil
}

// String returns the string representation of PreferencesID.
func (id PreferencesID) String() string {
	return string(id)
}

// Theme represents the user interface theme preference.
type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeAuto  Theme = "auto"
)

// String returns the string representation of Theme.
func (t Theme) String() string {
	return string(t)
}

// ParseTheme converts a string to a Theme value.
func ParseTheme(s string) (Theme, error) {
	switch s {
	case string(ThemeLight):
		return ThemeLight, nil
	case string(ThemeDark):
		return ThemeDark, nil
	case string(ThemeAuto):
		return ThemeAuto, nil
	default:
		return "", fmt.Errorf("invalid theme: %s", s)
	}
}

// Language represents a language preference.
type Language string

const (
	LanguageEn Language = "en"
	LanguageUk Language = "uk"
	LanguageDe Language = "de"
	LanguageFr Language = "fr"
	LanguageEs Language = "es"
)

// String returns the string representation of Language.
func (l Language) String() string {
	return string(l)
}

// ParseLanguage converts a string to a Language value.
func ParseLanguage(s string) (Language, error) {
	switch s {
	case string(LanguageEn):
		return LanguageEn, nil
	case string(LanguageUk):
		return LanguageUk, nil
	case string(LanguageDe):
		return LanguageDe, nil
	case string(LanguageFr):
		return LanguageFr, nil
	case string(LanguageEs):
		return LanguageEs, nil
	default:
		return "", fmt.Errorf("invalid language: %s", s)
	}
}

// DateFormat represents the date format preference.
type DateFormat string

const (
	DateFormatDMY DateFormat = "dd/mm/yyyy"
	DateFormatMDY DateFormat = "mm/dd/yyyy"
	DateFormatYMD DateFormat = "yyyy-mm-dd"
)

// String returns the string representation of DateFormat.
func (d DateFormat) String() string {
	return string(d)
}

// ParseDateFormat converts a string to a DateFormat value.
func ParseDateFormat(s string) (DateFormat, error) {
	switch s {
	case string(DateFormatDMY):
		return DateFormatDMY, nil
	case string(DateFormatMDY):
		return DateFormatMDY, nil
	case string(DateFormatYMD):
		return DateFormatYMD, nil
	default:
		return "", fmt.Errorf("invalid date format: %s", s)
	}
}

// NotificationSettings contains user notification preferences.
type NotificationSettings struct {
	emailEnabled    bool
	smsEnabled      bool
	pushEnabled     bool
	inAppEnabled    bool
	marketingEmails bool
	weeklyDigest    bool
	quietHoursStart *int
	quietHoursEnd   *int
	timezone        string
}

// NewNotificationSettings creates default notification settings.
func NewNotificationSettings() NotificationSettings {
	return NotificationSettings{
		emailEnabled:    true,
		smsEnabled:      false,
		pushEnabled:     true,
		inAppEnabled:    true,
		marketingEmails: false,
		weeklyDigest:    true,
		timezone:        "UTC",
	}
}

// EmailEnabled returns whether email notifications are enabled.
func (n NotificationSettings) EmailEnabled() bool {
	return n.emailEnabled
}

// SMSEnabled returns whether SMS notifications are enabled.
func (n NotificationSettings) SMSEnabled() bool {
	return n.smsEnabled
}

// PushEnabled returns whether push notifications are enabled.
func (n NotificationSettings) PushEnabled() bool {
	return n.pushEnabled
}

// InAppEnabled returns whether in-app notifications are enabled.
func (n NotificationSettings) InAppEnabled() bool {
	return n.inAppEnabled
}

// MarketingEmails returns whether marketing emails are enabled.
func (n NotificationSettings) MarketingEmails() bool {
	return n.marketingEmails
}

// WeeklyDigest returns whether weekly digest emails are enabled.
func (n NotificationSettings) WeeklyDigest() bool {
	return n.weeklyDigest
}

// QuietHoursStart returns the quiet hours start hour (0-23).
func (n NotificationSettings) QuietHoursStart() *int {
	return n.quietHoursStart
}

// QuietHoursEnd returns the quiet hours end hour (0-23).
func (n NotificationSettings) QuietHoursEnd() *int {
	return n.quietHoursEnd
}

// Timezone returns the user's timezone.
func (n NotificationSettings) Timezone() string {
	return n.timezone
}

// UserPreferences represents user preference settings.
type UserPreferences struct {
	id            PreferencesID
	userID        UserID
	theme         Theme
	language      Language
	dateFormat    DateFormat
	timezone      string
	notifications NotificationSettings
	createdAt     time.Time
	updatedAt     time.Time
}

// NewUserPreferences creates default preferences for a user.
func NewUserPreferences(userID UserID) (*UserPreferences, error) {
	now := time.Now().UTC()
	return &UserPreferences{
		id:            NewPreferencesID(),
		userID:        userID,
		theme:         ThemeAuto,
		language:      LanguageEn,
		dateFormat:    DateFormatYMD,
		timezone:      "UTC",
		notifications: NewNotificationSettings(),
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ReconstituteUserPreferences reconstructs UserPreferences from persisted state.
func ReconstituteUserPreferences(
	id PreferencesID,
	userID UserID,
	theme Theme,
	language Language,
	dateFormat DateFormat,
	timezone string,
	notifications NotificationSettings,
	createdAt time.Time,
	updatedAt time.Time,
) *UserPreferences {
	return &UserPreferences{
		id:            id,
		userID:        userID,
		theme:         theme,
		language:      language,
		dateFormat:    dateFormat,
		timezone:      timezone,
		notifications: notifications,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// ID returns the preferences ID.
func (p *UserPreferences) ID() PreferencesID {
	return p.id
}

// UserID returns the user ID.
func (p *UserPreferences) UserID() UserID {
	return p.userID
}

// Theme returns the theme preference.
func (p *UserPreferences) Theme() Theme {
	return p.theme
}

// Language returns the language preference.
func (p *UserPreferences) Language() Language {
	return p.language
}

// DateFormat returns the date format preference.
func (p *UserPreferences) DateFormat() DateFormat {
	return p.dateFormat
}

// Timezone returns the timezone preference.
func (p *UserPreferences) Timezone() string {
	return p.timezone
}

// Notifications returns the notification settings.
func (p *UserPreferences) Notifications() NotificationSettings {
	return p.notifications
}

// CreatedAt returns the creation time.
func (p *UserPreferences) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update time.
func (p *UserPreferences) UpdatedAt() time.Time {
	return p.updatedAt
}

// SetTheme updates the theme preference.
func (p *UserPreferences) SetTheme(theme Theme) error {
	if theme != ThemeLight && theme != ThemeDark && theme != ThemeAuto {
		return fmt.Errorf("invalid theme: %s", theme)
	}
	p.theme = theme
	p.updatedAt = time.Now().UTC()
	return nil
}

// SetLanguage updates the language preference.
func (p *UserPreferences) SetLanguage(language Language) error {
	if language != LanguageEn && language != LanguageUk &&
		language != LanguageDe && language != LanguageFr &&
		language != LanguageEs {
		return fmt.Errorf("invalid language: %s", language)
	}
	p.language = language
	p.updatedAt = time.Now().UTC()
	return nil
}

// SetDateFormat updates the date format preference.
func (p *UserPreferences) SetDateFormat(format DateFormat) error {
	if format != DateFormatDMY && format != DateFormatMDY && format != DateFormatYMD {
		return fmt.Errorf("invalid date format: %s", format)
	}
	p.dateFormat = format
	p.updatedAt = time.Now().UTC()
	return nil
}

// SetTimezone updates the timezone preference.
func (p *UserPreferences) SetTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	p.timezone = timezone
	p.updatedAt = time.Now().UTC()
	return nil
}

// SetEmailNotifications enables or disables email notifications.
func (p *UserPreferences) SetEmailNotifications(enabled bool) {
	p.notifications.emailEnabled = enabled
	p.updatedAt = time.Now().UTC()
}

// SetSMSNotifications enables or disables SMS notifications.
func (p *UserPreferences) SetSMSNotifications(enabled bool) {
	p.notifications.smsEnabled = enabled
	p.updatedAt = time.Now().UTC()
}

// SetPushNotifications enables or disables push notifications.
func (p *UserPreferences) SetPushNotifications(enabled bool) {
	p.notifications.pushEnabled = enabled
	p.updatedAt = time.Now().UTC()
}

// SetInAppNotifications enables or disables in-app notifications.
func (p *UserPreferences) SetInAppNotifications(enabled bool) {
	p.notifications.inAppEnabled = enabled
	p.updatedAt = time.Now().UTC()
}

// SetMarketingEmails enables or disables marketing emails.
func (p *UserPreferences) SetMarketingEmails(enabled bool) {
	p.notifications.marketingEmails = enabled
	p.updatedAt = time.Now().UTC()
}

// SetWeeklyDigest enables or disables weekly digest.
func (p *UserPreferences) SetWeeklyDigest(enabled bool) {
	p.notifications.weeklyDigest = enabled
	p.updatedAt = time.Now().UTC()
}

// SetQuietHours sets the quiet hours period.
// Hours must be between 0 and 23.
func (p *UserPreferences) SetQuietHours(startHour, endHour int) error {
	if startHour < 0 || startHour > 23 {
		return fmt.Errorf("start hour must be between 0 and 23")
	}
	if endHour < 0 || endHour > 23 {
		return fmt.Errorf("end hour must be between 0 and 23")
	}
	p.notifications.quietHoursStart = &startHour
	p.notifications.quietHoursEnd = &endHour
	p.updatedAt = time.Now().UTC()
	return nil
}

// ClearQuietHours removes the quiet hours setting.
func (p *UserPreferences) ClearQuietHours() {
	p.notifications.quietHoursStart = nil
	p.notifications.quietHoursEnd = nil
	p.updatedAt = time.Now().UTC()
}

// SetNotificationsTimezone sets the timezone for notifications.
func (p *UserPreferences) SetNotificationsTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	p.notifications.timezone = timezone
	p.updatedAt = time.Now().UTC()
	return nil
}
