package service

import (
	"context"
	"fmt"
	"nip"
	"strings"
)

type authServiceImpl struct {
	userRepo nip.UserRepository
	otpSvc   nip.OtpService
	emailSvc nip.EmailService
	phoneSvc nip.PhoneService
}

// NewAuthService creates a new implementation of nip.AuthService.
func NewAuthService(
	userRepo nip.UserRepository,
	otpSvc nip.OtpService,
	emailSvc nip.EmailService,
	phoneSvc nip.PhoneService,
) nip.AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
		otpSvc:   otpSvc,
		emailSvc: emailSvc,
		phoneSvc: phoneSvc,
	}
}

// SignUp handles new user creation, including input validation and formatting.
func (s *authServiceImpl) SignUp(ctx context.Context,
	userName string, emailOrPhoneNumber string) (uint, error) {

	userName = strings.TrimSpace(userName)
	emailOrPhoneNumber = strings.TrimSpace(emailOrPhoneNumber)
	if userName == "" {
		return 0, fmt.Errorf("userName is required")
	}

	isEmail, err := s.isEmailOrPhoneNumber(ctx, emailOrPhoneNumber)
	if err != nil {
		return 0, fmt.Errorf("parsing email or phone number: %w", err)
	}

	user := nip.User{
		UserName:           userName,
		EmailOrPhoneNumber: emailOrPhoneNumber,
		IsEmail:            isEmail,
	}
	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
	}

	return id, nil
}

// SendOtp retrieves the user and dispatches an OTP via Email or SMS.
func (s *authServiceImpl) SendOtp(ctx context.Context,
	emailOrPhoneNumber string) (int, error) {

	user, err := s.userRepo.GetByEmailOrPhone(ctx, emailOrPhoneNumber)
	if err != nil {
		return 0, fmt.Errorf("getting user: %w", err)
	}
	otpCode, err := s.otpSvc.GenerateOtp(ctx, user.Id)
	if err != nil {
		return 0, fmt.Errorf("generating OTP: %w", err)
	}
	if user.IsEmail {
		err = s.emailSvc.SendOtpEmail(ctx, emailOrPhoneNumber, otpCode)
	} else {
		err = s.phoneSvc.SendOtpSms(ctx, emailOrPhoneNumber, otpCode)
	}
	return otpCode, err
}

// Login validates the user's OTP and returns their full profile.
func (s *authServiceImpl) Login(ctx context.Context,
	emailOrPhoneNumber string, otpCode int) (nip.User, error) {

	user, err := s.userRepo.GetByEmailOrPhone(ctx, emailOrPhoneNumber)
	if err != nil {
		return nip.User{}, fmt.Errorf("getting user: %w", err)
	}
	err = s.otpSvc.ValidateOtpCode(ctx, user.Id, otpCode)
	if err != nil {
		return nip.User{}, fmt.Errorf("validating OTP code: %w", err)
	}
	return user, nil
}

func (s *authServiceImpl) isEmailOrPhoneNumber(ctx context.Context, emailOrPhoneNumber string) (bool, error) {
	err := s.emailSvc.ValidateEmail(ctx, emailOrPhoneNumber)
	if err == nil {
		return true, nil
	}
	err = s.phoneSvc.ValidatePhoneNumber(ctx, emailOrPhoneNumber)
	if err == nil {
		return false, nil
	}
	return false, fmt.Errorf("invalid email or phone number")
}
