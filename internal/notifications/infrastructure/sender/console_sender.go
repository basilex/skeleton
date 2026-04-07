package sender

import (
	"context"
	"fmt"
)

type ConsoleEmailSender struct{}

func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{}
}

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

func NewConsoleSMSSender() *ConsoleSMSSender {
	return &ConsoleSMSSender{}
}

func (s *ConsoleSMSSender) Send(ctx context.Context, to, message string) error {
	fmt.Printf("\n========== SMS ==========\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("--------------------------\n")
	fmt.Printf("%s\n", message)
	fmt.Printf("==========================\n\n")
	return nil
}
