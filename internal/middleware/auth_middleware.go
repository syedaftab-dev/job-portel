package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/Akshatt02/job-portal-backend/internal/config"
	"github.com/Akshatt02/job-portal-backend/pkg/utils"
)

// AuthRequired is a Fiber middleware that validates JWT tokens
//
// Protected Route Middleware:
// - Checks Authorization header for valid JWT token
// - Validates token signature using JWT_SECRET
// - Extracts user ID from token claims
// - Makes user ID available to handler via c.Locals("user_id")
//
// Token Format:
// Authorization: Bearer <jwt_token>
// Example: Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
//
// Process:
// 1. Check if Authorization header exists
// 2. Parse "Bearer <token>" format
// 3. Validate JWT signature using secret key
// 4. Extract user ID from token claims
// 5. Store user ID in Fiber context locals
// 6. Call next handler
//
// Return Codes:
// - 401 Unauthorized: Missing or invalid token
// - 200 + Next Handler: Valid token, user ID set
//
// Usage:
// In main.go:
//
//	protected := app.Group("", middleware.AuthRequired())
//	protected.Post("/jobs", handlers.CreateJob)
//
// In handler:
//
//	userID := c.Locals("user_id").(string)
//
// Notes:
// - JWT token created at login by handlers.Login
// - Token contains user ID + expiration time
// - Frontend sends token in every protected request
// - Middleware validates before handler executes
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract Authorization header
		// Expected format: "Authorization: Bearer <token>"
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		// Parse "Bearer <token>" format
		// Split on first space, expect 2 parts: ["Bearer", "<token>"]
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header"})
		}

		// Extract token string (second part after Bearer)
		tokenStr := parts[1]

		// Load JWT secret from environment config
		cfg := config.LoadConfig()

		// Validate token signature and extract user ID
		// Returns error if signature invalid or token expired
		userID, err := utils.ParseToken(tokenStr, cfg.JWTSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		// Store user ID in Fiber context locals
		// Available in handler via: c.Locals("user_id")
		c.Locals("user_id", userID)

		// Continue to next middleware/handler
		return c.Next()
	}
}
