package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jmoiron/sqlx"
)

// NotificationRepository implements the notification repository interface
// using SQL database storage.
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new notification repository with the provided database connection.
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

type notificationRow struct {
	ID            string         `db:"id"`
	UserID        sql.NullString `db:"user_id"`
	Email         sql.NullString `db:"email"`
	Phone         sql.NullString `db:"phone"`
	DeviceToken   sql.NullString `db:"device_token"`
	Channel       string         `db:"channel"`
	Subject       sql.NullString `db:"subject"`
	Content       string         `db:"content"`
	HTMLContent   sql.NullString `db:"html_content"`
	Status        string         `db:"status"`
	Priority      string         `db:"priority"`
	ScheduledAt   sql.NullString `db:"scheduled_at"`
	SentAt        sql.NullString `db:"sent_at"`
	DeliveredAt   sql.NullString `db:"delivered_at"`
	FailedAt      sql.NullString `db:"failed_at"`
	FailureReason sql.NullString `db:"failure_reason"`
	Attempts      int            `db:"attempts"`
	MaxAttempts   int            `db:"max_attempts"`
	Metadata      sql.NullString `db:"metadata"`
	CreatedAt     string         `db:"created_at"`
	UpdatedAt     string         `db:"updated_at"`
}

// Create persists a new notification to the database.
func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	metadataJSON, err := json.Marshal(notification.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, user_id, email, phone, device_token, channel, subject, content, html_content,
			status, priority, scheduled_at, attempts, max_attempts, metadata, created_at, updated_at
		) VALUES (
			:id, :user_id, :email, :phone, :device_token, :channel, :subject, :content, :html_content,
			:status, :priority, :scheduled_at, :attempts, :max_attempts, :metadata, :created_at, :updated_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":           notification.ID().String(),
		"user_id":      nullStringFromUserID(notification.Recipient().UserID),
		"email":        nullString(notification.Recipient().Email),
		"phone":        nullString(notification.Recipient().Phone),
		"device_token": nullString(notification.Recipient().DeviceToken),
		"channel":      notification.Channel().String(),
		"subject":      nullString(notification.Subject()),
		"content":      notification.Content().Text,
		"html_content": nullString(notification.Content().HTML),
		"status":       notification.Status().String(),
		"priority":     notification.Priority().String(),
		"scheduled_at": nullTime(notification.ScheduledAt()),
		"attempts":     notification.Attempts(),
		"max_attempts": notification.MaxAttempts(),
		"metadata":     string(metadataJSON),
		"created_at":   notification.CreatedAt().Format(time.RFC3339),
		"updated_at":   notification.UpdatedAt().Format(time.RFC3339),
	})

	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	return nil
}

// Update modifies an existing notification in the database.
func (r *NotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	metadataJSON, err := json.Marshal(notification.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		UPDATE notifications SET
			status = :status,
			sent_at = :sent_at,
			delivered_at = :delivered_at,
			failed_at = :failed_at,
			failure_reason = :failure_reason,
			attempts = :attempts,
			scheduled_at = :scheduled_at,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id
	`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             notification.ID().String(),
		"status":         notification.Status().String(),
		"sent_at":        nullTime(notification.SentAt()),
		"delivered_at":   nullTime(notification.DeliveredAt()),
		"failed_at":      nullTime(notification.FailedAt()),
		"failure_reason": nullString(notification.FailureReason()),
		"attempts":       notification.Attempts(),
		"scheduled_at":   nullTime(notification.ScheduledAt()),
		"metadata":       string(metadataJSON),
		"updated_at":     notification.UpdatedAt().Format(time.RFC3339),
	})

	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by its unique identifier.
// Returns domain.ErrNotificationNotFound if no matching notification exists.
func (r *NotificationRepository) GetByID(ctx context.Context, id domain.NotificationID) (*domain.Notification, error) {
	var row notificationRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications WHERE id = ?
	`, id.String())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("get notification by id: %w", err)
	}

	return r.scanNotification(row)
}

// GetByStatus retrieves notifications by status with a limit, ordered by priority and creation time.
func (r *NotificationRepository) GetByStatus(ctx context.Context, status domain.Status, limit int) ([]*domain.Notification, error) {
	var rows []notificationRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = ?
		ORDER BY priority DESC, created_at ASC
		LIMIT ?
	`, status.String(), limit)

	if err != nil {
		return nil, fmt.Errorf("get notifications by status: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		n, err := r.scanNotification(row)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications[i] = n
	}

	return notifications, nil
}

// GetPendingByUser retrieves all pending, queued, or sending notifications for a user.
func (r *NotificationRepository) GetPendingByUser(ctx context.Context, userID string) ([]*domain.Notification, error) {
	var rows []notificationRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE user_id = ? AND status IN (?, ?, ?)
		ORDER BY created_at ASC
	`, userID, domain.StatusPending, domain.StatusQueued, domain.StatusSending)

	if err != nil {
		return nil, fmt.Errorf("get pending notifications by user: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		n, err := r.scanNotification(row)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications[i] = n
	}

	return notifications, nil
}

// GetScheduled retrieves scheduled notifications ready to be sent.
func (r *NotificationRepository) GetScheduled(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	var rows []notificationRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?
		ORDER BY scheduled_at ASC
		LIMIT ?
	`, domain.StatusPending, before.Format(time.RFC3339), limit)

	if err != nil {
		return nil, fmt.Errorf("get scheduled notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		n, err := r.scanNotification(row)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications[i] = n
	}

	return notifications, nil
}

// GetStalled retrieves notifications stuck in sending status for too long.
func (r *NotificationRepository) GetStalled(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.Notification, error) {
	cutoff := time.Now().Add(-olderThan)
	var rows []notificationRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = ? AND updated_at < ?
		ORDER BY updated_at ASC
		LIMIT ?
	`, domain.StatusSending, cutoff.Format(time.RFC3339), limit)

	if err != nil {
		return nil, fmt.Errorf("get stalled notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		n, err := r.scanNotification(row)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications[i] = n
	}

	return notifications, nil
}

// Delete removes a notification from the database.
func (r *NotificationRepository) Delete(ctx context.Context, id domain.NotificationID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notifications WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

// DeleteCompleted removes completed (delivered or failed) notifications older than the specified duration.
// Returns the number of notifications deleted.
func (r *NotificationRepository) DeleteCompleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM notifications
		WHERE status IN (?, ?) AND updated_at < ?
	`, domain.StatusDelivered, domain.StatusFailed, cutoff.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete completed notifications: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}

	return affected, nil
}

// scanNotification converts a database row into a domain Notification entity.
func (r *NotificationRepository) scanNotification(row notificationRow) (*domain.Notification, error) {
	recipient := domain.Recipient{}
	if row.UserID.Valid {
		userID := identityDomain.UserID(row.UserID.String)
		recipient.UserID = &userID
	}
	if row.Email.Valid {
		recipient.Email = row.Email.String
	}
	if row.Phone.Valid {
		recipient.Phone = row.Phone.String
	}
	if row.DeviceToken.Valid {
		recipient.DeviceToken = row.DeviceToken.String
	}

	channel, err := domain.ParseChannel(row.Channel)
	if err != nil {
		return nil, fmt.Errorf("parse channel: %w", err)
	}

	content := domain.Content{Text: row.Content}
	if row.HTMLContent.Valid {
		content.HTML = row.HTMLContent.String
	}

	status, err := domain.ParseStatus(row.Status)
	if err != nil {
		return nil, fmt.Errorf("parse status: %w", err)
	}

	priority, err := domain.ParsePriority(row.Priority)
	if err != nil {
		return nil, fmt.Errorf("parse priority: %w", err)
	}

	metadata := make(map[string]string)
	if row.Metadata.Valid && row.Metadata.String != "" {
		if err := json.Unmarshal([]byte(row.Metadata.String), &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	opts := []domain.NotificationOption{
		domain.WithMaxAttempts(row.MaxAttempts),
	}
	if row.ScheduledAt.Valid {
		t, err := time.Parse(time.RFC3339, row.ScheduledAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse scheduled_at: %w", err)
		}
		opts = append(opts, domain.WithScheduledAt(t))
	}

	var subject string
	if row.Subject.Valid {
		subject = row.Subject.String
	}

	notification, err := domain.NewNotification(
		recipient,
		channel,
		subject,
		content,
		priority,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	if len(metadata) > 0 {
		for k, v := range metadata {
			notification.Metadata()[k] = v
		}
	}

	for i := 0; i < row.Attempts; i++ {
		notification.IncrementAttempts()
	}

	switch status {
	case domain.StatusQueued:
		_ = notification.Queue()
	case domain.StatusSending:
		_ = notification.Queue()
		_ = notification.StartSending()
	case domain.StatusSent:
		_ = notification.Queue()
		_ = notification.StartSending()
		_ = notification.MarkSent()
	case domain.StatusDelivered:
		_ = notification.Queue()
		_ = notification.StartSending()
		_ = notification.MarkSent()
		_ = notification.MarkDelivered()
	case domain.StatusFailed:
		reason := ""
		if row.FailureReason.Valid {
			reason = row.FailureReason.String
		}
		_ = notification.MarkFailed(reason)
	}

	return notification, nil
}

// nullString converts a string to a sql.NullString for database storage.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullStringFromUserID converts a user ID pointer to a sql.NullString for database storage.
func nullStringFromUserID(userID *identityDomain.UserID) sql.NullString {
	if userID == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(*userID), Valid: true}
}

// nullTime converts a time pointer to a sql.NullString for database storage.
func nullTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: t.Format(time.RFC3339), Valid: true}
}
