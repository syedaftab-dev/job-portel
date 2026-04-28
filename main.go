// Package main initializes and runs the Job Portal API server
//
// Job Portal Backend Server
// Purpose: REST API for AI-powered job matching with blockchain payments
// Port: 8080 (configurable via PORT env var)
//
// Features:
// - User authentication (JWT tokens)
// - Job posting with blockchain payment verification
// - AI-powered skill extraction and job matching
// - PostgreSQL database storage
//
// Architecture:
// - Fiber: Lightweight web framework (Go)
// - PostgreSQL: User and job data storage
// - JWT: Secure token-based authentication
// - OpenAI: AI for skill extraction and matching
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/Akshatt02/job-portal-backend/internal/config"
	"github.com/Akshatt02/job-portal-backend/internal/db"
	"github.com/Akshatt02/job-portal-backend/internal/handlers"
	"github.com/Akshatt02/job-portal-backend/internal/middleware"
)

// main initializes the server and configures routes
//
// Initialization sequence:
// 1. Load configuration from environment variables
// 2. Connect to PostgreSQL database
// 3. Create Fiber app with middleware (logging, CORS)
// 4. Define public routes (no authentication required)
// 5. Define protected routes (JWT authentication required)
// 6. Start HTTP server on configured port
func main() {
	// Load environment configuration (DATABASE_URL, PORT, JWT_SECRET, etc.)
	cfg := config.LoadConfig()

	// Establish database connection pool
	db.Connect(cfg.DatabaseURL)
	defer db.Close()

	// Initialize Fiber web application
	app := fiber.New()

	// Middleware: Log all incoming requests
	app.Use(logger.New())

	// Middleware: CORS (Cross-Origin Resource Sharing)
	// Allows frontend to communicate with backend
	// The frontend URL is loaded from environment configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.FrontendURL, // Frontend URL from config
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// PUBLIC ROUTES (no authentication required)

	// User registration endpoint
	// POST /auth/register { name, email, password }
	app.Head("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/auth/register", handlers.Register)

	// User login endpoint
	// POST /auth/login { email, password } -> returns JWT token
	app.Post("/auth/login", handlers.Login)

	// Get public user profile (view someone else's profile)
	// GET /profile/:id -> returns user info without sensitive data
	app.Get("/profile/:id", handlers.GetProfile)

	// List all jobs (browseable by anyone)
	// GET /jobs -> returns array of job listings
	app.Get("/jobs", handlers.ListJobs)

	// List all posts from social feed (browseable by anyone)
	// GET /posts -> returns array of user posts
	app.Get("/posts", handlers.GetPosts)

	// Get user's posts (public user profile posts)
	// GET /posts/:user_id -> returns posts by specific user
	app.Get("/posts/:user_id", handlers.GetUserPosts)

	// PROTECTED ROUTES (JWT authentication required)

	// All routes in this group require valid Authorization header
	// Format: Authorization: Bearer <token>
	//
	protected := app.Group("", middleware.AuthRequired())

	// Get current authenticated user's profile
	// GET /me -> returns logged-in user's full profile
	protected.Get("/me", handlers.Me)

	// Update authenticated user's profile
	// PUT /profile { name, bio, skills, linkedin_url, wallet_address }
	protected.Put("/profile", handlers.UpdateProfile)

	// Get job details with AI-computed match score
	// GET /jobs/:id -> returns job + match_score based on user's skills
	protected.Get("/jobs/:id", handlers.GetJob)

	// Create a new job posting (requires blockchain payment)
	// POST /jobs { title, description, location, payment_tx_hash }
	// payment_tx_hash: Sepolia ETH transaction hash as proof of payment
	protected.Post("/jobs", handlers.CreateJob)

	// Extract skills from resume/bio text using AI
	// POST /ai/extract-skills { bio } -> returns { skills: [...] }
	protected.Post("/ai/extract-skills", handlers.ExtractSkills)

	// Create a new social feed post (career advice, updates)
	// POST /posts { content } -> returns { id, message }
	protected.Post("/posts", handlers.CreatePost)

	// Start HTTP server
	log.Println("Starting server on port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
