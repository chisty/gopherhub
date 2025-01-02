package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (title, content, user_id, tags) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID,
		pq.Array(&post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `SELECT id, title, content, user_id, tags, version, created_at, updated_at FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.UserID, pq.Array(&post.Tags), &post.Version, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil

}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `UPDATE posts SET title = $1, content = $2, version = version + 1 
			WHERE id = $3 AND version = $4 RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
