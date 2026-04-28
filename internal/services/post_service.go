package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Akshatt02/job-portal-backend/internal/db"
	"github.com/Akshatt02/job-portal-backend/internal/models"
	"github.com/google/uuid"
)

// CreatePost creates a new social feed post for the authenticated user.
//
// Parameters:
// - userID: UUID of the user creating the post
// - content: Post content (career advice, updates, thoughts)
//
// Returns:
// - postID: UUID of the created post
// - error: if content is empty or database insert fails
//
// Database: Inserts into posts table with generated UUID and current timestamp
func CreatePost(userID, content string) (string, error) {
	if content == "" {
		return "", errors.New("post content cannot be empty")
	}

	postID := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO posts (id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Pool.Exec(context.Background(), query, postID, userID, content, now)
	if err != nil {
		return "", err
	}

	return postID, nil
}

// GetPosts retrieves all posts from the social feed, ordered by newest first.
//
// Parameters:
// - limit: Maximum number of posts to return
//
// Returns:
// - posts: Slice of Post objects with user details
// - error: if database query fails
//
// Includes user name for each post (via JOIN with users table)
func GetPosts(limit int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, u.name, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1
	`

	rows, err := db.Pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.Content, &p.CreatedAt); err != nil {
			log.Println(err)
			return nil, err
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetUserPosts retrieves all posts created by a specific user.
//
// Parameters:
// - userID: UUID of the user whose posts to fetch
//
// Returns:
// - posts: Slice of Post objects
// - error: if database query fails
func GetUserPosts(userID string) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, u.name, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := db.Pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
