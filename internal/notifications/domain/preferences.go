package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

type PreferencesID string

func NewPreferencesID() PreferencesID {
	return PreferencesID(uuid.NewV7().String())
}

func ParsePreferencesID(s string) (PreferencesID, error) {
	if s == "" {
		return "", fmt.Errorf("preferences ID cannot be empty")
	}
	return PreferencesID(s), nil
}

func (id PreferencesID) String() string {
	return string(id)
}

type Frequency string

const (
	FrequencyImmediate Frequency = "immediate"
	FrequencyDaily     Frequency = "daily"
	FrequencyWeekly    Frequency = "weekly"
)

func (f Frequency) String() string {
	return string(f)
}

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

type QuietHours struct {
	startHour int
	endHour   int
	timezone  string
}

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

func (q *QuietHours) StartHour() int {
	return q.startHour
}

func (q *QuietHours) EndHour() int {
	return q.endHour
}

func (q *QuietHours) Timezone() string {
	return q.timezone
}

func (q *QuietHours) IsQuietHour(t time.Time) bool {
	hour := t.Hour()
	if q.startHour <= q.endHour {
		return hour >= q.startHour && hour < q.endHour
	}
	return hour >= q.startHour || hour < q.endHour
}

type ChannelPreference struct {
	enabled    bool
	frequency  Frequency
	quietHours *QuietHours
}

func NewChannelPreference(enabled bool, frequency Frequency, quietHours *QuietHours) ChannelPreference {
	return ChannelPreference{
		enabled:    enabled,
		frequency:  frequency,
		quietHours: quietHours,
	}
}

func (c *ChannelPreference) Enabled() bool {
	return c.enabled
}

func (c *ChannelPreference) Frequency() Frequency {
	return c.frequency
}

func (c *ChannelPreference) QuietHours() *QuietHours {
	return c.quietHours
}

func (c *ChannelPreference) Enable() {
	c.enabled = true
}

func (c *ChannelPreference) Disable() {
	c.enabled = false
}

func (c *ChannelPreference) SetFrequency(f Frequency) {
	c.frequency = f
}

func (c *ChannelPreference) SetQuietHours(qh *QuietHours) {
	c.quietHours = qh
}

type NotificationPreferences struct {
	id        PreferencesID
	userID    domain.UserID
	channels  map[Channel]ChannelPreference
	createdAt time.Time
	updatedAt time.Time
}

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

func (p *NotificationPreferences) ID() PreferencesID {
	return p.id
}

func (p *NotificationPreferences) UserID() domain.UserID {
	return p.userID
}

func (p *NotificationPreferences) Channels() map[Channel]ChannelPreference {
	return p.channels
}

func (p *NotificationPreferences) CreatedAt() time.Time {
	return p.createdAt
}

func (p *NotificationPreferences) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *NotificationPreferences) GetChannelPreference(channel Channel) (ChannelPreference, bool) {
	pref, ok := p.channels[channel]
	return pref, ok
}

func (p *NotificationPreferences) EnableChannel(channel Channel) {
	if pref, ok := p.channels[channel]; ok {
		pref.Enable()
		p.channels[channel] = pref
	} else {
		p.channels[channel] = NewChannelPreference(true, FrequencyImmediate, nil)
	}
	p.updatedAt = time.Now()
}

func (p *NotificationPreferences) DisableChannel(channel Channel) {
	if pref, ok := p.channels[channel]; ok {
		pref.Disable()
		p.channels[channel] = pref
	}
	p.updatedAt = time.Now()
}

func (p *NotificationPreferences) SetChannelFrequency(channel Channel, frequency Frequency) {
	if pref, ok := p.channels[channel]; ok {
		pref.SetFrequency(frequency)
		p.channels[channel] = pref
	} else {
		p.channels[channel] = NewChannelPreference(true, frequency, nil)
	}
	p.updatedAt = time.Now()
}

func (p *NotificationPreferences) SetChannelQuietHours(channel Channel, quietHours *QuietHours) {
	if pref, ok := p.channels[channel]; ok {
		pref.SetQuietHours(quietHours)
		p.channels[channel] = pref
	}
	p.updatedAt = time.Now()
}

func (p *NotificationPreferences) IsChannelEnabled(channel Channel) bool {
	pref, ok := p.channels[channel]
	if !ok {
		return false
	}
	return pref.Enabled()
}

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
