package nip

import "context"

// AuthService handles user registration and authentication flows.
type AuthService interface {
	// SignUp registers a new user securely.
	SignUp(ctx context.Context, userName string, emailOrPhoneNumber string) (uint, error)
	// SendOtp generates and sends a one-time password to the user.
	SendOtp(ctx context.Context, emailOrPhoneNumber string) (int, error)
	// Login validates the OTP and returns the authenticated User.
	Login(ctx context.Context, emailOrPhoneNumber string, otpCode int) (User, error)
}

// TimelineService manages the generation and real-time updates of user feeds.
type TimelineService interface {
	// GetTimeline returns tweets for a user which includes
	// reverse chronologically ordered global tweets,
	// reverse chronologically ordered mentioned tweets,
	// and reverse chronologically ordered followed tweets.
	GetTimeline(ctx context.Context, userId uint) ([]Tweet, error)
	// NewTweetAdded distributes a new tweet to the global list if the tweet contains "NipNip" hashtag.
	// Otherwise it distributes the tweet to the mentioned users' timelines and followers' timelines.
	NewTweetAdded(ctx context.Context, tweetId string) error
	// UserFollowedUser updates the timeline of the follower user with the tweets of the followed user
	// if these tweets are not already in the timeline.
	UserFollowedUser(ctx context.Context, followerId uint, followedId uint) error
}

// TweetService handles the business logic for tweet creation and token extraction.
type TweetService interface {
	// CreateTweet extracts tokens (hashtags/mentions) and persists a new tweet.
	CreateTweet(ctx context.Context, userId uint, content string) (string, error)
}

// UserService handles user-specific social actions.
type UserService interface {
	// Follow establishes a follow relationship between two users.
	Follow(ctx context.Context, userId uint, followedUserId uint) error
}

// OtpService manages the generation, storage, and validation of one-time passwords.
type OtpService interface {
	// GenerateOtp creates a unique OTP for a user with a short TTL.
	GenerateOtp(ctx context.Context, userId uint) (int, error)
	// ValidateOtpCode checks if the provided code is valid and not expired.
	ValidateOtpCode(ctx context.Context, userId uint, otpCode int) error
}

// NotificationService handles outbound communications to users.
type NotificationService interface {
	// SendPushNotification dispatches a message to multiple users via their preferred channel.
	SendPushNotification(ctx context.Context, userIds []uint, content string)
}

// EmailService simulates an external email provider.
type EmailService interface {
	SendOtpEmail(ctx context.Context, email string, otpCode int) error
	SendPushNotification(ctx context.Context, email string, content string) error
	ValidateEmail(ctx context.Context, email string) error
}

// PhoneService simulates an external SMS provider.
type PhoneService interface {
	SendOtpSms(ctx context.Context, phoneNumber string, otpCode int) error
	SendPushNotification(ctx context.Context, phoneNumber string, content string) error
	ValidatePhoneNumber(ctx context.Context, phoneNumber string) error
}
