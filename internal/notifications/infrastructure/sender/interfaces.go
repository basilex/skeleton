// Package sender provides notification sender implementations.
// This package contains interfaces and implementations for sending notifications
// via various channels including email, SMS, push, and in-app notifications.
package sender

import (
	"context"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
)

// EmailSender defines the interface for sending email notifications.
type EmailSender interface {
	Send(ctx context.Context, to, subject, textBody, htmlBody string) error
}

// SMSSender defines the interface for sending SMS notifications.
type SMSSender interface {
	Send(ctx context.Context, to, message string) error
}

// PushSender defines the interface for sending push notifications.
type PushSender interface {
	Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error
}

// InAppSender defines the interface for sending in-app notifications.
type InAppSender interface {
	Send(ctx context.Context, userID identityDomain.UserID, title, body string, data map[string]string) error
}

// CompositeSender routes notifications to the appropriate sender based on channel type.
// It provides a unified interface for sending notifications across multiple channels.
type CompositeSender struct {
	emailSender EmailSender
	smsSender   SMSSender
	pushSender  PushSender
	inAppSender InAppSender
}

// NewCompositeSender creates a new composite sender with the provided channel senders.
// Any sender can be nil if that channel is not supported.
func NewCompositeSender(
	emailSender EmailSender,
	smsSender SMSSender,
	pushSender PushSender,
	inAppSender InAppSender,
) *CompositeSender {
	return &CompositeSender{
		emailSender: emailSender,
		smsSender:   smsSender,
		pushSender:  pushSender,
		inAppSender: inAppSender,
	}
}

// Send routes the notification to the appropriate sender based on the notification channel.
// Returns nil if the channel sender is not configured.
func (s *CompositeSender) Send(ctx context.Context, notification *domain.Notification) error {
	switch notification.Channel() {
	case domain.ChannelEmail:
		if s.emailSender == nil {
			return nil
		}
		return s.emailSender.Send(
			ctx,
			notification.Recipient().Email,
			notification.Subject(),
			notification.Content().Text,
			notification.Content().HTML,
		)
	case domain.ChannelSMS:
		if s.smsSender == nil {
			return nil
		}
		return s.smsSender.Send(ctx, notification.Recipient().Phone, notification.Content().Text)
	case domain.ChannelPush:
		if s.pushSender == nil {
			return nil
		}
		return s.pushSender.Send(
			ctx,
			notification.Recipient().DeviceToken,
			notification.Subject(),
			notification.Content().Text,
			notification.Metadata(),
		)
	case domain.ChannelInApp:
		if s.inAppSender == nil || notification.Recipient().UserID == nil {
			return nil
		}
		return s.inAppSender.Send(
			ctx,
			*notification.Recipient().UserID,
			notification.Subject(),
			notification.Content().Text,
			notification.Metadata(),
		)
	default:
		return nil
	}
}
