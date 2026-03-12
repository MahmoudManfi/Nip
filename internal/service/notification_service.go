package service

import (
	"context"
	"nip"
)

type notificationServiceImpl struct {
	userRepo nip.UserRepository
	emailSvc nip.EmailService
	phoneSvc nip.PhoneService
}

// NewNotificationService creates a new implementation of nip.NotificationService.
func NewNotificationService(
	userRepo nip.UserRepository,
	emailSvc nip.EmailService,
	phoneSvc nip.PhoneService,
) nip.NotificationService {
	return &notificationServiceImpl{
		userRepo: userRepo,
		emailSvc: emailSvc,
		phoneSvc: phoneSvc,
	}
}

// SendPushNotification resolves user contacts and sends messages via the appropriate channel.
func (s *notificationServiceImpl) SendPushNotification(ctx context.Context,
	userIds []uint, content string) {

	for _, userId := range userIds {
		user, err := s.userRepo.GetByID(ctx, userId)
		if err != nil {
			continue
		}
		if user.IsEmail {
			s.emailSvc.SendPushNotification(ctx, user.EmailOrPhoneNumber, content)
		} else {
			s.phoneSvc.SendPushNotification(ctx, user.EmailOrPhoneNumber, content)
		}
	}
}
