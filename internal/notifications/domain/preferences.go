// Package domain provides domain entities and value objects for the notifications module.
// This package contains the core business logic types for notification management,
// including preferences, templates, and domain events.
package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

// PreferencesID is a unique identifier for notification preferences.
type PreferencesID string

// NewPreferencesID generates a new unique PreferencesID using UUID v7.
func NewPreferencesID() PreferencesID {
	return PreferencesID(uuid.NewV7().String())
}

// ParsePreferencesID validates and converts a string to PreferencesID.
func ParsePreferencesID(s string) (PreferencesID, error) {
	if s == "" {
		return "", fmt.Errorf("preferences ID cannot be empty")
	}
	return PreferencesID(s), nil
}

// String returns the string representation of PreferencesID.
func (id PreferencesID) String() string {
	return string(id)
}

// Frequency represents how often notifications should be delivered.
type Frequency string

// Frequency type constants.
const (
	FrequencyImmediate Frequency = "immediate"
	FrequencyDaily     Frequency = "daily"
	FrequencyWeekly    Frequency = "weekly"
)

// String returns the string representation of the frequency.
func (f Frequency) String() string {
	return string(f)
}

// ParseFrequency converts a string to a Frequency value.
func ParseFrequency(s string) (Frequency, error) {
	switch s {
	case string(FrequencyImmediate):
		return FrequencyImmediate, nil
	case string(FrequencyDaily):
		return FrequencyDaily, nil
	case string(FrequencyWeekly):
		return FrequencyWeekly, nil
	default:
		return "", fmt.Errorf("invalid frequency: %s", s)
	}
}

// QuietHours defines a period during which notifications should not be sent.
type QuietHours struct {
	startHour int
	endHour   int
	timezone  string
}

// NewQuietHours creates a QuietHours configuration with validation.
// Hours must be between 0 and 23. If timezone is empty, UTC is used.
func NewQuietHours(startHour, endHour int, timezone string) (*QuietHours, error) {
	if startHour < 0 || startHour > 23 {
		return nil, fmt.Errorf("start hour must be between 0 and 23")
	}
	if endHour < 0 || endHour > 23 {
		return nil, fmt.Errorf("end hour must be between 0 and 23")
	}
	if timezone == "" {
		timezone = "UTC"
	}
	return &QuietHours{
		startHour: startHour,
		endHour:   endHour,
		timezone:  timezone,
	}, nil
}

// StartHour returns the hour when quiet hours begin.
func (q *QuietHours) StartHour() int {
	return q.startHour
}

// EndHour returns the hour when quiet hours end.
func (q *QuietHours) EndHour() int {
	return q.endHour
}

// Timezone returns the timezone for the quiet hours configuration.
func (q *QuietHours) Timezone() string {
	return q.timezone
}

// IsQuietHour checks if the given time falls within the quiet hours period.
func (q *QuietHours) IsQuietHour(t time.Time) bool {
	hour := t.Hour()
	if q.startHour <= q.endHour {
		return hour >= q.startHour && hour < q.endHour
	}
	return hour >= q.startHour || hour < q.endHour
}

// ChannelPreference represents notification preferences for a specific channel.
type ChannelPreference struct {
	enabled    bool
	frequency  Frequency
	quietHours *QuietHours
}

// NewChannelPreference creates a new ChannelPreference with the specified settings.
func NewChannelPreference(enabled bool, frequency Frequency, quietHours *QuietHours) ChannelPreference {
	return ChannelPreference{
		enabled:    enabled,
		frequency:  frequency,
		quietHours: quietHours,
	}
}

// Enabled returns whether the channel is enabled.
func (c *ChannelPreference) Enabled() bool {
	return c.enabled
}

// Frequency returns the notification frequency for this channel.
func (c *ChannelPreference) Frequency() Frequency {
	return c.frequency
}

// QuietHours returns the quiet hours configuration for this channel.
func (c *ChannelPreference) QuietHours() *QuietHours {
	return c.quietHours
}

// Enable activates the notification channel.
func (c *ChannelPreference) Enable() {
	c.enabled = true
}

// Disable deactivates the notification channel.
func (c *ChannelPreference) Disable() {
	c.enabled = false
}

// SetFrequency updates the notification frequency for this channel.
func (c *ChannelPreference) SetFrequency(f Frequency) {
	c.frequency = f
}

// SetQuietHours updates the quiet hours configuration for this channel.
func (c *ChannelPreference) SetQuietHours(qh *QuietHours) {
	c.quietHours = qh
}

// NotificationPreferences represents a user's notification preferences across all channels.
type NotificationPreferences struct {
	id        PreferencesID
	userID    domain.UserID
	channels  map[Channel]ChannelPreference
	createdAt time.Time
	updatedAt time.Time
}

// NewNotificationPreferences creates default notification preferences for a user.
// By default, email and in-app notifications are enabled, while SMS and push are disabled.
func NewNotificationPreferences(userID domain.UserID) *NotificationPreferences {
	now := time.Now()
	return &NotificationPreferences{
		id:     NewPreferencesID(),
		userID: userID,
		channels: map[Channel]ChannelPreference{
			ChannelEmail: NewChannelPreference(true, FrequencyImmediate, nil),
			ChannelSMS:   NewChannelPreference(false, FrequencyImmediate, nil),
			ChannelPush:  NewChannelPreference(false, FrequencyImmediate, nil),
			ChannelInApp: NewChannelPreference(true, FrequencyImmediate, nil),
		},
		createdAt: now,
		updatedAt: now,
	}
}

// ID returns the preferences' unique identifier.
func (p *NotificationPreferences) ID() PreferencesID {
	return p.id
}

// UserID returns the user's unique identifier.
func (p *NotificationPreferences) UserID() domain.UserID {
	return p.userID
}

// Channels returns the map of channel preferences.
func (p *NotificationPreferences) Channels() map[Channel]ChannelPreference {
	return p.channels
}

// CreatedAt returns when the preferences were created.
func (p *NotificationPreferences) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns when the preferences were last updated.
func (p *NotificationPreferences) UpdatedAt() time.Time {
	return p.updatedAt
}

// GetChannelPreference returns the preference for a specific channel.
func (p *NotificationPreferences) GetChannelPreference(channel Channel) (ChannelPreference, bool) {
	pref, ok := p.channels[channel]
	return pref, ok
}

// EnableChannel enables notifications for a specific channel.
func (p *NotificationPreferences) EnableChannel(channel Channel) {
	if pref, ok := p.channels[channel]; ok {
		pref.Enable()
		p.channels[channel] = pref
	} else {
		p.channels[channel] = NewChannelPreference(true, FrequencyImmediate, nil)
	}
	p.updatedAt = time.Now()
}

// DisableChannel disables notifications for a specific channel.
func (p *NotificationPreferences) DisableChannel(channel Channel) {
	if pref, ok := p.channels[channel]; ok {
		pref.Disable()
		p.channels[channel] = pref
	}
	p.updatedAt = time.Now()
}

// SetChannelFrequency sets the notification frequency for a specific channel.
func (p *NotificationPreferences) SetChannelFrequency(channel Channel, frequency Frequency) {
	if pref, ok := p.channels[channel]; ok {
		pref.SetFrequency(frequency)
		p.channels[channel] = pref
	} else {
		p.channels[channel] = NewChannelPreference(true, frequency, nil)
	}
	p.updatedAt = time.Now()
}

// SetChannelQuietHours sets the quiet hours configuration for a specific channel.
func (p *NotificationPreferences) SetChannelQuietHours(channel Channel, quietHours *QuietHours) {
	if pref, ok := p.channels[channel]; ok {
		pref.SetQuietHours(quietHours)
		p.channels[channel] = pref
	}
	p.updatedAt = time.Now()
}

// IsChannelEnabled returns whether notifications are enabled for a specific channel.
func (p *NotificationPreferences) IsChannelEnabled(channel Channel) bool {
	pref, ok := p.channels[channel]
	if !ok {
		return false
	}
	return pref.Enabled()
}

// CanSendNow checks if a notification can be sent now based on channel preferences and quiet hours.
func (p *NotificationPreferences) CanSendNow(channel Channel, t time.Time) bool {
	pref, ok := p.channels[channel]
	if !ok || !pref.Enabled() {
		return false
	}

	if pref.Frequency() != FrequencyImmediate {
		return false
	}

	if qh := pref.QuietHours(); qh != nil {
		loc, err := time.LoadLocation(qh.Timezone())
		if err != nil {
			loc = time.UTC
		}
		localTime := t.In(loc)
		return !qh.IsQuietHour(localTime)
	}

	return true
}
