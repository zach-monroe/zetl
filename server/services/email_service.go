package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	appURL   string
}

// NewEmailService creates a new email service from environment variables
func NewEmailService() *EmailService {
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
		appURL:   os.Getenv("APP_URL"),
	}
}

// IsConfigured checks if the email service is properly configured
func (e *EmailService) IsConfigured() bool {
	return e.host != "" && e.port != "" && e.username != "" && e.password != "" && e.from != ""
}

// SendPasswordResetEmail sends a password reset email to the user
func (e *EmailService) SendPasswordResetEmail(toEmail, token string) error {
	if !e.IsConfigured() {
		return fmt.Errorf("email service is not configured")
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", e.appURL, token)

	subject := "Reset Your Zetl Password"
	body := fmt.Sprintf(`Hello,

You requested to reset your password for your Zetl account.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email. Your password won't be changed.

Best,
The Zetl Team`, resetLink)

	return e.sendEmailWithTLS(toEmail, subject, body)
}

// sendEmailWithTLS sends an email using STARTTLS (required for Gmail port 587)
func (e *EmailService) sendEmailWithTLS(toEmail, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", e.host, e.port)

	// Connect to the SMTP server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Send STARTTLS command
	tlsConfig := &tls.Config{
		ServerName: e.host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(e.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		e.from, toEmail, subject, body)

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
