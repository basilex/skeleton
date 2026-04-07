package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

type NotificationID string

func NewNotificationID() NotificationID {
	return NotificationID(uuid.NewV7().String())
}

func ParseNotificationID(s string) (NotificationID, error) {
	if s == "" {
		return "", fmt.Errorf("notification ID cannot be empty")
	}
	return NotificationID(s), nil
}

func (id NotificationID) String() string {
	return string(id)
}

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
	ChannelInApp Channel = "in_app"
)

func (c Channel) String() string {
	return string(c)
}

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

type Status string

const (
	StatusPending   Status = "pending"
	StatusQueued    Status = "queued"
	StatusSending   Status = "sending"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

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

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

func (p Priority) String() string {
	return string(p)
}

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

type Recipient struct {
	UserID      *domain.UserID
	Email       string
	Phone       string
	DeviceToken string
}

type Content struct {
	Text string
	HTML string
}

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

type NotificationOption func(*Notification)

func WithScheduledAt(t time.Time) NotificationOption {
	return func(n *Notification) {
		n.scheduledAt = &t
	}
}

func WithMaxAttempts(max int) NotificationOption {
	return func(n *Notification) {
		n.maxAttempts = max
	}
}

func WithMetadata(metadata map[string]string) NotificationOption {
	return func(n *Notification) {
		for k, v := range metadata {
			n.metadata[k] = v
		}
	}
}

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

func (n *Notification) ID() NotificationID {
	return n.id
}

func (n *Notification) Recipient() Recipient {
	return n.recipient
}

func (n *Notification) Channel() Channel {
	return n.channel
}

func (n *Notification) Subject() string {
	return n.subject
}

func (n *Notification) Content() Content {
	return n.content
}

func (n *Notification) Status() Status {
	return n.status
}

func (n *Notification) Priority() Priority {
	return n.priority
}

func (n *Notification) ScheduledAt() *time.Time {
	return n.scheduledAt
}

func (n *Notification) SentAt() *time.Time {
	return n.sentAt
}

func (n *Notification) DeliveredAt() *time.Time {
	return n.deliveredAt
}

func (n *Notification) FailedAt() *time.Time {
	return n.failedAt
}

func (n *Notification) FailureReason() string {
	return n.failureReason
}

func (n *Notification) Attempts() int {
	return n.attempts
}

func (n *Notification) MaxAttempts() int {
	return n.maxAttempts
}

func (n *Notification) Metadata() map[string]string {
	return n.metadata
}

func (n *Notification) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Notification) UpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Notification) Queue() error {
	if n.status != StatusPending {
		return fmt.Errorf("cannot queue notification with status %s", n.status)
	}
	n.status = StatusQueued
	n.updatedAt = time.Now()
	return nil
}

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

func (n *Notification) MarkSent() error {
	if n.status != StatusSending {
		return fmt.Errorf("cannot mark as sent notification with status %s", n.status)
	}
	n.status = StatusSent
	n.updatedAt = time.Now()
	return nil
}

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

func (n *Notification) MarkFailed(reason string) error {
	n.status = StatusFailed
	now := time.Now()
	n.failedAt = &now
	n.failureReason = reason
	n.updatedAt = now
	return nil
}

func (n *Notification) CanRetry() bool {
	return n.attempts < n.maxAttempts && n.status == StatusSending
}

func (n *Notification) IncrementAttempts() {
	n.attempts++
	n.updatedAt = time.Now()
}

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

func (n *Notification) IsScheduled() bool {
	return n.scheduledAt != nil && n.scheduledAt.After(time.Now())
}

func (n *Notification) IsPending() bool {
	return n.status == StatusPending
}

func (n *Notification) IsQueued() bool {
	return n.status == StatusQueued
}

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
