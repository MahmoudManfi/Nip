package service

import (
	"context"
	"fmt"
	"nip"
)

type phoneServiceImpl struct{}

// NewPhoneService creates a new implementation of nip.PhoneService.
func NewPhoneService() nip.PhoneService {
	return &phoneServiceImpl{}
}

// ValidatePhoneNumber is a placeholder for phone number validation.
func (s *phoneServiceImpl) ValidatePhoneNumber(ctx context.Context, phoneNumber string) error {
	return nil
}

// SendOtpSms simulates sending an OTP via SMS.
func (s *phoneServiceImpl) SendOtpSms(ctx context.Context, phoneNumber string, otpCode int) error {
	fmt.Printf("[SMS] Sending OTP code to %s\n", phoneNumber)
	return nil
}

// SendPushNotification simulates sending a push notification via SMS.
func (s *phoneServiceImpl) SendPushNotification(ctx context.Context, phoneNumber string,
	content string) error {

	fmt.Printf("[SMS] Notification for %s: %s\n", phoneNumber, content)
	return nil
}
