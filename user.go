package nip

import (
	"context"
	"errors"
	"time"
)

// Common domain errors used across the application.
var (
	// ErrNotFound indicates that the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict indicates that a resource with the same identifier already exists.
	ErrConflict = errors.New("resource already exists")
)

// User represents a registered user in the system.
type User struct {
	Id                 uint
	UserName           string
	EmailOrPhoneNumber string
	IsEmail            bool
	CreatedAt          time.Time
}

// UserRepository defines the persistence contract for User entities.
type UserRepository interface {
	// Create persists a new user and returns their unique ID.
	Create(ctx context.Context, user User) (uint, error)
	// List retrieves all registered users.
	List(ctx context.Context) ([]User, error)
	// ListIDsByNames retrieves the IDs for a given list of usernames.
	ListIDsByNames(ctx context.Context, userNames []string) ([]uint, error)
	// GetByID retrieves a user by their unique primary key.
	GetByID(ctx context.Context, userId uint) (User, error)
	// GetByEmailOrPhone retrieves a user by their email or phone number.
	GetByEmailOrPhone(ctx context.Context, emailOrPhone string) (User, error)
	// GetByUserName retrieves a user by their unique username.
	GetByUserName(ctx context.Context, userName string) (User, error)
}
