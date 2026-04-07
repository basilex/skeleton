// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	notificationDomain "github.com/basilex/skeleton/internal/notifications/domain"
)

// UpdatePreferencesCommand represents a command to update user notification preferences.
// It allows enabling/disabling channels, setting frequency, and configuring quiet hours.
type UpdatePreferencesCommand struct {
	UserID          string
	Channel         notificationDomain.Channel
	Enabled         bool
	Frequency       notificationDomain.Frequency
	QuietHoursStart *int
	QuietHoursEnd   *int
	QuietHoursTZ    *string
}

// UpdatePreferencesHandler handles commands to update notification preferences.
// It retrieves existing preferences, applies updates, and persists the changes.
type UpdatePreferencesHandler struct {
	preferencesRepo notificationDomain.PreferencesRepository
}

// NewUpdatePreferencesHandler creates a new UpdatePreferencesHandler with the required repository.
func NewUpdatePreferencesHandler(
	preferencesRepo notificationDomain.PreferencesRepository,
) *UpdatePreferencesHandler {
	return &UpdatePreferencesHandler{
		preferencesRepo: preferencesRepo,
	}
}

// Handle executes the UpdatePreferencesCommand to update user notification settings.
// It creates new preferences if none exist, applies channel settings, and handles quiet hours configuration.
func (h *UpdatePreferencesHandler) Handle(ctx context.Context, cmd UpdatePreferencesCommand) error {
	prefs, err := h.preferencesRepo.GetByUserID(ctx, cmd.UserID)
	if err != nil || prefs == nil {
		prefs = notificationDomain.NewNotificationPreferences(domain.UserID(cmd.UserID))
	}

	if cmd.Enabled {
		prefs.EnableChannel(cmd.Channel)
	} else {
		prefs.DisableChannel(cmd.Channel)
	}

	if cmd.Frequency != "" {
		prefs.SetChannelFrequency(cmd.Channel, cmd.Frequency)
	}

	if cmd.QuietHoursStart != nil && cmd.QuietHoursEnd != nil {
		start := *cmd.QuietHoursStart
		end := *cmd.QuietHoursEnd
		tz := "UTC"
		if cmd.QuietHoursTZ != nil {
			tz = *cmd.QuietHoursTZ
		}

		quietHours, err := notificationDomain.NewQuietHours(start, end, tz)
		if err != nil {
			return fmt.Errorf("invalid quiet hours: %w", err)
		}
		prefs.SetChannelQuietHours(cmd.Channel, quietHours)
	}

	if err := h.preferencesRepo.Upsert(ctx, prefs); err != nil {
		return fmt.Errorf("upsert preferences: %w", err)
	}

	return nil
}
