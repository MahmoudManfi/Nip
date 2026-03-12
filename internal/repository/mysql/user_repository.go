package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"nip"
	"strings"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepo creates a new MySQL implementation of nip.UserRepository.
func NewUserRepo(db *sql.DB) nip.UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the MySQL database.
func (r *userRepository) Create(ctx context.Context, user nip.User) (uint, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (user_name, email_or_phone, is_email) VALUES (?, ?, ?)`,
		user.UserName, user.EmailOrPhoneNumber, user.IsEmail,
	)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return uint(id), nil
}

// List retrieves all users from the database, ordered by ID descending.
func (r *userRepository) List(ctx context.Context) ([]nip.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_name, email_or_phone, created_at FROM users ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []nip.User
	for rows.Next() {
		var u nip.User
		if err := rows.Scan(&u.Id, &u.UserName, &u.EmailOrPhoneNumber, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// ListIDsByNames retrieves user IDs for a set of usernames using an IN clause.
func (r *userRepository) ListIDsByNames(ctx context.Context, userNames []string) ([]uint, error) {
	if len(userNames) == 0 {
		return []uint{}, nil
	}

	placeholders := make([]string, len(userNames))
	args := make([]any, len(userNames))
	for i, userName := range userNames {
		placeholders[i] = "?"
		args[i] = userName
	}

	query := fmt.Sprintf(`SELECT id FROM users WHERE user_name IN (%s) ORDER BY id DESC`,
		strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list by names: %w", err)
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

// GetByID finds a user by their primary key. Returns nip.ErrNotFound if not found.
func (r *userRepository) GetByID(ctx context.Context, userId uint) (nip.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_name, email_or_phone, is_email, created_at FROM users WHERE id = ?`,
		userId,
	)
	var u nip.User
	err := row.Scan(&u.Id, &u.UserName, &u.EmailOrPhoneNumber, &u.IsEmail, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nip.User{}, nip.ErrNotFound
	}
	if err != nil {
		return nip.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetByEmailOrPhone finds a user by their contact info. Returns nip.ErrNotFound if not found.
func (r *userRepository) GetByEmailOrPhone(ctx context.Context, emailOrPhone string) (nip.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_name, email_or_phone, is_email, created_at FROM users WHERE email_or_phone = ?`,
		emailOrPhone,
	)
	var u nip.User
	err := row.Scan(&u.Id, &u.UserName, &u.EmailOrPhoneNumber, &u.IsEmail, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nip.User{}, nip.ErrNotFound
	}
	if err != nil {
		return nip.User{}, fmt.Errorf("get user by email/phone: %w", err)
	}
	return u, nil
}

// GetByUserName finds a user by their login name. Returns nip.ErrNotFound if not found.
func (r *userRepository) GetByUserName(ctx context.Context, userName string) (nip.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_name, email_or_phone, is_email, created_at FROM users WHERE user_name = ?`,
		userName,
	)
	var u nip.User
	err := row.Scan(&u.Id, &u.UserName, &u.EmailOrPhoneNumber, &u.IsEmail, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nip.User{}, nip.ErrNotFound
	}
	if err != nil {
		return nip.User{}, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}
