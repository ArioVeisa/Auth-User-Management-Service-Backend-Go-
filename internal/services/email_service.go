package services

import (
	"fmt"

	"github.com/auth-service/internal/config"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	cfg    *config.Config
	dialer *gomail.Dialer
}

func NewEmailService(cfg *config.Config) *EmailService {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)
	return &EmailService{
		cfg:    cfg,
		dialer: dialer,
	}
}

func (s *EmailService) SendVerificationEmail(to, displayName, token string) error {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Thank you for registering. Please verify your email address by using the following token:</p>
		<p><strong>Token: %s</strong></p>
		<p>Or click the link below:</p>
		<a href="http://localhost:8080/api/v1/auth/verify-email?token=%s">Verify Email</a>
		<p>This token will expire in 1 hour.</p>
		<p>If you did not create an account, please ignore this email.</p>
	`, displayName, token, token)

	return s.sendEmail(to, subject, body)
}

func (s *EmailService) SendPasswordResetEmail(to, displayName, token string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>You have requested to reset your password. Use the following token:</p>
		<p><strong>Token: %s</strong></p>
		<p>This token will expire in 1 hour.</p>
		<p>If you did not request a password reset, please ignore this email.</p>
	`, displayName, token)

	return s.sendEmail(to, subject, body)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	if s.cfg.SMTPUser == "" {
		fmt.Printf("[EMAIL] To: %s, Subject: %s\n", to, subject)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SMTPUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return s.dialer.DialAndSend(m)
}
