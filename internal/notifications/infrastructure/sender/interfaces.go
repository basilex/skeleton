package sender

import (
	"context"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, textBody, htmlBody string) error
}

type SMSSender interface {
	Send(ctx context.Context, to, message string) error
}

type PushSender interface {
	Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error
}

type InAppSender interface {
	Send(ctx context.Context, userID identityDomain.UserID, title, body string, data map[string]string) error
}

type CompositeSender struct {
	emailSender EmailSender
	smsSender   SMSSender
	pushSender  PushSender
	inAppSender InAppSender
}

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
