package store

import (
	"context"
	"database/sql"
	"log"

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
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
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

func (s *PostStore) GetUserFeed(ctx context.Context, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {

	log.Println("GetUserFeed:", fq.UserID)

	query := `SELECT p.id, p.title, p.content, p.user_id, p.tags, p.version, p.created_at,
			u.username,
			COUNT(c.id) AS comments_count 
			FROM posts p 
			LEFT JOIN comments c ON p.id = c.post_id
			LEFT JOIN users u ON p.user_id = u.id
			JOIN followers f ON p.user_id = f.follower_id OR p.user_id = $1
			WHERE 
				f.user_id = $1 OR p.user_id = $1 AND
				(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
			GROUP BY p.id , u.username
			ORDER BY p.created_at ` + fq.Sort + `  
			LIMIT $2 OFFSET $3;`

	ctx, cancel := context.WithTimeout(ctx, DBTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, fq.UserID, fq.Limit, fq.Offset, fq.Search)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var detailPost PostWithMetadata
		err := rows.Scan(&detailPost.ID, &detailPost.Title, &detailPost.Content, &detailPost.UserID,
			pq.Array(&detailPost.Tags), &detailPost.Version, &detailPost.CreatedAt, &detailPost.User.Username,
			&detailPost.CommentsCount)
		if err != nil {
			return nil, err
		}

		feed = append(feed, detailPost)
	}

	log.Println(feed)

	return feed, nil
}
