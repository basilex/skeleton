// Package domain provides domain entities and value objects for the notifications module.
// This package contains the core business logic types for notification management,
// including preferences, templates, and domain events.
package domain

import (
	"fmt"
)

// Notifications domain error definitions.
var (
	ErrNotificationNotFound      = fmt.Errorf("notification not found")
	ErrTemplateNotFound          = fmt.Errorf("template not found")
	ErrPreferencesNotFound       = fmt.Errorf("preferences not found")
	ErrInvalidChannel            = fmt.Errorf("invalid channel")
	ErrInvalidStatus             = fmt.Errorf("invalid status")
	ErrInvalidPriority           = fmt.Errorf("invalid priority")
	ErrNotificationAlreadySent   = fmt.Errorf("notification already sent")
	ErrNotificationAlreadyFailed = fmt.Errorf("notification already failed")
	ErrMaxAttemptsExceeded       = fmt.Errorf("max attempts exceeded")
	ErrTemplateVariableMissing   = fmt.Errorf("template variable missing")
	ErrRecipientRequired         = fmt.Errorf("recipient required")
	ErrContentRequired           = fmt.Errorf("content required")
	ErrQuietHoursInvalid         = fmt.Errorf("invalid quiet hours configuration")
	ErrChannelDisabled           = fmt.Errorf("channel disabled for user")
	errCannotSendInQuietHours    = fmt.Errorf("cannot send during quiet hours")
)

// NewErrChannelDisabled creates an error indicating the specified channel is disabled.
func NewErrChannelDisabled(channel Channel) error {
	return fmt.Errorf("%w: %s", ErrChannelDisabled, channel)
}

// NewErrTemplateVariableMissing creates an error indicating a required template variable is missing.
func NewErrTemplateVariableMissing(variable string) error {
	return fmt.Errorf("%w: %s", ErrTemplateVariableMissing, variable)
}

// NewErrMaxAttemptsExceeded creates an error indicating the maximum retry attempts have been exceeded.
func NewErrMaxAttemptsExceeded(attempts, maxAttempts int) error {
	return fmt.Errorf("%w: attempts=%d, max=%d", ErrMaxAttemptsExceeded, attempts, maxAttempts)
}
