package domain

import (
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

type NotificationCreated struct {
	notificationID NotificationID
	recipient      Recipient
	channel        Channel
	subject        string
	priority       Priority
	createdAt      time.Time
	occurredAt     time.Time
}

func NewNotificationCreated(
	notificationID NotificationID,
	recipient Recipient,
	channel Channel,
	subject string,
	priority Priority,
) NotificationCreated {
	now := time.Now()
	return NotificationCreated{
		notificationID: notificationID,
		recipient:      recipient,
		channel:        channel,
		subject:        subject,
		priority:       priority,
		createdAt:      now,
		occurredAt:     now,
	}
}

func (e NotificationCreated) EventName() string {
	return "notification.created"
}

func (e NotificationCreated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e NotificationCreated) NotificationID() NotificationID {
	return e.notificationID
}

func (e NotificationCreated) Recipient() Recipient {
	return e.recipient
}

func (e NotificationCreated) Channel() Channel {
	return e.channel
}

func (e NotificationCreated) Subject() string {
	return e.subject
}

func (e NotificationCreated) Priority() Priority {
	return e.priority
}

func (e NotificationCreated) CreatedAt() time.Time {
	return e.createdAt
}

type NotificationQueued struct {
	notificationID NotificationID
	channel        Channel
	priority       Priority
	queuedAt       time.Time
	occurredAt     time.Time
}

func NewNotificationQueued(
	notificationID NotificationID,
	channel Channel,
	priority Priority,
) NotificationQueued {
	now := time.Now()
	return NotificationQueued{
		notificationID: notificationID,
		channel:        channel,
		priority:       priority,
		queuedAt:       now,
		occurredAt:     now,
	}
}

func (e NotificationQueued) EventName() string {
	return "notification.queued"
}

func (e NotificationQueued) OccurredAt() time.Time {
	return e.occurredAt
}

func (e NotificationQueued) NotificationID() NotificationID {
	return e.notificationID
}

func (e NotificationQueued) Channel() Channel {
	return e.channel
}

func (e NotificationQueued) Priority() Priority {
	return e.priority
}

func (e NotificationQueued) QueuedAt() time.Time {
	return e.queuedAt
}

type NotificationSent struct {
	notificationID NotificationID
	channel        Channel
	sentAt         time.Time
	occurredAt     time.Time
}

func NewNotificationSent(
	notificationID NotificationID,
	channel Channel,
) NotificationSent {
	now := time.Now()
	return NotificationSent{
		notificationID: notificationID,
		channel:        channel,
		sentAt:         now,
		occurredAt:     now,
	}
}

func (e NotificationSent) EventName() string {
	return "notification.sent"
}

func (e NotificationSent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e NotificationSent) NotificationID() NotificationID {
	return e.notificationID
}

func (e NotificationSent) Channel() Channel {
	return e.channel
}

func (e NotificationSent) SentAt() time.Time {
	return e.sentAt
}

type NotificationDelivered struct {
	notificationID NotificationID
	channel        Channel
	deliveredAt    time.Time
	occurredAt     time.Time
}

func NewNotificationDelivered(
	notificationID NotificationID,
	channel Channel,
) NotificationDelivered {
	now := time.Now()
	return NotificationDelivered{
		notificationID: notificationID,
		channel:        channel,
		deliveredAt:    now,
		occurredAt:     now,
	}
}

func (e NotificationDelivered) EventName() string {
	return "notification.delivered"
}

func (e NotificationDelivered) OccurredAt() time.Time {
	return e.occurredAt
}

func (e NotificationDelivered) NotificationID() NotificationID {
	return e.notificationID
}

func (e NotificationDelivered) Channel() Channel {
	return e.channel
}

func (e NotificationDelivered) DeliveredAt() time.Time {
	return e.deliveredAt
}

type NotificationFailed struct {
	notificationID NotificationID
	channel        Channel
	error          string
	attempts       int
	willRetry      bool
	nextRetryAt    *time.Time
	failedAt       time.Time
	occurredAt     time.Time
}

func NewNotificationFailed(
	notificationID NotificationID,
	channel Channel,
	errorMsg string,
	attempts int,
	willRetry bool,
	nextRetryAt *time.Time,
) NotificationFailed {
	now := time.Now()
	return NotificationFailed{
		notificationID: notificationID,
		channel:        channel,
		error:          errorMsg,
		attempts:       attempts,
		willRetry:      willRetry,
		nextRetryAt:    nextRetryAt,
		failedAt:       now,
		occurredAt:     now,
	}
}

func (e NotificationFailed) EventName() string {
	return "notification.failed"
}

func (e NotificationFailed) OccurredAt() time.Time {
	return e.occurredAt
}

func (e NotificationFailed) NotificationID() NotificationID {
	return e.notificationID
}

func (e NotificationFailed) Channel() Channel {
	return e.channel
}

func (e NotificationFailed) Error() string {
	return e.error
}

func (e NotificationFailed) Attempts() int {
	return e.attempts
}

func (e NotificationFailed) WillRetry() bool {
	return e.willRetry
}

func (e NotificationFailed) NextRetryAt() *time.Time {
	return e.nextRetryAt
}

func (e NotificationFailed) FailedAt() time.Time {
	return e.failedAt
}

type PreferenceUpdated struct {
	userID     domain.UserID
	channel    Channel
	enabled    bool
	frequency  Frequency
	updatedAt  time.Time
	occurredAt time.Time
}

func NewPreferenceUpdated(
	userID domain.UserID,
	channel Channel,
	enabled bool,
	frequency Frequency,
) PreferenceUpdated {
	now := time.Now()
	return PreferenceUpdated{
		userID:     userID,
		channel:    channel,
		enabled:    enabled,
		frequency:  frequency,
		updatedAt:  now,
		occurredAt: now,
	}
}

func (e PreferenceUpdated) EventName() string {
	return "notification.preference_updated"
}

func (e PreferenceUpdated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e PreferenceUpdated) UserID() domain.UserID {
	return e.userID
}

func (e PreferenceUpdated) Channel() Channel {
	return e.channel
}

func (e PreferenceUpdated) Enabled() bool {
	return e.enabled
}

func (e PreferenceUpdated) Frequency() Frequency {
	return e.frequency
}

func (e PreferenceUpdated) UpdatedAt() time.Time {
	return e.updatedAt
}
