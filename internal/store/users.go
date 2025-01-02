package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id,created_at`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE username = $1`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, username, email, password, created_at FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
