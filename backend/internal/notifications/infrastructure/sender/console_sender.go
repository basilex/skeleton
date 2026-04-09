// Package sender provides notification sender implementations.
// This package contains interfaces and implementations for sending notifications
// via various channels including email, SMS, push, and in-app notifications.
package sender

import (
	"context"
	"fmt"
)

// ConsoleEmailSender is a debug sender that prints emails to stdout.
// Use for development and testing only; not suitable for production.
type ConsoleEmailSender struct{}

// NewConsoleEmailSender creates a new console email sender.
func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{}
}

// Send prints the email to stdout in a formatted manner for debugging.
func (s *ConsoleEmailSender) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	fmt.Printf("\n========== EMAIL ==========\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("----------------------------\n")
	fmt.Printf("%s\n", textBody)
	if htmlBody != "" {
		fmt.Printf("----------------------------\n")
		fmt.Printf("HTML:\n%s\n", htmlBody)
	}
	fmt.Printf("============================\n\n")
	return nil
}

type ConsoleSMSSender struct{}

// NewConsoleSMSSender creates a new console SMS sender.
func NewConsoleSMSSender() *ConsoleSMSSender {
	return &ConsoleSMSSender{}
}

// Send prints the SMS to stdout for debugging purposes.
func (s *ConsoleSMSSender) Send(ctx context.Context, to, message string) error {
	fmt.Printf("\n========== SMS ==========\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("--------------------------\n")
	fmt.Printf("%s\n", message)
	fmt.Printf("==========================\n\n")
	return nil
}
