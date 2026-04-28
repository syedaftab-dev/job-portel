package handlers

import (
	"github.com/Akshatt02/job-portal-backend/internal/models"
	"github.com/Akshatt02/job-portal-backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// createPostRequest represents the JSON payload for creating a new post.
type createPostRequest struct {
	Content string `json:"content"`
}

// CreatePost handles creating a new social feed post (POST /posts).
// Requires: Authorization: Bearer <token>
//
// Request body:
// { "content": "Just launched my new project using Go and React!" }
//
// Response on success (201 Created):
// { "id": "post-uuid", "message": "Post created successfully" }
//
// Error responses:
// - 400: Post content is empty
// - 401: User not authenticated
// - 500: Database error
func CreatePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	uidStr := userID.(string)

	var req createPostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "post content cannot be empty"})
	}

	postID, err := services.CreatePost(uidStr, req.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create post"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      postID,
		"message": "Post created successfully",
	})
}

// GetPosts handles fetching all posts from the social feed (GET /posts).
// No authentication required - returns all posts ordered by newest first.
//
// Optional Query Parameters:
// - ?limit=10 (default: 50, max: 100)
//
// Returns: Array of posts with user details
// [
//
//	{
//	  "id": "post-uuid",
//	  "user_id": "user-uuid",
//	  "user_name": "John Doe",
//	  "user_bio": "Software Engineer",
//	  "content": "Just launched my new project...",
//	  "created_at": "2025-02-10T10:30:00Z"
//	}
//
// ]
func GetPosts(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}

	posts, err := services.GetPosts(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch posts"})
	}

	if posts == nil {
		posts = []models.Post{}
	}

	return c.JSON(posts)
}

// GetUserPosts handles fetching all posts by a specific user (GET /posts/:user_id).
// No authentication required - returns public user posts.
//
// Returns: Array of posts by the specified user
func GetUserPosts(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id required"})
	}

	posts, err := services.GetUserPosts(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch user posts"})
	}

	if posts == nil {
		posts = []models.Post{}
	}

	return c.JSON(posts)
}
