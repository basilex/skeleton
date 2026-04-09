// Package command provides command handlers for modifying preferences state.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// UpdatePreferencesHandler handles commands to update user preferences.
type UpdatePreferencesHandler struct {
	preferences domain.PreferencesRepository
	bus         eventbus.Bus
}

// NewUpdatePreferencesHandler creates a new UpdatePreferencesHandler.
func NewUpdatePreferencesHandler(
	preferences domain.PreferencesRepository,
	bus eventbus.Bus,
) *UpdatePreferencesHandler {
	return &UpdatePreferencesHandler{
		preferences: preferences,
		bus:         bus,
	}
}

// UpdatePreferencesCommand represents a command to update user preferences.
type UpdatePreferencesCommand struct {
	UserID                string
	Theme                 *string
	Language              *string
	DateFormat            *string
	Timezone              *string
	EmailEnabled          *bool
	SMSEnabled            *bool
	PushEnabled           *bool
	InAppEnabled          *bool
	MarketingEmails       *bool
	WeeklyDigest          *bool
	QuietHoursStart       *int
	QuietHoursEnd         *int
	NotificationsTimezone *string
}

// Handle executes the UpdatePreferencesCommand.
func (h *UpdatePreferencesHandler) Handle(ctx context.Context, cmd UpdatePreferencesCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	prefs, err := h.preferences.FindByUserID(ctx, userID)
	if err != nil {
		prefs, err = domain.NewUserPreferences(userID)
		if err != nil {
			return fmt.Errorf("create preferences: %w", err)
		}
	}

	if cmd.Theme != nil {
		theme, err := domain.ParseTheme(*cmd.Theme)
		if err != nil {
			return fmt.Errorf("parse theme: %w", err)
		}
		_ = prefs.SetTheme(theme)
	}

	if cmd.Language != nil {
		lang, err := domain.ParseLanguage(*cmd.Language)
		if err != nil {
			return fmt.Errorf("parse language: %w", err)
		}
		_ = prefs.SetLanguage(lang)
	}

	if cmd.DateFormat != nil {
		format, err := domain.ParseDateFormat(*cmd.DateFormat)
		if err != nil {
			return fmt.Errorf("parse date format: %w", err)
		}
		_ = prefs.SetDateFormat(format)
	}

	if cmd.Timezone != nil {
		_ = prefs.SetTimezone(*cmd.Timezone)
	}

	if cmd.EmailEnabled != nil {
		prefs.SetEmailNotifications(*cmd.EmailEnabled)
	}

	if cmd.SMSEnabled != nil {
		prefs.SetSMSNotifications(*cmd.SMSEnabled)
	}

	if cmd.PushEnabled != nil {
		prefs.SetPushNotifications(*cmd.PushEnabled)
	}

	if cmd.InAppEnabled != nil {
		prefs.SetInAppNotifications(*cmd.InAppEnabled)
	}

	if cmd.MarketingEmails != nil {
		prefs.SetMarketingEmails(*cmd.MarketingEmails)
	}

	if cmd.WeeklyDigest != nil {
		prefs.SetWeeklyDigest(*cmd.WeeklyDigest)
	}

	if cmd.QuietHoursStart != nil && cmd.QuietHoursEnd != nil {
		_ = prefs.SetQuietHours(*cmd.QuietHoursStart, *cmd.QuietHoursEnd)
	} else if cmd.QuietHoursStart == nil && cmd.QuietHoursEnd == nil {
	}

	if cmd.NotificationsTimezone != nil {
		_ = prefs.SetNotificationsTimezone(*cmd.NotificationsTimezone)
	}

	if err := h.preferences.Save(ctx, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	return nil
}

// SetThemeHandler handles commands to set user theme preference.
type SetThemeHandler struct {
	preferences domain.PreferencesRepository
}

// NewSetThemeHandler creates a new SetThemeHandler.
func NewSetThemeHandler(preferences domain.PreferencesRepository) *SetThemeHandler {
	return &SetThemeHandler{
		preferences: preferences,
	}
}

// SetThemeCommand represents a command to set the theme.
type SetThemeCommand struct {
	UserID string
	Theme  string
}

// Handle executes the SetThemeCommand.
func (h *SetThemeHandler) Handle(ctx context.Context, cmd SetThemeCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	theme, err := domain.ParseTheme(cmd.Theme)
	if err != nil {
		return fmt.Errorf("parse theme: %w", err)
	}

	prefs, err := h.preferences.FindByUserID(ctx, userID)
	if err != nil {
		prefs, err = domain.NewUserPreferences(userID)
		if err != nil {
			return fmt.Errorf("create preferences: %w", err)
		}
	}

	if err := prefs.SetTheme(theme); err != nil {
		return fmt.Errorf("set theme: %w", err)
	}

	if err := h.preferences.Save(ctx, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	return nil
}

// SetLanguageHandler handles commands to set user language preference.
type SetLanguageHandler struct {
	preferences domain.PreferencesRepository
}

// NewSetLanguageHandler creates a new SetLanguageHandler.
func NewSetLanguageHandler(preferences domain.PreferencesRepository) *SetLanguageHandler {
	return &SetLanguageHandler{
		preferences: preferences,
	}
}

// SetLanguageCommand represents a command to set the language.
type SetLanguageCommand struct {
	UserID   string
	Language string
}

// Handle executes the SetLanguageCommand.
func (h *SetLanguageHandler) Handle(ctx context.Context, cmd SetLanguageCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	lang, err := domain.ParseLanguage(cmd.Language)
	if err != nil {
		return fmt.Errorf("parse language: %w", err)
	}

	prefs, err := h.preferences.FindByUserID(ctx, userID)
	if err != nil {
		prefs, err = domain.NewUserPreferences(userID)
		if err != nil {
			return fmt.Errorf("create preferences: %w", err)
		}
	}

	if err := prefs.SetLanguage(lang); err != nil {
		return fmt.Errorf("set language: %w", err)
	}

	if err := h.preferences.Save(ctx, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	return nil
}
