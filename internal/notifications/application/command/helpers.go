// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"time"
)

// parseDuration parses a duration string pointer and returns the duration.
// Returns zero duration if the pointer is nil or the string is empty or invalid.
func parseDuration(s *string) time.Duration {
	if s == nil || *s == "" {
		return 0
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return 0
	}
	return d
}
