package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// EmailService handles sending emails via SMTP.
type EmailService struct {
	host     string
	port     string
	username string
	password string
	logger   *zap.Logger
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(host, port, username, password string, logger *zap.Logger) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		logger:   logger,
	}
}

// SendEmail sends an HTML email to the specified recipients.
func (s *EmailService) SendEmail(to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	addr := net.JoinHostPort(s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.username, strings.Join(to, ","), subject, body,
	))

	// QQ SMTP on port 465 requires implicit TLS (direct SSL connection).
	tlsConfig := &tls.Config{ServerName: s.host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS to %s: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err = client.Mail(s.username); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed for %s: %w", recipient, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP data close failed: %w", err)
	}

	s.logger.Info("email sent successfully",
		zap.Strings("to", to),
		zap.String("subject", subject),
	)
	return client.Quit()
}
