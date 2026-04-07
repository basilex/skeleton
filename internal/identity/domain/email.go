package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

type Email string

func NewEmail(email string) (Email, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" {
		return "", fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(e) {
		return "", fmt.Errorf("email: must be a valid email address")
	}
	return Email(e), nil
}

func (e Email) String() string {
	return string(e)
}
