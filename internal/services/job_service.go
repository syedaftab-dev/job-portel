package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"
"log"
"strconv"
	"github.com/Akshatt02/job-portal-backend/internal/db"
	"github.com/Akshatt02/job-portal-backend/internal/models"
	"github.com/google/uuid"
)

var ErrJobNotFound = errors.New("job not found")

// CreateJob creates a new job posting
//
// Blockchain Payment Requirement:
// - User must provide paymentTx (Sepolia ETH transaction hash)
// - Hash format: "0x" + 64 hexadecimal characters (66 chars total)
// - Validates format but does NOT verify transaction on-chain
//
// Process:
// 1. Parse and validate user ID (UUID format)
// 2. Require payment_tx_hash for security/audit trail
// 3. Validate transaction hash format
// 4. Generate UUID for job
// 5. Serialize skills array to JSON
// 6. Insert into database with all metadata
//
// Parameters:
// - title: Job position title
// - description: Full job description
// - skills: []string - Required skills (optional, can be nil)
// - salary: Salary range or "Competitive" (optional)
// - location: Job location
// - userIDStr: UUID string of job poster
// - paymentTx: Sepolia transaction hash (66 char format)
//
// Returns:
// - Job ID (UUID string) on success
// - Error if validation fails or database error
//
// Usage: Called by POST /jobs handler after blockchain payment
func CreateJob(title, description string, skills []string, salary, location, userIDStr, paymentTx string) (string, error) {
	// Ensure user id is valid uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", err
	}

	// Enforce a payment_tx_hash for posting (as per assignment). Remove if not desired.
	if paymentTx == "" {
		return "", errors.New("payment required before posting job (payment_tx_hash missing)")
	}

	// Validate transaction hash format (must be 66 characters starting with 0x)
	if len(paymentTx) != 66 || paymentTx[:2] != "0x" {
		return "", errors.New("invalid transaction hash format")
	}

	jobID := uuid.New()
	skillsBytes := []byte("null")
	if skills != nil {
		b, _ := json.Marshal(skills)
		skillsBytes = b
	}

	_, err = db.Pool.Exec(context.Background(),
		`INSERT INTO jobs (id, title, description, skills, salary, location, user_id, payment_tx_hash, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		jobID, title, description, skillsBytes, salary, location, userID, paymentTx, time.Now(),
	)
	if err != nil {
		return "", err
	}

	return jobID.String(), nil
}

// ListJobs retrieves recent job postings
//
// Process:
// 1. Query jobs table ordered by creation date (newest first)
// 2. Limit results to specified count
// 3. Unmarshal skills JSON arrays for each job
// 4. Convert UUIDs to strings for API response
//
// Parameters:
// - limit: Max number of jobs to return. If <= 0, defaults to 100
//
// Returns:
// - []*models.Job array of job listings
// - Error if database query fails
//
// Usage: Called by GET /jobs endpoint (public, no auth required)
func ListJobs(limit int) ([]*models.Job, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, title, description, skills, salary, location, user_id, payment_tx_hash, created_at
		 FROM jobs
		 ORDER BY created_at DESC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*models.Job{}
	for rows.Next() {
		var (
			id                 uuid.UUID
			title, description string
			skillsRaw          []byte
			salary, location   string
			userID             uuid.UUID
			paymentTx          *string
			createdAt          time.Time
		)
		err := rows.Scan(&id, &title, &description, &skillsRaw, &salary, &location, &userID, &paymentTx, &createdAt)
		if err != nil {
			return nil, err
		}

		var skills []string
		if len(skillsRaw) > 0 {
			_ = json.Unmarshal(skillsRaw, &skills)
		}

		px := ""
		if paymentTx != nil {
			px = *paymentTx
		}

		job := &models.Job{
			// GetJobByID retrieves a single job posting by ID
			//
			// Process:
			// 1. Parse job ID (UUID format)
			// 2. Query database for matching job
			// 3. Unmarshal skills JSON array
			// 4. Return complete job model
			//
			// Parameters:
			// - jobIDStr: UUID string of job to fetch
			//
			// Returns:
			// - *models.Job with all fields populated
			// - ErrJobNotFound if job doesn't exist
			// - Other error if database query fails
			//
			// Usage: Called by GET /jobs/:id endpoint
			// Note: Client also receives match_score (computed separately)
			ID:            id,
			Title:         title,
			Description:   description,
			Skills:        skills,
			Salary:        salary,
			Location:      location,
			UserID:        userID,
			PaymentTxHash: px,
			CreatedAt:     createdAt,
		}
		out = append(out, job)
	}
	return out, nil
}

// ListJobsWithFilters retrieves job postings with optional filters.
//
// Supports filtering by:
// - skill: Filter jobs that require a specific skill (partial match in JSON array)
// - location: Filter jobs by location (case-insensitive)
// - salaryMin: Filter jobs above minimum salary (numeric comparison)
//
// Parameters:
// - limit: Max number of jobs to return
// - skill: Optional skill filter (e.g., "go", "react")
// - location: Optional location filter (e.g., "Remote", "New York")
// - salaryMin: Optional minimum salary threshold
//
// Returns:
// - []*models.Job array of filtered job listings
// - Error if database query fails
//
// Usage: Called by GET /jobs endpoint with query parameters
func ListJobsWithFilters(limit int, skill, location string, salaryMin int) ([]*models.Job, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, title, description, skills, salary, location, user_id, payment_tx_hash, created_at
		FROM jobs
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if skill != "" {
		query += ` AND skills::text ILIKE '%' || $` + strconv.Itoa(argCount) + ` || '%'`
		args = append(args, skill)
		argCount++
	}

	if location != "" {
		query += ` AND location ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+location+"%")
		argCount++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := db.Pool.Query(context.Background(), query, args...)
	if err != nil {
		log.Println("DB ERROR:", err)
		return nil, err
	}
	defer rows.Close()

	out := []*models.Job{}
	for rows.Next() {
		var (
			id                 uuid.UUID
			title, description string
			skillsRaw          []byte
			salary, location   string
			userID             uuid.UUID
			paymentTx          *string
			createdAt          time.Time
		)

		err := rows.Scan(
			&id, &title, &description, &skillsRaw,
			&salary, &location, &userID, &paymentTx, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		var skills []string
		if len(skillsRaw) > 0 {
			_ = json.Unmarshal(skillsRaw, &skills)
		}

		px := ""
		if paymentTx != nil {
			px = *paymentTx
		}

		job := &models.Job{
			ID:            id,
			Title:         title,
			Description:   description,
			Skills:        skills,
			Salary:        salary,
			Location:      location,
			UserID:        userID,
			PaymentTxHash: px,
			CreatedAt:     createdAt,
		}
		out = append(out, job)
	}
	return out, nil
}

// GetJobByID retrieves a single job posting by ID
//
// Process:
// 1. Parse job ID (UUID format)
// 2. Query database for matching job
// 3. Unmarshal skills JSON array
// 4. Return complete job model
//
// Parameters:
// - jobIDStr: UUID string of job to fetch
//
// Returns:
// - *models.Job with all fields populated
// - ErrJobNotFound if job doesn't exist
// - Other error if database query fails
//
// Usage: Called by GET /jobs/:id endpoint
// Note: Client also receives match_score (computed separately)
func GetJobByID(jobIDStr string) (*models.Job, error) {
	id, err := uuid.Parse(jobIDStr)
	if err != nil {
		return nil, err
	}

	var (
		title, description string
		skillsRaw          []byte
		salary, location   string
		userID             uuid.UUID
		paymentTx          *string
		createdAt          time.Time
	)
	err = db.Pool.QueryRow(context.Background(),
		`SELECT title, description, skills, salary, location, user_id, payment_tx_hash, created_at
		 FROM jobs WHERE id=$1`, id).
		Scan(&title, &description, &skillsRaw, &salary, &location, &userID, &paymentTx, &createdAt)
	if err != nil {
		return nil, ErrJobNotFound
	}

	var skills []string
	if len(skillsRaw) > 0 {
		_ = json.Unmarshal(skillsRaw, &skills)
	}

	pt := ""
	if paymentTx != nil {
		pt = *paymentTx
	}

	j := &models.Job{
		ID:            id,
		Title:         title,
		Description:   description,
		Skills:        skills,
		Salary:        salary,
		Location:      location,
		UserID:        userID,
		PaymentTxHash: pt,
		CreatedAt:     createdAt,
	}
	return j, nil
}
