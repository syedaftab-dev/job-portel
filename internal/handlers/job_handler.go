package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Akshatt02/job-portal-backend/internal/models"
	"github.com/Akshatt02/job-portal-backend/internal/services"
)

// createJobRequest represents the JSON payload for creating a new job posting.
type createJobRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Skills        []string `json:"skills,omitempty"`
	Salary        string   `json:"salary,omitempty"`
	Location      string   `json:"location,omitempty"`
	PaymentTxHash string   `json:"payment_tx_hash,omitempty"`
}

// jobWithScoreResponse represents a job with its AI-computed match score.
type jobWithScoreResponse struct {
	*models.Job `json:"job"`
	MatchScore  int `json:"match_score"`
}

// CreateJob handles job posting creation (POST /jobs).
// Requires: Authorization: Bearer <token>
// Request body:
//
//	{
//	  "title": "Senior Go Developer",
//	  "description": "Looking for experienced Go developer...",
//	  "location": "Remote",
//	  "salary": "$120k-150k",
//	  "skills": ["go", "postgresql", "docker"],
//	  "payment_tx_hash": "0x123abc...(66 chars)"
//	}
//
// Payment Requirements:
// - payment_tx_hash: Ethereum Sepolia transaction hash (format: 0x + 64 hex chars)
// - Must be a valid transaction from user to platform wallet (0.001 SETH)
// - Validates format before accepting job posting
//
// Response on success (201 Created):
// { "id": "job-uuid", "message": "Job posted successfully with blockchain payment confirmation" }
func CreateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	uidStr := userID.(string)

	var req createJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Title == "" || req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and description required"})
	}

	if req.PaymentTxHash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment_tx_hash is required - blockchain payment must be completed first"})
	}

	jobID, err := services.CreateJob(req.Title, req.Description, req.Skills, req.Salary, req.Location, uidStr, req.PaymentTxHash)
	if err != nil {
		// Return appropriate error messages for different failure scenarios
		errorMsg := err.Error()
		if errorMsg == "invalid transaction hash format" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid transaction hash format - must be a valid Ethereum transaction hash"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errorMsg})
	}

	// Respond with created job id
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      jobID,
		"message": "Job posted successfully with blockchain payment confirmation",
	})
}

// ListJobs handles public job listing (GET /jobs).
// No authentication required - returns all available job postings.
// Supports filtering by skill, location.
//
// Query Parameters:
// - ?skill=go - Filter by required skill
// - ?location=remote - Filter by location (case-insensitive, partial match)
// - ?limit=20 - Number of jobs to return (default: 50, max: 100)
//
// Examples:
// - GET /jobs - All jobs
// - GET /jobs?skill=react&location=remote - React jobs in Remote locations
// - GET /jobs?location=New%20York&limit=10 - First 10 jobs in New York
//
// Returns: Array of jobs ordered by newest first (created_at DESC)
func ListJobs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}

	skill := c.Query("skill")
	location := c.Query("location")

	// If filters provided, use filtered query
	if skill != "" || location != "" {
		jobs, err := services.ListJobsWithFilters(limit, skill, location, 0)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list jobs"})
		}
		if jobs == nil {
			jobs = []*models.Job{}
		}
		return c.JSON(jobs)
	}

	// Otherwise use basic list
	jobs, err := services.ListJobs(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list jobs"})
	}
	if jobs == nil {
		jobs = []*models.Job{}
	}
	return c.JSON(jobs)
}

// GetJob handles job detail retrieval with skill match scoring (GET /jobs/:id).
// Returns the job posting along with a personalized match score based on the
// authenticated user's skills compared to job requirements.
//
// Requires: Authorization: Bearer <token>
// Returns: { job: {...}, match_score: 85 }
// - match_score: 0-100% indicating how well user's skills match the job
func GetJob(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id required"})
	}

	// User is guaranteed to be logged in (protected route)
	userID := c.Locals("user_id")
	uidStr := userID.(string)

	// Get the job
	job, err := services.GetJobByID(id)
	if err != nil {
		if err == services.ErrJobNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch job"})
	}

	// Get user details to compute match score
	user, err := services.GetUserByID(uidStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch user"})
	}

	// Compute match score
	score, err := services.ComputeMatchScore(c.Context(), user.Skills, job.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to compute match score"})
	}

	// Return job with match score
	return c.JSON(jobWithScoreResponse{
		Job:        job,
		MatchScore: score,
	})
}
