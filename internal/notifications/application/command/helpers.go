package command

import (
	"time"
)

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
