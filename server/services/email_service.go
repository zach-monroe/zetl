package services

import (
	"fmt"
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

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.from, toEmail, subject, body)

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	addr := fmt.Sprintf("%s:%s", e.host, e.port)

	err := smtp.SendMail(addr, auth, e.from, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
