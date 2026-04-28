// AI handler contains endpoints for AI-powered skill extraction.
// Uses Google Gemini to analyze text and extract professional skills.
package handlers

import (
	"context"

	"github.com/Akshatt02/job-portal-backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// extractSkillsRequest represents the JSON payload for skill extraction.
type extractSkillsRequest struct {
	Bio string `json:"bio"`
}

// ExtractSkills extracts professional skills from text using AI (POST /ai/extract-skills).
// Calls Google Gemini API to analyze the provided bio/resume text.
//
// Requires: Authorization: Bearer <token>
// Request body:
// { "bio": "I have 5 years of experience with Go, React, and PostgreSQL..." }
//
// Response on success (200 OK):
// { "skills": ["go", "react", "postgresql", ...] }
//
// The extracted skills are automatically saved to the user's profile.
func ExtractSkills(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	uidStr := userID.(string)

	var req extractSkillsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Bio == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bio is required"})
	}

	skills, err := services.ExtractSkillsFromText(context.Background(), req.Bio)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ai extraction failed", "details": err.Error()})
	}

	// update user skills in DB
	updates := map[string]interface{}{"skills": skills}
	if err := services.UpdateUser(uidStr, updates); err != nil {
		// still return the skills even if DB update fails
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user skills", "details": err.Error()})
	}

	return c.JSON(fiber.Map{"skills": skills})
}
