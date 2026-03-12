package nip

import (
	"context"
	"time"
)

// Tweet represents a short message posted by a user.
type Tweet struct {
	Id               string
	UserId           uint
	Content          string
	MentionedUserIds []uint
	Hashtags         []string
	CreatedAt        time.Time
}

// TweetRepository defines the persistence contract for Tweet entities.
type TweetRepository interface {
	// Create persists a new tweet and returns its unique document ID.
	Create(ctx context.Context, t Tweet) (string, error)
	// ListAllLatest retrieves all tweets in the system, sorted by creation date (descending).
	ListAllLatest(ctx context.Context) ([]Tweet, error)
	// ListLatest retrieves a subset of tweets by their IDs, sorted by creation date (descending).
	ListLatest(ctx context.Context, tweetIds []string) ([]Tweet, error)
	// ListIDsByUserID retrieves the list of tweet IDs posted by a specific user.
	ListIDsByUserID(ctx context.Context, userId uint) ([]string, error)
	// GetByID retrieves a single tweet by its unique document ID.
	GetByID(ctx context.Context, tweetId string) (Tweet, error)
}
