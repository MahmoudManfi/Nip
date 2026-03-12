package nip

import (
	"context"
	"time"
)

// Follow represents a directional relationship between two users.
type Follow struct {
	FollowerId uint
	FollowedId uint
	CreatedAt  time.Time
}

// FollowRepository defines the persistence contract for follow relationships.
type FollowRepository interface {
	// Create establishes a new follow relationship.
	Create(ctx context.Context, f Follow) error
	// ListFollowerIds retrieves the IDs of all users following a specific user.
	ListFollowerIds(ctx context.Context, followedId uint) ([]uint, error)
}
