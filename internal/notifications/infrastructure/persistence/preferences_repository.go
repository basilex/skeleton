// Package persistence provides database repository implementations for the notifications domain.
// This package contains SQLite-based repositories for notifications, templates, and preferences.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jmoiron/sqlx"
)

// PreferencesRepository implements the notification preferences repository interface
// using SQL database storage.
type PreferencesRepository struct {
	db *sqlx.DB
}

// NewPreferencesRepository creates a new preferences repository with the provided database connection.
func NewPreferencesRepository(db *sqlx.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

type preferencesRow struct {
	ID          string `db:"id"`
	UserID      string `db:"user_id"`
	Preferences string `db:"preferences"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

// GetByUserID retrieves notification preferences for a specific user.
// Returns domain.ErrPreferencesNotFound if no preferences exist.
func (r *PreferencesRepository) GetByUserID(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	var row preferencesRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, user_id, preferences, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = ?
	`, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPreferencesNotFound
		}
		return nil, fmt.Errorf("get preferences by user id: %w", err)
	}

	return r.scanPreferences(row)
}

// Upsert creates or updates notification preferences for a user.
func (r *PreferencesRepository) Upsert(ctx context.Context, preferences *domain.NotificationPreferences) error {
	preferencesJSON, err := r.marshalPreferences(preferences)
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}

	query := `
		INSERT INTO notification_preferences (id, user_id, preferences, created_at, updated_at)
		VALUES (:id, :user_id, :preferences, :created_at, :updated_at)
		ON CONFLICT(user_id) DO UPDATE SET
			preferences = excluded.preferences,
			updated_at = excluded.updated_at
	`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          preferences.ID().String(),
		"user_id":     string(preferences.UserID()),
		"preferences": preferencesJSON,
		"created_at":  preferences.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":  preferences.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})

	if err != nil {
		return fmt.Errorf("upsert preferences: %w", err)
	}

	return nil
}

// Delete removes notification preferences for a user.
func (r *PreferencesRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_preferences WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete preferences: %w", err)
	}
	return nil
}

// scanPreferences converts a database row into a domain NotificationPreferences entity.
func (r *PreferencesRepository) scanPreferences(row preferencesRow) (*domain.NotificationPreferences, error) {
	userID := identityDomain.UserID(row.UserID)
	preferences := domain.NewNotificationPreferences(userID)

	var data struct {
		Channels map[string]struct {
			Enabled    bool   `json:"enabled"`
			Frequency  string `json:"frequency"`
			QuietHours *struct {
				StartHour int    `json:"start_hour"`
				EndHour   int    `json:"end_hour"`
				Timezone  string `json:"timezone"`
			} `json:"quiet_hours"`
		} `json:"channels"`
	}

	if err := json.Unmarshal([]byte(row.Preferences), &data); err != nil {
		return nil, fmt.Errorf("unmarshal preferences: %w", err)
	}

	for ch, pref := range data.Channels {
		channel, err := domain.ParseChannel(ch)
		if err != nil {
			continue
		}

		if !pref.Enabled {
			preferences.DisableChannel(channel)
		}

		freq, err := domain.ParseFrequency(pref.Frequency)
		if err != nil {
			freq = domain.FrequencyImmediate
		}
		preferences.SetChannelFrequency(channel, freq)

		if pref.QuietHours != nil {
			qh, err := domain.NewQuietHours(pref.QuietHours.StartHour, pref.QuietHours.EndHour, pref.QuietHours.Timezone)
			if err == nil {
				preferences.SetChannelQuietHours(channel, qh)
			}
		}
	}

	return preferences, nil
}

// marshalPreferences serializes preferences to JSON for storage.
func (r *PreferencesRepository) marshalPreferences(preferences *domain.NotificationPreferences) (string, error) {
	data := struct {
		Channels map[string]struct {
			Enabled    bool   `json:"enabled"`
			Frequency  string `json:"frequency"`
			QuietHours *struct {
				StartHour int    `json:"start_hour"`
				EndHour   int    `json:"end_hour"`
				Timezone  string `json:"timezone"`
			} `json:"quiet_hours,omitempty"`
		} `json:"channels"`
	}{
		Channels: make(map[string]struct {
			Enabled    bool   `json:"enabled"`
			Frequency  string `json:"frequency"`
			QuietHours *struct {
				StartHour int    `json:"start_hour"`
				EndHour   int    `json:"end_hour"`
				Timezone  string `json:"timezone"`
			} `json:"quiet_hours,omitempty"`
		}),
	}

	for ch, pref := range preferences.Channels() {
		channelPref := struct {
			Enabled    bool   `json:"enabled"`
			Frequency  string `json:"frequency"`
			QuietHours *struct {
				StartHour int    `json:"start_hour"`
				EndHour   int    `json:"end_hour"`
				Timezone  string `json:"timezone"`
			} `json:"quiet_hours,omitempty"`
		}{
			Enabled:   pref.Enabled(),
			Frequency: pref.Frequency().String(),
		}

		if qh := pref.QuietHours(); qh != nil {
			channelPref.QuietHours = &struct {
				StartHour int    `json:"start_hour"`
				EndHour   int    `json:"end_hour"`
				Timezone  string `json:"timezone"`
			}{
				StartHour: qh.StartHour(),
				EndHour:   qh.EndHour(),
				Timezone:  qh.Timezone(),
			}
		}

		data.Channels[ch.String()] = channelPref
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
