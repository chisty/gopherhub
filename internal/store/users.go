package store

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id,created_at`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, user.Username, user.Password.hash, user.Email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
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

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invExpiry time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		// create user invitation
		if err := s.createUserInvitation(ctx, tx, user.ID, token, invExpiry); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, userID int64, token string, invExpiry time.Duration) error {
	query := `INSERT INTO invitation(token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(invExpiry))
	if err != nil {
		return err
	}

	return nil
}
