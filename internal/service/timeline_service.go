package service

import (
	"container/list"
	"context"
	"nip"
	"slices"
	"sync"
)

// UserTimeline stores a thread-safe user's timeline
type UserTimeline struct {
	sync.RWMutex
	mentionedTweetIdSet map[string]struct{}
	followedTweetIdSet  map[string]struct{}
}

// GlobalTweets is a thread-safe list of the tweets with "NipNip" hashtag which will be shown to all users.
// The list is ordered by the time the tweet was created.
type GlobalTweets struct {
	sync.RWMutex
	list.List
}

type timelineServiceImpl struct {
	sync.Mutex
	userIdToTimelineMap map[uint]*UserTimeline
	tweetRepo           nip.TweetRepository
	userRepo            nip.UserRepository
	followRepo          nip.FollowRepository
	globalTweets        GlobalTweets
}

// NewTimelineService creates a new implementation of nip.TimelineService and initializes it with existing data.
func NewTimelineService(ctx context.Context,
	tweetRepo nip.TweetRepository,
	userRepo nip.UserRepository, followRepo nip.FollowRepository) nip.TimelineService {

	timelineService := &timelineServiceImpl{
		userIdToTimelineMap: make(map[uint]*UserTimeline),
		tweetRepo:           tweetRepo,
		userRepo:            userRepo,
		followRepo:          followRepo,
	}

	timelineService.initTimeline(ctx)
	return timelineService
}

func (s *timelineServiceImpl) initTimeline(ctx context.Context) error {
	allTweets, err := s.tweetRepo.ListAllLatest(ctx)
	if err != nil {
		return err
	}

	for _, tweet := range allTweets {
		s.newTweetAdded(ctx, tweet)
	}

	return nil
}

// GetTimeline returns tweets for a user which includes
// reverse chronologically ordered global tweets,
// reverse chronologically ordered mentioned tweets,
// and reverse chronologically ordered followed tweets.
func (s *timelineServiceImpl) GetTimeline(
	ctx context.Context, userId uint) ([]nip.Tweet, error) {

	resultTweets := s.getGlobalTweets()

	userTimeline := s.getUserTimeline(userId)
	userTimeline.RLock()
	mentionedTweetIds := setToSlice(userTimeline.mentionedTweetIdSet)
	followedTweetIds := setToSlice(userTimeline.followedTweetIdSet)
	userTimeline.RUnlock()

	latestMentionedTweets, err := s.tweetRepo.ListLatest(ctx, mentionedTweetIds)
	if err != nil {
		return nil, err
	}
	resultTweets = append(resultTweets, latestMentionedTweets...)

	latestFollowedTweets, err :=
		s.tweetRepo.ListLatest(ctx, followedTweetIds)
	if err != nil {
		return nil, err
	}
	resultTweets = append(resultTweets, latestFollowedTweets...)

	return resultTweets, nil
}

// NewTweetAdded distributes a new tweet to the global list if the tweet contains "NipNip" hashtag.
// Otherwise it distributes the tweet to the mentioned users' timelines and followers' timelines.
func (s *timelineServiceImpl) NewTweetAdded(
	ctx context.Context, tweetId string) error {

	tweet, err := s.tweetRepo.GetByID(ctx, tweetId)
	if err != nil {
		return err
	}
	s.newTweetAdded(ctx, tweet)

	return nil
}

// UserFollowedUser updates the timeline of the follower user with the tweets of the followed user
// if these tweets are not already in the timeline.
func (s *timelineServiceImpl) UserFollowedUser(
	ctx context.Context, followerId uint, followedId uint) error {

	followedTweetIds, err := s.tweetRepo.ListIDsByUserID(ctx, followedId)

	if err != nil {
		return err
	}

	userTimeline := s.getUserTimeline(followerId)
	userTimeline.Lock()
	defer userTimeline.Unlock()
	for _, tweetId := range followedTweetIds {
		if _, ok := userTimeline.mentionedTweetIdSet[tweetId]; !ok {
			userTimeline.followedTweetIdSet[tweetId] = struct{}{}
		}
	}
	return nil
}

func (s *timelineServiceImpl) getUserTimeline(userId uint) *UserTimeline {
	s.Lock()
	defer s.Unlock()
	userTimeline, ok := s.userIdToTimelineMap[userId]
	if !ok {
		userTimeline = &UserTimeline{mentionedTweetIdSet: make(map[string]struct{}),
			followedTweetIdSet: make(map[string]struct{})}
		s.userIdToTimelineMap[userId] = userTimeline
	}
	return userTimeline
}

func (s *timelineServiceImpl) updateMentionedUsersTimeline(
	mentionedUserIds []uint, tweetId string) {

	for _, mentionedUserId := range mentionedUserIds {
		userTimeline := s.getUserTimeline(mentionedUserId)
		userTimeline.Lock()
		userTimeline.mentionedTweetIdSet[tweetId] = struct{}{}
		userTimeline.Unlock()
	}
}

func (s *timelineServiceImpl) updateFollowersTimeline(
	ctx context.Context, userId uint, tweetId string) error {

	followerIds, err := s.followRepo.ListFollowerIds(ctx, userId)
	if err != nil {
		return err
	}
	for _, followerId := range followerIds {
		userTimeline := s.getUserTimeline(followerId)
		userTimeline.Lock()
		if _, ok := userTimeline.mentionedTweetIdSet[tweetId]; !ok {
			userTimeline.followedTweetIdSet[tweetId] = struct{}{}
		}
		userTimeline.Unlock()
	}
	return nil
}

func (s *timelineServiceImpl) getGlobalTweets() []nip.Tweet {
	resultTweets := []nip.Tweet{}
	s.globalTweets.RLock()
	defer s.globalTweets.RUnlock()

	for e := s.globalTweets.Front(); e != nil; e = e.Next() {
		if tweet, ok := e.Value.(nip.Tweet); ok {
			resultTweets = append(resultTweets, tweet)
		}
	}

	return resultTweets
}

func (s *timelineServiceImpl) newTweetAdded(ctx context.Context, tweet nip.Tweet) {
	if slices.Contains(tweet.Hashtags, "NipNip") {
		s.globalTweets.Lock()
		s.globalTweets.PushBack(tweet)
		s.globalTweets.Unlock()
	} else {
		s.updateMentionedUsersTimeline(tweet.MentionedUserIds, tweet.Id)
		s.updateFollowersTimeline(ctx, tweet.UserId, tweet.Id)
	}
}

func setToSlice[T comparable](set map[T]struct{}) []T {
	result := make([]T, 0, len(set))
	for item := range set {
		result = append(result, item)
	}
	return result
}
