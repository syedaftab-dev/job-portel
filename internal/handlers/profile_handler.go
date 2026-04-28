// Profile handler contains endpoints for user profile management.
package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Akshatt02/job-portal-backend/internal/services"
)

// updateProfileRequest represents the JSON payload for profile updates.
type updateProfileRequest struct {
	Name          *string  `json:"name,omitempty"`
	Bio           *string  `json:"bio,omitempty"`
	LinkedinURL   *string  `json:"linkedin_url,omitempty"`
	Skills        []string `json:"skills,omitempty"`
	WalletAddress *string  `json:"wallet_address,omitempty"`
}

// GetProfile handles public profile viewing (GET /profile/:id).
// No authentication required - returns limited public user information.
//
// Returns: { id, name, bio, linkedin_url, skills }
func GetProfile(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id required"})
	}

	user, err := services.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

// Me handles authenticated user profile retrieval (GET /me).
// Returns the complete profile of the currently logged-in user.
//
// Requires: Authorization: Bearer <token>
// Returns: { id, name, email, bio, linkedin_url, skills, wallet_address }
func Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	idStr := userID.(string)

	user, err := services.GetUserByID(idStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(user)
}

// UpdateProfile handles profile updates for authenticated users (PUT /profile).
// Supports partial updates - only provided fields are modified.
//
// Requires: Authorization: Bearer <token>
// Request body (all fields optional):
//
//	{
//	  "name": "New Name",
//	  "bio": "Bio text",
//	  "linkedin_url": "https://linkedin.com/in/username",
//	  "wallet_address": "0x123...",
//	  "skills": ["go", "react", "postgres"]
//	}
//
// Response on success (200 OK): { message: "Profile updated successfully" }
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	idStr := userID.(string)

	var req updateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.LinkedinURL != nil {
		updates["linkedin_url"] = *req.LinkedinURL
	}
	if req.WalletAddress != nil {
		updates["wallet_address"] = *req.WalletAddress
	}
	if req.Skills != nil {
		updates["skills"] = req.Skills
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no updates provided"})
	}

	if err := services.UpdateUser(idStr, updates); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update"})
	}

	// Return new profile
	u, err := services.GetUserByID(idStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch updated user"})
	}
	return c.JSON(u)
}
