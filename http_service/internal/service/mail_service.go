/*
 * Email delivery provider.
 * 1. Define mail sender boundary for auth verification emails.
 * 2. Provide mock and SMTP implementations.
 * 3. Build plain text RFC822 messages for outbound SMTP delivery.
 */
package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"

	"ajoliving_web/http_service/internal/config"
)

// 1. MailSender sends plain text emails.
type MailSender interface {
	Send(ctx context.Context, recipient string, subject string, body string) error
}

// 2. MockMailSender accepts email delivery without external side effects.
type MockMailSender struct{}

// 3. SMTPMailSender sends emails through SMTP.
type SMTPMailSender struct {
	config *config.Config
}

// 4. NewMailSender returns the configured email sender.
func NewMailSender(cfg *config.Config) MailSender {
	if !cfg.MailEnabled {
		return &MockMailSender{}
	}

	return &SMTPMailSender{config: cfg}
}

// 5. Send accepts mock email delivery.
func (s *MockMailSender) Send(context.Context, string, string, string) error {
	return nil
}

// 6. Send sends one plain text email through SMTP.
func (s *SMTPMailSender) Send(_ context.Context, recipient string, subject string, body string) error {
	if strings.TrimSpace(recipient) == "" {
		return nil
	}
	if err := validateSMTPConfig(s.config); err != nil {
		return err
	}

	message := buildEmailMessage(s.config.SMTPFrom, recipient, subject, body)
	conn, err := smtp.Dial(s.config.SMTPAddress())
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer conn.Close()

	if strings.TrimSpace(s.config.SMTPHelloName) != "" {
		if err := conn.Hello(s.config.SMTPHelloName); err != nil {
			return fmt.Errorf("smtp hello: %w", err)
		}
	}

	if ok, _ := conn.Extension("STARTTLS"); ok && !isLocalSMTPHost(s.config.SMTPHost) {
		tlsConfig := &tls.Config{ServerName: tlsServerName(s.config.SMTPHost)}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if strings.TrimSpace(s.config.SMTPUsername) != "" {
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, tlsServerName(s.config.SMTPHost))
		if err := conn.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := conn.Mail(s.config.SMTPFrom); err != nil {
		return fmt.Errorf("set mail from: %w", err)
	}
	if err := conn.Rcpt(recipient); err != nil {
		return fmt.Errorf("set rcpt to: %w", err)
	}

	writer, err := conn.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("write email body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close email body: %w", err)
	}
	if err := conn.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}

	return nil
}

// 7. validateSMTPConfig checks required SMTP fields.
func validateSMTPConfig(cfg *config.Config) error {
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return fmt.Errorf("SMTP_HOST is required when MAIL_ENABLED=true")
	}
	if cfg.SMTPPort <= 0 {
		return fmt.Errorf("SMTP_PORT must be greater than zero")
	}
	if strings.TrimSpace(cfg.SMTPFrom) == "" {
		return fmt.Errorf("SMTP_FROM is required when MAIL_ENABLED=true")
	}

	return nil
}

// 8. buildEmailMessage creates the RFC822 plain text message.
func buildEmailMessage(from string, recipient string, subject string, body string) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject)))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(body)
	return buffer.Bytes()
}

// 9. isLocalSMTPHost checks whether STARTTLS should be skipped for local relay.
func isLocalSMTPHost(host string) bool {
	value := strings.TrimSpace(host)
	return value == "127.0.0.1" || value == "localhost"
}

// 10. tlsServerName extracts the hostname used by TLS and SMTP auth.
func tlsServerName(host string) string {
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}

	return host
}
