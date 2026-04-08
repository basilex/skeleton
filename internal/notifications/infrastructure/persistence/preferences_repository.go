package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferencesRepository struct {
	pool *pgxpool.Pool
}

func NewPreferencesRepository(pool *pgxpool.Pool) *PreferencesRepository {
	return &PreferencesRepository{pool: pool}
}

func (r *PreferencesRepository) GetByUserID(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, preferences, created_at, updated_at
		FROM notification_preference
		WHERE user_id = $1
	`

	row := r.pool.QueryRow(ctx, query, userID)

	return r.scanPreferences(row)
}

func (r *PreferencesRepository) Upsert(ctx context.Context, preferences *domain.NotificationPreferences) error {
	preferencesJSON, err := r.marshalPreferences(preferences)
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}

	query := `
		INSERT INTO notification_preference (id, user_id, preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(user_id) DO UPDATE SET
			preferences = EXCLUDED.preferences,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.pool.Exec(ctx, query,
		preferences.ID().String(),
		preferences.UserID().String(),
		preferencesJSON,
		preferences.CreatedAt(),
		preferences.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("upsert preferences: %w", err)
	}

	return nil
}

func (r *PreferencesRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_preference WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete preferences: %w", err)
	}
	return nil
}

func (r *PreferencesRepository) scanPreferences(row pgx.Row) (*domain.NotificationPreferences, error) {
	var id, userID string
	var preferencesBytes []byte
	var createdAt, updatedAt string

	err := row.Scan(&id, &userID, &preferencesBytes, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPreferencesNotFound
		}
		return nil, fmt.Errorf("scan preferences: %w", err)
	}

	uid, err := identityDomain.ParseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	preferences := domain.NewNotificationPreferences(uid)

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

	if err := json.Unmarshal(preferencesBytes, &data); err != nil {
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
