package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	notificationDomain "github.com/basilex/skeleton/internal/notifications/domain"
)

type UpdatePreferencesCommand struct {
	UserID          string
	Channel         notificationDomain.Channel
	Enabled         bool
	Frequency       notificationDomain.Frequency
	QuietHoursStart *int
	QuietHoursEnd   *int
	QuietHoursTZ    *string
}

type UpdatePreferencesHandler struct {
	preferencesRepo notificationDomain.PreferencesRepository
}

func NewUpdatePreferencesHandler(
	preferencesRepo notificationDomain.PreferencesRepository,
) *UpdatePreferencesHandler {
	return &UpdatePreferencesHandler{
		preferencesRepo: preferencesRepo,
	}
}

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
