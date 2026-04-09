package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	metadataJSON, err := json.Marshal(notification.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, user_id, email, phone, device_token, channel, subject, content, html_content,
			status, priority, scheduled_at, attempts, max_attempts, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = r.pool.Exec(ctx, query,
		notification.ID().String(),
		nullStringFromUserID(notification.Recipient().UserID),
		notification.Recipient().Email,
		notification.Recipient().Phone,
		notification.Recipient().DeviceToken,
		notification.Channel().String(),
		notification.Subject(),
		notification.Content().Text,
		notification.Content().HTML,
		notification.Status().String(),
		notification.Priority().String(),
		notification.ScheduledAt(),
		notification.Attempts(),
		notification.MaxAttempts(),
		string(metadataJSON),
		notification.CreatedAt().Format(time.RFC3339),
		notification.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	return nil
}

func (r *NotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	metadataJSON, err := json.Marshal(notification.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		UPDATE notifications SET
			status = $1,
			sent_at = $2,
			delivered_at = $3,
			failed_at = $4,
			failure_reason = $5,
			attempts = $6,
			scheduled_at = $7,
			metadata = $8,
			updated_at = $9
		WHERE id = $10
	`

	_, err = r.pool.Exec(ctx, query,
		notification.Status().String(),
		notification.SentAt(),
		notification.DeliveredAt(),
		notification.FailedAt(),
		notification.FailureReason(),
		notification.Attempts(),
		notification.ScheduledAt(),
		string(metadataJSON),
		notification.UpdatedAt().Format(time.RFC3339),
		notification.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id domain.NotificationID) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id.String())

	return r.scanNotification(row)
}

func (r *NotificationRepository) GetByStatus(ctx context.Context, status domain.Status, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = $1
		ORDER BY priority DESC, created_at ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, status.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("get notifications by status: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		n, err := r.scanNotificationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) GetPendingByUser(ctx context.Context, userID string) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND status IN ($2, $3, $4)
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, userID, domain.StatusPending, domain.StatusQueued, domain.StatusSending)
	if err != nil {
		return nil, fmt.Errorf("get pending notifications by user: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		n, err := r.scanNotificationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) GetScheduled(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = $1 AND scheduled_at IS NOT NULL AND scheduled_at <= $2
		ORDER BY scheduled_at ASC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, domain.StatusPending, before, limit)
	if err != nil {
		return nil, fmt.Errorf("get scheduled notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		n, err := r.scanNotificationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) GetStalled(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.Notification, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `
		SELECT id, user_id, email, phone, device_token, channel, subject, content, html_content,
			   status, priority, scheduled_at, sent_at, delivered_at, failed_at, failure_reason,
			   attempts, max_attempts, metadata, created_at, updated_at
		FROM notifications
		WHERE status = $1 AND updated_at < $2
		ORDER BY updated_at ASC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, domain.StatusSending, cutoff.Format(time.RFC3339), limit)
	if err != nil {
		return nil, fmt.Errorf("get stalled notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		n, err := r.scanNotificationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id domain.NotificationID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) DeleteCompleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.pool.Exec(ctx, `
		DELETE FROM notifications
		WHERE status IN ($1, $2) AND updated_at < $3
	`, domain.StatusDelivered, domain.StatusFailed, cutoff.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete completed notifications: %w", err)
	}

	return result.RowsAffected(), nil
}

func (r *NotificationRepository) scanNotification(row pgx.Row) (*domain.Notification, error) {
	var id, channel, content, status, priority string
	var subject, htmlContent, failureReason *string
	var userID, email, phone, deviceToken *string
	var scheduledAt, sentAt, deliveredAt, failedAt *time.Time
	var createdAt, updatedAt time.Time
	var attempts, maxAttempts int
	var metadataBytes []byte

	err := row.Scan(
		&id, &userID, &email, &phone, &deviceToken, &channel, &subject, &content, &htmlContent,
		&status, &priority, &scheduledAt, &sentAt, &deliveredAt, &failedAt, &failureReason,
		&attempts, &maxAttempts, &metadataBytes, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("scan notification: %w", err)
	}

	return r.reconstituteNotification(
		id, userID, email, phone, deviceToken, channel, subject, content, htmlContent,
		status, priority, scheduledAt, sentAt, deliveredAt, failedAt, failureReason,
		attempts, maxAttempts, metadataBytes, createdAt, updatedAt,
	)
}

func (r *NotificationRepository) scanNotificationFromRows(rows pgx.Rows) (*domain.Notification, error) {
	var id, channel, content, status, priority string
	var subject, htmlContent, failureReason *string
	var userID, email, phone, deviceToken *string
	var scheduledAt, sentAt, deliveredAt, failedAt *time.Time
	var createdAt, updatedAt time.Time
	var attempts, maxAttempts int
	var metadataBytes []byte

	err := rows.Scan(
		&id, &userID, &email, &phone, &deviceToken, &channel, &subject, &content, &htmlContent,
		&status, &priority, &scheduledAt, &sentAt, &deliveredAt, &failedAt, &failureReason,
		&attempts, &maxAttempts, &metadataBytes, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan notification: %w", err)
	}

	return r.reconstituteNotification(
		id, userID, email, phone, deviceToken, channel, subject, content, htmlContent,
		status, priority, scheduledAt, sentAt, deliveredAt, failedAt, failureReason,
		attempts, maxAttempts, metadataBytes, createdAt, updatedAt,
	)
}

func (r *NotificationRepository) reconstituteNotification(
	id string, userID, email, phone, deviceToken *string, channel string, subject *string,
	content string, htmlContent *string, status, priority string,
	scheduledAt, sentAt, deliveredAt, failedAt *time.Time, failureReason *string,
	attempts, maxAttempts int, metadataBytes []byte, createdAt, updatedAt time.Time,
) (*domain.Notification, error) {
	recipient := domain.Recipient{}
	if userID != nil {
		uid, parseErr := identityDomain.ParseUserID(*userID)
		if parseErr != nil {
			return nil, fmt.Errorf("parse user id: %w", parseErr)
		}
		recipient.UserID = &uid
	}
	if email != nil {
		recipient.Email = *email
	}
	if phone != nil {
		recipient.Phone = *phone
	}
	if deviceToken != nil {
		recipient.DeviceToken = *deviceToken
	}

	channelEnum, err := domain.ParseChannel(channel)
	if err != nil {
		return nil, fmt.Errorf("parse channel: %w", err)
	}

	contentObj := domain.Content{Text: content}
	if htmlContent != nil {
		contentObj.HTML = *htmlContent
	}

	statusEnum, err := domain.ParseStatus(status)
	if err != nil {
		return nil, fmt.Errorf("parse status: %w", err)
	}

	priorityEnum, err := domain.ParsePriority(priority)
	if err != nil {
		return nil, fmt.Errorf("parse priority: %w", err)
	}

	metadata := make(map[string]string)
	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	opts := []domain.NotificationOption{
		domain.WithMaxAttempts(maxAttempts),
	}
	if scheduledAt != nil {
		opts = append(opts, domain.WithScheduledAt(*scheduledAt))
	}

	var subj string
	if subject != nil {
		subj = *subject
	}

	notification, err := domain.NewNotification(
		recipient,
		channelEnum,
		subj,
		contentObj,
		priorityEnum,
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

	for i := 0; i < attempts; i++ {
		notification.IncrementAttempts()
	}

	switch statusEnum {
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
		if failureReason != nil {
			reason = *failureReason
		}
		_ = notification.MarkFailed(reason)
	}

	return notification, nil
}

func nullStringFromUserID(userID *identityDomain.UserID) *string {
	if userID == nil {
		return nil
	}
	uid := userID.String()
	return &uid
}
