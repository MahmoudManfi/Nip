package service

import (
	"context"
	"nip"
	"regexp"
)

var (
	hashtagRegex = regexp.MustCompile(`#(\w+)`)
	mentionRegex = regexp.MustCompile(`@(\w+)`)
)

type tweetServiceImpl struct {
	tweetRepo       nip.TweetRepository
	userRepo        nip.UserRepository
	timelineSvc     nip.TimelineService
	notificationSvc nip.NotificationService
}

// NewTweetService creates a new implementation of nip.TweetService.
func NewTweetService(tweetRepo nip.TweetRepository,
	userRepo nip.UserRepository, timelineSvc nip.TimelineService,
	notificationSvc nip.NotificationService) nip.TweetService {

	return &tweetServiceImpl{
		tweetRepo:       tweetRepo,
		userRepo:        userRepo,
		timelineSvc:     timelineSvc,
		notificationSvc: notificationSvc,
	}
}

// CreateTweet parses content for hashtags and mentions, persists the tweet, and triggers distributions.
func (s *tweetServiceImpl) CreateTweet(ctx context.Context,
	userId uint, content string) (string, error) {

	hashtags := extractTokens(content, hashtagRegex)
	mentionedUserNames := extractTokens(content, mentionRegex)
	mentionedUserIds, err := s.userRepo.ListIDsByNames(ctx, mentionedUserNames)
	if err != nil {
		return "", err
	}

	tweetId, err := s.tweetRepo.Create(ctx, nip.Tweet{UserId: userId,
		Content:          content,
		Hashtags:         hashtags,
		MentionedUserIds: mentionedUserIds})

	if err != nil {
		return "", err
	}

	// TODO: should be done on another thread
	s.notificationSvc.SendPushNotification(
		ctx, mentionedUserIds, "You are mentioned inside a tweet")

	// TODO: should be done on another thread
	s.timelineSvc.NewTweetAdded(ctx, tweetId)

	return tweetId, nil
}

func extractTokens(content string, re *regexp.Regexp) []string {
	matches := re.FindAllStringSubmatch(content, -1)

	var tokens []string
	for _, m := range matches {
		if len(m) > 1 {
			tokens = append(tokens, m[1])
		}
	}

	return tokens
}
