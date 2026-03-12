package service

import (
	"context"
	"nip"
)

type userServiceImpl struct {
	userRepo    nip.UserRepository
	timelineSvc nip.TimelineService
	followRepo  nip.FollowRepository
}

// NewUserService creates a new implementation of nip.UserService.
func NewUserService(userRepo nip.UserRepository,
	timeLineSvc nip.TimelineService,
	followRepo nip.FollowRepository) nip.UserService {
	return &userServiceImpl{
		userRepo:    userRepo,
		timelineSvc: timeLineSvc,
		followRepo:  followRepo,
	}
}

// Follow establishes a relationship and triggers timeline synchronization.
func (s *userServiceImpl) Follow(ctx context.Context,
	userId uint, followedUserId uint) error {

	err := s.followRepo.Create(ctx, nip.Follow{
		FollowerId: userId,
		FollowedId: followedUserId,
	})
	if err != nil {
		return err
	}
	// TODO: should be done on another thread/async
	s.timelineSvc.UserFollowedUser(ctx, userId, followedUserId)
	return nil
}
