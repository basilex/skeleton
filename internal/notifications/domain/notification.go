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

// NotificationID is a unique identifier for a notification.
type NotificationID uuid.UUID

// NewNotificationID generates a new unique NotificationID using UUID v7.
func NewNotificationID() NotificationID {
	return NotificationID(uuid.NewV7())
}

// ParseNotificationID validates and converts a string to NotificationID.
func ParseNotificationID(s string) (NotificationID, error) {
	if s == "" {
		return NotificationID{}, fmt.Errorf("notification ID cannot be empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return NotificationID{}, fmt.Errorf("invalid notification id: %w", err)
	}
	return NotificationID(u), nil
}

// String returns the string representation of NotificationID.
func (id NotificationID) String() string {
	return uuid.UUID(id).String()
}

// Channel represents a notification delivery channel.
type Channel string

// Channel type constants.
const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
	ChannelInApp Channel = "in_app"
)

// String returns the string representation of the channel.
func (c Channel) String() string {
	return string(c)
}

// ParseChannel converts a string to a Channel value.
func ParseChannel(s string) (Channel, error) {
	switch s {
	case string(ChannelEmail):
		return ChannelEmail, nil
	case string(ChannelSMS):
		return ChannelSMS, nil
	case string(ChannelPush):
		return ChannelPush, nil
	case string(ChannelInApp):
		return ChannelInApp, nil
	default:
		return "", fmt.Errorf("invalid channel: %s", s)
	}
}

// Status represents the lifecycle state of a notification.
type Status string

// Status type constants.
const (
	StatusPending   Status = "pending"
	StatusQueued    Status = "queued"
	StatusSending   Status = "sending"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// ParseStatus converts a string to a Status value.
func ParseStatus(s string) (Status, error) {
	switch s {
	case string(StatusPending):
		return StatusPending, nil
	case string(StatusQueued):
		return StatusQueued, nil
	case string(StatusSending):
		return StatusSending, nil
	case string(StatusSent):
		return StatusSent, nil
	case string(StatusDelivered):
		return StatusDelivered, nil
	case string(StatusFailed):
		return StatusFailed, nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}

// Priority represents the urgency level of a notification.
type Priority string

// Priority level constants.
const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// String returns the string representation of the priority.
func (p Priority) String() string {
	return string(p)
}

// ParsePriority converts a string to a Priority value.
// ParsePriority converts a string to a Priority value.
func ParsePriority(s string) (Priority, error) {
	switch s {
	case string(PriorityLow):
		return PriorityLow, nil
	case string(PriorityNormal):
		return PriorityNormal, nil
	case string(PriorityHigh):
		return PriorityHigh, nil
	case string(PriorityCritical):
		return PriorityCritical, nil
	default:
		return "", fmt.Errorf("invalid priority: %s", s)
	}
}

// Recipient contains the delivery information for a notification recipient.
type Recipient struct {
	UserID      *domain.UserID
	Email       string
	Phone       string
	DeviceToken string
}

// Content contains the notification message content in different formats.
type Content struct {
	Text string
	HTML string
}

// Notification represents a notification entity in the domain.
type Notification struct {
	id            NotificationID
	recipient     Recipient
	channel       Channel
	subject       string
	content       Content
	status        Status
	priority      Priority
	scheduledAt   *time.Time
	sentAt        *time.Time
	deliveredAt   *time.Time
	failedAt      *time.Time
	failureReason string
	attempts      int
	maxAttempts   int
	metadata      map[string]string
	createdAt     time.Time
	updatedAt     time.Time
}

// NewNotification creates a new notification with the provided details.
// Optional configuration can be applied via NotificationOption functions.
func NewNotification(
	recipient Recipient,
	channel Channel,
	subject string,
	content Content,
	priority Priority,
	opts ...NotificationOption,
) (*Notification, error) {
	if err := validateNotification(recipient, channel, subject, content); err != nil {
		return nil, err
	}

	now := time.Now()
	notification := &Notification{
		id:          NewNotificationID(),
		recipient:   recipient,
		channel:     channel,
		subject:     subject,
		content:     content,
		status:      StatusPending,
		priority:    priority,
		attempts:    0,
		maxAttempts: 5,
		metadata:    make(map[string]string),
		createdAt:   now,
		updatedAt:   now,
	}

	for _, opt := range opts {
		opt(notification)
	}

	return notification, nil
}

// NotificationOption is a functional option for configuring a Notification.
type NotificationOption func(*Notification)

// WithScheduledAt sets the scheduled delivery time for the notification.
func WithScheduledAt(t time.Time) NotificationOption {
	return func(n *Notification) {
		n.scheduledAt = &t
	}
}

// WithMaxAttempts sets the maximum number of delivery attempts.
func WithMaxAttempts(max int) NotificationOption {
	return func(n *Notification) {
		n.maxAttempts = max
	}
}

// WithMetadata adds metadata key-value pairs to the notification.
func WithMetadata(metadata map[string]string) NotificationOption {
	return func(n *Notification) {
		for k, v := range metadata {
			n.metadata[k] = v
		}
	}
}

// validateNotification validates that the notification has required fields for its channel.
func validateNotification(recipient Recipient, channel Channel, subject string, content Content) error {
	if channel == ChannelEmail && recipient.Email == "" && recipient.UserID == nil {
		return fmt.Errorf("email channel requires recipient email or user ID")
	}
	if channel == ChannelSMS && recipient.Phone == "" {
		return fmt.Errorf("SMS channel requires recipient phone")
	}
	if channel == ChannelPush && recipient.DeviceToken == "" {
		return fmt.Errorf("push channel requires device token")
	}
	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if content.Text == "" && content.HTML == "" {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}

// ID returns the notification's unique identifier.
func (n *Notification) ID() NotificationID {
	return n.id
}

// Recipient returns the notification recipient information.
func (n *Notification) Recipient() Recipient {
	return n.recipient
}

// Channel returns the notification delivery channel.
func (n *Notification) Channel() Channel {
	return n.channel
}

// Subject returns the notification subject line.
func (n *Notification) Subject() string {
	return n.subject
}

// Content returns the notification content.
func (n *Notification) Content() Content {
	return n.content
}

// Status returns the current notification status.
func (n *Notification) Status() Status {
	return n.status
}

// Priority returns the notification priority level.
func (n *Notification) Priority() Priority {
	return n.priority
}

// ScheduledAt returns the scheduled delivery time, if set.
func (n *Notification) ScheduledAt() *time.Time {
	return n.scheduledAt
}

// SentAt returns when the notification was sent, if applicable.
func (n *Notification) SentAt() *time.Time {
	return n.sentAt
}

// DeliveredAt returns when the notification was delivered, if applicable.
func (n *Notification) DeliveredAt() *time.Time {
	return n.deliveredAt
}

// FailedAt returns when the notification failed, if applicable.
func (n *Notification) FailedAt() *time.Time {
	return n.failedAt
}

// FailureReason returns the error message if the notification failed.
func (n *Notification) FailureReason() string {
	return n.failureReason
}

// Attempts returns the number of delivery attempts made.
func (n *Notification) Attempts() int {
	return n.attempts
}

// MaxAttempts returns the maximum number of delivery attempts allowed.
func (n *Notification) MaxAttempts() int {
	return n.maxAttempts
}

// Metadata returns additional key-value metadata attached to the notification.
func (n *Notification) Metadata() map[string]string {
	return n.metadata
}

// CreatedAt returns when the notification was created.
func (n *Notification) CreatedAt() time.Time {
	return n.createdAt
}

// UpdatedAt returns when the notification was last updated.
func (n *Notification) UpdatedAt() time.Time {
	return n.updatedAt
}

// Queue transitions the notification to queued status.
// Returns an error if the notification is not in pending status.
func (n *Notification) Queue() error {
	if n.status != StatusPending {
		return fmt.Errorf("cannot queue notification with status %s", n.status)
	}
	n.status = StatusQueued
	n.updatedAt = time.Now()
	return nil
}

// StartSending transitions the notification to sending status.
// Returns an error if the notification is not in queued or pending status.
func (n *Notification) StartSending() error {
	if n.status != StatusQueued && n.status != StatusPending {
		return fmt.Errorf("cannot start sending notification with status %s", n.status)
	}
	n.status = StatusSending
	now := time.Now()
	n.sentAt = &now
	n.updatedAt = now
	return nil
}

// MarkSent transitions the notification to sent status.
// Returns an error if the notification is not in sending status.
func (n *Notification) MarkSent() error {
	if n.status != StatusSending {
		return fmt.Errorf("cannot mark as sent notification with status %s", n.status)
	}
	n.status = StatusSent
	n.updatedAt = time.Now()
	return nil
}

// MarkDelivered transitions the notification to delivered status.
// Returns an error if the notification is not in sent status.
func (n *Notification) MarkDelivered() error {
	if n.status != StatusSent {
		return fmt.Errorf("cannot mark as delivered notification with status %s", n.status)
	}
	n.status = StatusDelivered
	now := time.Now()
	n.deliveredAt = &now
	n.updatedAt = now
	return nil
}

// MarkFailed transitions the notification to failed status with the given reason.
func (n *Notification) MarkFailed(reason string) error {
	n.status = StatusFailed
	now := time.Now()
	n.failedAt = &now
	n.failureReason = reason
	n.updatedAt = now
	return nil
}

// CanRetry returns whether the notification can be retried.
func (n *Notification) CanRetry() bool {
	return n.attempts < n.maxAttempts && n.status == StatusSending
}

// IncrementAttempts increments the delivery attempt counter.
func (n *Notification) IncrementAttempts() {
	n.attempts++
	n.updatedAt = time.Now()
}

// ScheduleRetry schedules the notification for retry after the specified delay.
// Returns an error if the notification cannot be retried.
func (n *Notification) ScheduleRetry(delay time.Duration) error {
	if !n.CanRetry() {
		return fmt.Errorf("cannot retry notification with status %s and %d attempts", n.status, n.attempts)
	}
	n.status = StatusPending
	nextRetry := time.Now().Add(delay)
	n.scheduledAt = &nextRetry
	n.updatedAt = time.Now()
	return nil
}

// IsScheduled returns whether the notification is scheduled for future delivery.
func (n *Notification) IsScheduled() bool {
	return n.scheduledAt != nil && n.scheduledAt.After(time.Now())
}

// IsPending returns whether the notification is in pending status.
func (n *Notification) IsPending() bool {
	return n.status == StatusPending
}

// IsQueued returns whether the notification is in queued status.
func (n *Notification) IsQueued() bool {
	return n.status == StatusQueued
}

// NextRetryDelay calculates the delay before the next retry attempt using exponential backoff.
func (n *Notification) NextRetryDelay() time.Duration {
	delays := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		15 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		6 * time.Hour,
	}

	if n.attempts >= len(delays) {
		return delays[len(delays)-1]
	}

	return delays[n.attempts]
}
