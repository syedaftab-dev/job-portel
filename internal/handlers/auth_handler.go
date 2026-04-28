// Package handlers provides HTTP request handlers for all API endpoints.
//
// This package contains the business logic for handling incoming HTTP requests,
// validating input, calling services, and returning responses.
package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Akshatt02/job-portal-backend/internal/config"
	"github.com/Akshatt02/job-portal-backend/internal/services"
	"github.com/Akshatt02/job-portal-backend/pkg/utils"
)

// registerRequest represents the JSON payload for user registration.
type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest represents the JSON payload for user login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration (POST /auth/register).
//
// Request body:
//
//	{
//	  "name": "John Doe",
//	  "email": "john@example.com",
//	  "password": "secure_password"
//	}
//
// Response on success (201 Created):
// { "token": "eyJhbGc..." }
//
// Error responses:
// - 400: Invalid request, missing fields, or email already exists
// - 500: Internal server error
func Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, email and password required"})
	}

	id, err := services.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cfg := config.LoadConfig()
	token, err := utils.GenerateJWT(id, cfg.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}

	return c.JSON(fiber.Map{"token": token})
}

// Login handles user authentication (POST /auth/login).
//
// Request body:
//
//	{
//	  "email": "john@example.com",
//	  "password": "secure_password"
//	}
//
// Response on success (200 OK):
// { "token": "eyJhbGc..." }
//
// Error responses:
// - 400: Missing email or password
// - 401: Invalid credentials
// - 500: Internal server error
func Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
	}

	id, err := services.LoginUser(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	cfg := config.LoadConfig()
	token, err := utils.GenerateJWT(id, cfg.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}

	return c.JSON(fiber.Map{"token": token})
}
