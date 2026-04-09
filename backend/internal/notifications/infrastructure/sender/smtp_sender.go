package sender

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// SMTPConfig contains configuration for the SMTP email sender.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// SMTPSender sends emails via SMTP protocol.
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTP sender with the provided configuration.
func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{config: config}
}

// Send sends an email via SMTP with optional HTML body.
// Supports both plain text and multipart/alternative messages.
func (s *SMTPSender) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"

	if htmlBody != "" {
		headers["Content-Type"] = "multipart/alternative; boundary=\"boundary123\""
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n"

	if htmlBody != "" {
		message += "--boundary123\r\n"
		message += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
		message += textBody + "\r\n\r\n"
		message += "--boundary123\r\n"
		message += "Content-Type: text/html; charset=UTF-8\r\n\r\n"
		message += htmlBody + "\r\n\r\n"
		message += "--boundary123--\r\n"
	} else {
		message += textBody
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, to, []byte(message))
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message))
}

// sendWithTLS sends an email using STARTTLS for encrypted communication.
func (s *SMTPSender) sendWithTLS(addr string, auth smtp.Auth, to string, message []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start tls: %w", err)
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}
	}

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return client.Quit()
}
