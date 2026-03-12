package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"nip"
)

type followRepository struct {
	db *sql.DB
}

// NewFollowRepository creates a new MySQL implementation of nip.FollowRepository.
func NewFollowRepository(db *sql.DB) nip.FollowRepository {
	return &followRepository{db: db}
}

// Create inserts a new follow record into the MySQL database.
func (r *followRepository) Create(ctx context.Context, f nip.Follow) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO follows (follower_id, followed_id) VALUES (?, ?)`,
		f.FollowerId, f.FollowedId,
	)
	if err != nil {
		return fmt.Errorf("insert follow: %w", err)
	}
	return nil
}

// ListFollowerIds retrieves all user IDs that are following the given followedId.
func (r *followRepository) ListFollowerIds(ctx context.Context, followedId uint) ([]uint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT follower_id FROM follows WHERE followed_id = ?`,
		followedId,
	)
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}
	defer rows.Close()

	var out []uint
	for rows.Next() {
		var u uint
		if err := rows.Scan(&u); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
