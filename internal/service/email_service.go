package service

import (
	"context"
	"fmt"
	"net/mail"
	"nip"
)

type emailServiceImpl struct{}

// NewEmailService creates a new implementation of nip.EmailService.
func NewEmailService() nip.EmailService {
	return &emailServiceImpl{}
}

// ValidateEmail checks if the string follows a standard email format.
func (s *emailServiceImpl) ValidateEmail(ctx context.Context, email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// SendOtpEmail simulates sending an OTP email.
func (s *emailServiceImpl) SendOtpEmail(ctx context.Context, email string, otpCode int) error {
	fmt.Printf("[Email] Sending OTP code to %s\n", email)
	return nil
}

// SendPushNotification simulates sending a push notification via email.
func (s *emailServiceImpl) SendPushNotification(ctx context.Context, email string, content string) error {
	fmt.Printf("[Email] Notification for %s: %s\n", email, content)
	return nil
}
