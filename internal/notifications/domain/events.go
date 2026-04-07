// Package domain provides domain entities and value objects for the notifications module.
// This package contains the core business logic types for notification management,
// including preferences, templates, and domain events.
package domain

import (
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

// NotificationCreated is emitted when a new notification is created.
type NotificationCreated struct {
	notificationID NotificationID
	recipient      Recipient
	channel        Channel
	subject        string
	priority       Priority
	createdAt      time.Time
	occurredAt     time.Time
}

// NewNotificationCreated creates a new NotificationCreated event.
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

// EventName returns the event name for NotificationCreated.
func (e NotificationCreated) EventName() string {
	return "notification.created"
}

// OccurredAt returns when the NotificationCreated event occurred.
func (e NotificationCreated) OccurredAt() time.Time {
	return e.occurredAt
}

// NotificationID returns the notification's unique identifier.
func (e NotificationCreated) NotificationID() NotificationID {
	return e.notificationID
}

// Recipient returns the notification recipient.
func (e NotificationCreated) Recipient() Recipient {
	return e.recipient
}

// Channel returns the notification channel.
func (e NotificationCreated) Channel() Channel {
	return e.channel
}

// Subject returns the notification subject.
func (e NotificationCreated) Subject() string {
	return e.subject
}

// Priority returns the notification priority.
func (e NotificationCreated) Priority() Priority {
	return e.priority
}

// CreatedAt returns when the notification was created.
func (e NotificationCreated) CreatedAt() time.Time {
	return e.createdAt
}

// NotificationQueued is emitted when a notification is queued for sending.
type NotificationQueued struct {
	notificationID NotificationID
	channel        Channel
	priority       Priority
	queuedAt       time.Time
	occurredAt     time.Time
}

// NewNotificationQueued creates a new NotificationQueued event.
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

// EventName returns the event name for NotificationQueued.
func (e NotificationQueued) EventName() string {
	return "notification.queued"
}

// OccurredAt returns when the NotificationQueued event occurred.
func (e NotificationQueued) OccurredAt() time.Time {
	return e.occurredAt
}

// NotificationID returns the notification's unique identifier.
func (e NotificationQueued) NotificationID() NotificationID {
	return e.notificationID
}

// Channel returns the notification channel.
func (e NotificationQueued) Channel() Channel {
	return e.channel
}

// Priority returns the notification priority.
func (e NotificationQueued) Priority() Priority {
	return e.priority
}

// QueuedAt returns when the notification was queued.
func (e NotificationQueued) QueuedAt() time.Time {
	return e.queuedAt
}

// NotificationSent is emitted when a notification is successfully sent.
type NotificationSent struct {
	notificationID NotificationID
	channel        Channel
	sentAt         time.Time
	occurredAt     time.Time
}

// NewNotificationSent creates a new NotificationSent event.
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

// EventName returns the event name for NotificationSent.
func (e NotificationSent) EventName() string {
	return "notification.sent"
}

// OccurredAt returns when the NotificationSent event occurred.
func (e NotificationSent) OccurredAt() time.Time {
	return e.occurredAt
}

// NotificationID returns the notification's unique identifier.
func (e NotificationSent) NotificationID() NotificationID {
	return e.notificationID
}

// Channel returns the notification channel.
func (e NotificationSent) Channel() Channel {
	return e.channel
}

// SentAt returns when the notification was sent.
func (e NotificationSent) SentAt() time.Time {
	return e.sentAt
}

// NotificationDelivered is emitted when a notification is delivered to the recipient.
type NotificationDelivered struct {
	notificationID NotificationID
	channel        Channel
	deliveredAt    time.Time
	occurredAt     time.Time
}

// NewNotificationDelivered creates a new NotificationDelivered event.
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

// EventName returns the event name for NotificationDelivered.
func (e NotificationDelivered) EventName() string {
	return "notification.delivered"
}

// OccurredAt returns when the NotificationDelivered event occurred.
func (e NotificationDelivered) OccurredAt() time.Time {
	return e.occurredAt
}

// NotificationID returns the notification's unique identifier.
func (e NotificationDelivered) NotificationID() NotificationID {
	return e.notificationID
}

// Channel returns the notification channel.
func (e NotificationDelivered) Channel() Channel {
	return e.channel
}

// DeliveredAt returns when the notification was delivered.
func (e NotificationDelivered) DeliveredAt() time.Time {
	return e.deliveredAt
}

// NotificationFailed is emitted when a notification fails to send.
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

// NewNotificationFailed creates a new NotificationFailed event.
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

// EventName returns the event name for NotificationFailed.
func (e NotificationFailed) EventName() string {
	return "notification.failed"
}

// OccurredAt returns when the NotificationFailed event occurred.
func (e NotificationFailed) OccurredAt() time.Time {
	return e.occurredAt
}

// NotificationID returns the notification's unique identifier.
func (e NotificationFailed) NotificationID() NotificationID {
	return e.notificationID
}

// Channel returns the notification channel.
func (e NotificationFailed) Channel() Channel {
	return e.channel
}

// Error returns the error message describing the failure.
func (e NotificationFailed) Error() string {
	return e.error
}

// Attempts returns the number of delivery attempts made.
func (e NotificationFailed) Attempts() int {
	return e.attempts
}

// WillRetry returns whether the notification will be retried.
func (e NotificationFailed) WillRetry() bool {
	return e.willRetry
}

// NextRetryAt returns the scheduled time for the next retry attempt.
func (e NotificationFailed) NextRetryAt() *time.Time {
	return e.nextRetryAt
}

// FailedAt returns when the notification failed.
func (e NotificationFailed) FailedAt() time.Time {
	return e.failedAt
}

// PreferenceUpdated is emitted when a user updates their notification preferences.
type PreferenceUpdated struct {
	userID     domain.UserID
	channel    Channel
	enabled    bool
	frequency  Frequency
	updatedAt  time.Time
	occurredAt time.Time
}

// NewPreferenceUpdated creates a new PreferenceUpdated event.
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

// EventName returns the event name for PreferenceUpdated.
func (e PreferenceUpdated) EventName() string {
	return "notification.preference_updated"
}

// OccurredAt returns when the PreferenceUpdated event occurred.
func (e PreferenceUpdated) OccurredAt() time.Time {
	return e.occurredAt
}

// UserID returns the user's unique identifier.
func (e PreferenceUpdated) UserID() domain.UserID {
	return e.userID
}

// Channel returns the notification channel.
func (e PreferenceUpdated) Channel() Channel {
	return e.channel
}

// Enabled returns whether the channel is enabled for the user.
func (e PreferenceUpdated) Enabled() bool {
	return e.enabled
}

// Frequency returns the notification frequency setting.
func (e PreferenceUpdated) Frequency() Frequency {
	return e.frequency
}

// UpdatedAt returns when the preference was updated.
func (e PreferenceUpdated) UpdatedAt() time.Time {
	return e.updatedAt
}
