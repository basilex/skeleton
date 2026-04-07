package domain

import (
	"testing"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

func TestNewNotification(t *testing.T) {
	recipient := Recipient{
		Email: "test@example.com",
	}
	content := Content{
		Text: "Hello World",
	}

	notification, err := NewNotification(recipient, ChannelEmail, "Test Subject", content, PriorityNormal)
	require.NoError(t, err)
	require.NotNil(t, notification)
	require.NotEmpty(t, notification.ID())
	require.Equal(t, recipient, notification.Recipient())
	require.Equal(t, ChannelEmail, notification.Channel())
	require.Equal(t, "Test Subject", notification.Subject())
	require.Equal(t, content, notification.Content())
	require.Equal(t, PriorityNormal, notification.Priority())
	require.Equal(t, StatusPending, notification.Status())
	require.Equal(t, 0, notification.Attempts())
	require.Equal(t, 5, notification.MaxAttempts())
	require.False(t, notification.CreatedAt().IsZero())
}

func TestNewNotificationWithScheduledAt(t *testing.T) {
	recipient := Recipient{
		Email: "test@example.com",
	}
	content := Content{
		Text: "Hello World",
	}
	scheduledTime := time.Now().Add(1 * time.Hour)

	notification, err := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithScheduledAt(scheduledTime))
	require.NoError(t, err)
	require.NotNil(t, notification.ScheduledAt())
	require.Equal(t, scheduledTime.Unix(), notification.ScheduledAt().Unix())
}

func TestNewNotificationWithMaxAttempts(t *testing.T) {
	recipient := Recipient{
		Email: "test@example.com",
	}
	content := Content{
		Text: "Hello World",
	}

	notification, err := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithMaxAttempts(10))
	require.NoError(t, err)
	require.Equal(t, 10, notification.MaxAttempts())
}

func TestNewNotificationWithMetadata(t *testing.T) {
	recipient := Recipient{
		Email: "test@example.com",
	}
	content := Content{
		Text: "Hello World",
	}
	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	notification, err := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithMetadata(metadata))
	require.NoError(t, err)
	require.Equal(t, metadata, notification.Metadata())
}

func TestNewNotificationValidation(t *testing.T) {
	tests := []struct {
		name       string
		recipient  Recipient
		channel    Channel
		subject    string
		content    Content
		wantErr    bool
		errContain string
	}{
		{
			name:       "empty email for email channel",
			recipient:  Recipient{},
			channel:    ChannelEmail,
			subject:    "Test",
			content:    Content{Text: "Body"},
			wantErr:    true,
			errContain: "email channel requires recipient email",
		},
		{
			name:       "empty phone for SMS channel",
			recipient:  Recipient{},
			channel:    ChannelSMS,
			subject:    "Test",
			content:    Content{Text: "Body"},
			wantErr:    true,
			errContain: "SMS channel requires recipient phone",
		},
		{
			name:       "empty device token for push channel",
			recipient:  Recipient{},
			channel:    ChannelPush,
			subject:    "Test",
			content:    Content{Text: "Body"},
			wantErr:    true,
			errContain: "push channel requires device token",
		},
		{
			name:       "empty subject",
			recipient:  Recipient{Email: "test@example.com"},
			channel:    ChannelEmail,
			subject:    "",
			content:    Content{Text: "Body"},
			wantErr:    true,
			errContain: "subject cannot be empty",
		},
		{
			name:       "empty content",
			recipient:  Recipient{Email: "test@example.com"},
			channel:    ChannelEmail,
			subject:    "Test",
			content:    Content{},
			wantErr:    true,
			errContain: "content cannot be empty",
		},
		{
			name:      "valid email notification",
			recipient: Recipient{Email: "test@example.com"},
			channel:   ChannelEmail,
			subject:   "Test",
			content:   Content{Text: "Body"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNotification(tt.recipient, tt.channel, tt.subject, tt.content, PriorityNormal)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNotificationQueue(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)

	err := notification.Queue()
	require.NoError(t, err)
	require.Equal(t, StatusQueued, notification.Status())

	// Cannot queue from wrong status
	err = notification.Queue()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot queue notification")
}

func TestNotificationStartSending(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	t.Run("from pending", func(t *testing.T) {
		notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
		err := notification.StartSending()
		require.NoError(t, err)
		require.Equal(t, StatusSending, notification.Status())
		require.NotNil(t, notification.SentAt())
	})

	t.Run("from queued", func(t *testing.T) {
		notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
		_ = notification.Queue()
		err := notification.StartSending()
		require.NoError(t, err)
		require.Equal(t, StatusSending, notification.Status())
	})

	t.Run("from wrong status", func(t *testing.T) {
		notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
		_ = notification.MarkFailed("error")
		err := notification.StartSending()
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot start sending notification")
	})
}

func TestNotificationMarkSent(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	_ = notification.StartSending()

	err := notification.MarkSent()
	require.NoError(t, err)
	require.Equal(t, StatusSent, notification.Status())

	// Cannot mark sent from wrong status
	err = notification.MarkSent()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot mark as sent notification")
}

func TestNotificationMarkDelivered(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	_ = notification.StartSending()
	_ = notification.MarkSent()

	err := notification.MarkDelivered()
	require.NoError(t, err)
	require.Equal(t, StatusDelivered, notification.Status())
	require.NotNil(t, notification.DeliveredAt())

	// Cannot mark delivered from wrong status
	err = notification.MarkDelivered()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot mark as delivered notification")
}

func TestNotificationMarkFailed(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)

	err := notification.MarkFailed("SMTP error")
	require.NoError(t, err)
	require.Equal(t, StatusFailed, notification.Status())
	require.NotNil(t, notification.FailedAt())
	require.Equal(t, "SMTP error", notification.FailureReason())
}

func TestNotificationCanRetry(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithMaxAttempts(3))

	// Cannot retry in Pending status
	require.False(t, notification.CanRetry())

	_ = notification.StartSending()
	// Can retry in Sending status if attempts < maxAttempts
	require.True(t, notification.CanRetry())

	notification.IncrementAttempts()
	// Still can retry (attempts=1 < 3)
	require.True(t, notification.CanRetry())

	notification.IncrementAttempts()
	// Still can retry (attempts=2 < 3)
	require.True(t, notification.CanRetry())

	notification.IncrementAttempts()
	// Cannot retry (attempts=3 >= 3)
	require.False(t, notification.CanRetry())
}

func TestNotificationIncrementAttempts(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)

	require.Equal(t, 0, notification.Attempts())

	notification.IncrementAttempts()
	require.Equal(t, 1, notification.Attempts())

	notification.IncrementAttempts()
	require.Equal(t, 2, notification.Attempts())
}

func TestNotificationScheduleRetry(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithMaxAttempts(3))
	_ = notification.StartSending()
	notification.IncrementAttempts()

	delay := 5 * time.Minute
	err := notification.ScheduleRetry(delay)
	require.NoError(t, err)
	require.Equal(t, StatusPending, notification.Status())
	require.NotNil(t, notification.ScheduledAt())

	// Cannot schedule retry when CanRetry is false
	notification.IncrementAttempts()
	notification.IncrementAttempts() // attempts = 3 now

	err = notification.ScheduleRetry(delay)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot retry notification")
}

func TestNotificationNextRetryDelay(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)

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

	// First attempt (attempts=0)
	require.Equal(t, delays[0], notification.NextRetryDelay())

	notification.IncrementAttempts()
	require.Equal(t, delays[1], notification.NextRetryDelay())

	notification.IncrementAttempts()
	require.Equal(t, delays[2], notification.NextRetryDelay())

	// Beyond limit returns last delay
	for i := 0; i < 10; i++ {
		notification.IncrementAttempts()
	}
	require.Equal(t, delays[len(delays)-1], notification.NextRetryDelay())
}

func TestNotificationIsScheduled(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	require.False(t, notification.IsScheduled())

	scheduledTime := time.Now().Add(1 * time.Hour)
	notification2, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal, WithScheduledAt(scheduledTime))
	require.True(t, notification2.IsScheduled())
}

func TestNotificationIsPending(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	require.True(t, notification.IsPending())

	_ = notification.Queue()
	require.False(t, notification.IsPending())
}

func TestNotificationIsQueued(t *testing.T) {
	recipient := Recipient{Email: "test@example.com"}
	content := Content{Text: "Body"}

	notification, _ := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	_ = notification.Queue()
	require.True(t, notification.IsQueued())
}

func TestParseNotificationID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid ID", input: "019d65d6-de90-7200-b1cf-4f8745597e0a", wantErr: false},
		{name: "empty ID", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseNotificationID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, NotificationID(tt.input), id)
			}
		})
	}
}

func TestParseChannel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Channel
		wantErr bool
	}{
		{name: "email", input: "email", want: ChannelEmail, wantErr: false},
		{name: "sms", input: "sms", want: ChannelSMS, wantErr: false},
		{name: "push", input: "push", want: ChannelPush, wantErr: false},
		{name: "in_app", input: "in_app", want: ChannelInApp, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := ParseChannel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, channel)
			}
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Status
		wantErr bool
	}{
		{name: "pending", input: "pending", want: StatusPending, wantErr: false},
		{name: "queued", input: "queued", want: StatusQueued, wantErr: false},
		{name: "sending", input: "sending", want: StatusSending, wantErr: false},
		{name: "sent", input: "sent", want: StatusSent, wantErr: false},
		{name: "delivered", input: "delivered", want: StatusDelivered, wantErr: false},
		{name: "failed", input: "failed", want: StatusFailed, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParseStatus(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, status)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Priority
		wantErr bool
	}{
		{name: "low", input: "low", want: PriorityLow, wantErr: false},
		{name: "normal", input: "normal", want: PriorityNormal, wantErr: false},
		{name: "high", input: "high", want: PriorityHigh, wantErr: false},
		{name: "critical", input: "critical", want: PriorityCritical, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority, err := ParsePriority(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, priority)
			}
		})
	}
}

func TestNotificationRecipientWithUserID(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	recipient := Recipient{
		UserID: &userID,
		Email:  "test@example.com",
	}

	content := Content{Text: "Body"}
	notification, err := NewNotification(recipient, ChannelEmail, "Test", content, PriorityNormal)
	require.NoError(t, err)
	require.NotNil(t, notification.Recipient().UserID)
	require.Equal(t, userID, *notification.Recipient().UserID)
}

func TestChannelString(t *testing.T) {
	require.Equal(t, "email", ChannelEmail.String())
	require.Equal(t, "sms", ChannelSMS.String())
	require.Equal(t, "push", ChannelPush.String())
	require.Equal(t, "in_app", ChannelInApp.String())
}

func TestStatusString(t *testing.T) {
	require.Equal(t, "pending", StatusPending.String())
	require.Equal(t, "failed", StatusFailed.String())
}

func TestPriorityString(t *testing.T) {
	require.Equal(t, "low", PriorityLow.String())
	require.Equal(t, "critical", PriorityCritical.String())
}
