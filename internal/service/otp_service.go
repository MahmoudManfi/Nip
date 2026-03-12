package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"nip"
	"sync"
	"time"
)

type otpEntry struct {
	code  int
	timer *time.Timer
}

type otpServiceImpl struct {
	sync.Mutex
	otpStore map[uint]otpEntry
}

// NewOtpService creates a new in-memory implementation of nip.OtpService.
func NewOtpService() nip.OtpService {
	return &otpServiceImpl{
		otpStore: make(map[uint]otpEntry),
	}
}

// GenerateOtp creates a 6-digit OTP and schedules its deletion after 5 minutes.
func (s *otpServiceImpl) GenerateOtp(ctx context.Context, userId uint) (int, error) {
	const (
		minCode   = 100000
		codeRange = 900000
		ttl       = 5 * time.Minute
	)

	n, err := rand.Int(rand.Reader, big.NewInt(codeRange))
	if err != nil {
		return 0, err
	}
	code := int(n.Int64()) + minCode

	timer := time.AfterFunc(ttl, func() {
		s.Lock()
		delete(s.otpStore, userId)
		defer s.Unlock()
	})

	entry := otpEntry{
		code:  code,
		timer: timer,
	}

	s.Lock()
	s.otpStore[userId] = entry
	s.Unlock()

	return code, nil
}

// ValidateOtpCode verifies the OTP and immediately expires it upon successful match.
func (s *otpServiceImpl) ValidateOtpCode(ctx context.Context, userId uint, otpCode int) error {
	s.Lock()
	defer s.Unlock()

	entry, ok := s.otpStore[userId]
	if !ok {
		return errors.New("OTP not found or already used")
	}

	if entry.code != otpCode {
		return errors.New("invalid OTP code")
	}

	entry.timer.Stop()
	delete(s.otpStore, userId)
	return nil
}
