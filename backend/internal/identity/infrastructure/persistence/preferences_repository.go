package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferencesRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewPreferencesRepository(pool *pgxpool.Pool) *PreferencesRepository {
	return &PreferencesRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type preferencesDTO struct {
	ID                    string    `db:"id"`
	UserID                string    `db:"user_id"`
	Theme                 string    `db:"theme"`
	Language              string    `db:"language"`
	DateFormat            string    `db:"date_format"`
	Timezone              string    `db:"timezone"`
	EmailNotifications    bool      `db:"email_notifications"`
	SMSNotifications      bool      `db:"sms_notifications"`
	PushNotifications     bool      `db:"push_notifications"`
	InAppNotifications    bool      `db:"in_app_notifications"`
	MarketingEmails       bool      `db:"marketing_emails"`
	WeeklyDigest          bool      `db:"weekly_digest"`
	QuietHoursStart       *int      `db:"quiet_hours_start"`
	QuietHoursEnd         *int      `db:"quiet_hours_end"`
	NotificationsTimezone string    `db:"notifications_timezone"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

func (r *PreferencesRepository) Save(ctx context.Context, prefs *domain.UserPreferences) error {
	notifications := prefs.Notifications()

	query, args, err := r.psql.Insert("user_preferences").
		Columns("id", "user_id", "theme", "language", "date_format", "timezone",
			"email_notifications", "sms_notifications", "push_notifications", "in_app_notifications",
			"marketing_emails", "weekly_digest", "quiet_hours_start", "quiet_hours_end",
			"notifications_timezone", "created_at", "updated_at").
		Values(prefs.ID(), prefs.UserID(), prefs.Theme(), prefs.Language(), prefs.DateFormat(),
			prefs.Timezone(), notifications.EmailEnabled(), notifications.SMSEnabled(),
			notifications.PushEnabled(), notifications.InAppEnabled(),
			notifications.MarketingEmails(), notifications.WeeklyDigest(),
			notifications.QuietHoursStart(), notifications.QuietHoursEnd(), notifications.Timezone(),
			prefs.CreatedAt(), prefs.UpdatedAt()).
		Suffix("ON CONFLICT(user_id) DO UPDATE SET theme = EXCLUDED.theme, language = EXCLUDED.language, date_format = EXCLUDED.date_format, timezone = EXCLUDED.timezone, email_notifications = EXCLUDED.email_notifications, sms_notifications = EXCLUDED.sms_notifications, push_notifications = EXCLUDED.push_notifications, in_app_notifications = EXCLUDED.in_app_notifications, marketing_emails = EXCLUDED.marketing_emails, weekly_digest = EXCLUDED.weekly_digest, quiet_hours_start = EXCLUDED.quiet_hours_start, quiet_hours_end = EXCLUDED.quiet_hours_end, notifications_timezone = EXCLUDED.notifications_timezone, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build preferences insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	return nil
}

func (r *PreferencesRepository) FindByID(ctx context.Context, id domain.PreferencesID) (*domain.UserPreferences, error) {
	var dto preferencesDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, user_id, theme, language, date_format, timezone,
		 email_notifications, sms_notifications, push_notifications, in_app_notifications,
		 marketing_emails, weekly_digest, quiet_hours_start, quiet_hours_end,
		 notifications_timezone, created_at, updated_at
		 FROM user_preferences WHERE id = $1`,
		id)
	if err != nil {
		return nil, fmt.Errorf("find preferences by id: %w", err)
	}
	return r.dtoToDomain(dto)
}

func (r *PreferencesRepository) FindByUserID(ctx context.Context, userID domain.UserID) (*domain.UserPreferences, error) {
	var dto preferencesDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, user_id, theme, language, date_format, timezone,
		 email_notifications, sms_notifications, push_notifications, in_app_notifications,
		 marketing_emails, weekly_digest, quiet_hours_start, quiet_hours_end,
		 notifications_timezone, created_at, updated_at
		 FROM user_preferences WHERE user_id = $1`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("find preferences by user id: %w", err)
	}
	return r.dtoToDomain(dto)
}

func (r *PreferencesRepository) Delete(ctx context.Context, id domain.PreferencesID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM user_preferences WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete preferences: %w", err)
	}
	return nil
}

func (r *PreferencesRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM user_preferences WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete preferences by user: %w", err)
	}
	return nil
}

func (r *PreferencesRepository) dtoToDomain(dto preferencesDTO) (*domain.UserPreferences, error) {
	userID, err := domain.ParseUserID(dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	prefsID, err := domain.ParsePreferencesID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("parse preferences id: %w", err)
	}

	theme, err := domain.ParseTheme(dto.Theme)
	if err != nil {
		theme = domain.ThemeAuto
	}

	language, err := domain.ParseLanguage(dto.Language)
	if err != nil {
		language = domain.LanguageEn
	}

	dateFormat, err := domain.ParseDateFormat(dto.DateFormat)
	if err != nil {
		dateFormat = domain.DateFormatYMD
	}

	notifications := domain.ReconstituteNotificationSettings(
		dto.EmailNotifications,
		dto.SMSNotifications,
		dto.PushNotifications,
		dto.InAppNotifications,
		dto.MarketingEmails,
		dto.WeeklyDigest,
		dto.QuietHoursStart,
		dto.QuietHoursEnd,
		dto.NotificationsTimezone,
	)

	return domain.ReconstituteUserPreferences(
		prefsID,
		userID,
		theme,
		language,
		dateFormat,
		dto.Timezone,
		notifications,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}
