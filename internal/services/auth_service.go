package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Akshatt02/job-portal-backend/internal/db"
	"github.com/Akshatt02/job-portal-backend/internal/models"
	"github.com/Akshatt02/job-portal-backend/pkg/utils"
	"github.com/google/uuid"
)

// RegisterUser creates a new user account with email and password
//
// Process:
// 1. Check if email already registered (prevent duplicates)
// 2. Hash password using bcrypt (cost 10)
// 3. Generate UUID for new user
// 4. Insert user record into PostgreSQL
//
// Parameters:
// - name: User's full name
// - email: Unique email address
// - password: Plain text password (will be hashed)
//
// Returns:
// - User ID (UUID string) on success
// - Error if email exists or database fails
//
// Usage: Called by /auth/register endpoint
func RegisterUser(name, email, password string) (string, error) {
	// check if email exists
	var exists bool
	err := db.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("email already registered")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	id := uuid.New()

	_, err = db.Pool.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password_hash, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		id, name, email, hash, time.Now(),
	)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

// LoginUser authenticates user with email and password
//
// Process:
// 1. Query database for user with given email
// 2. Compare provided password hash with stored hash (bcrypt)
// 3. Return user ID on successful match
//
// Parameters:
// - email: User's email address
// - password: Plain text password to verify
//
// Returns:
// - User ID (UUID string) on success
// - Error if user not found or password doesn't match
//
// Usage: Called by /auth/login endpoint, returns ID for JWT token creation
func LoginUser(email, password string) (string, error) {
	var id uuid.UUID
	var hash string

	err := db.Pool.QueryRow(context.Background(),
		"SELECT id, password_hash FROM users WHERE email=$1",
		email,
	).Scan(&id, &hash)

	if err != nil {
		return "", err
	}

	if !utils.CheckPassword(password, hash) {
		return "", errors.New("invalid credentials")
	}

	return id.String(), nil
}

// GetUserByID retrieves complete user profile by ID
//
// Process:
// 1. Query users table by ID
// 2. Unmarshal skills JSON array
// 3. Handle nullable fields (bio, linkedin_url, wallet_address)
// 4. Return fully populated User model
//
// Parameters:
// - userID: UUID string of user to fetch
//
// Returns:
// - *models.User with all profile information
// - Error if user not found
//
// Usage: Called by /me, /profile/:id, and for loading user context
func GetUserByID(userID string) (*models.User, error) {
	var (
		id        uuid.UUID
		name      string
		email     string
		bio       *string
		linkedin  *string
		skillsRaw []byte
		wallet    *string
		createdAt time.Time
	)

	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, name, email, bio, linkedin_url, skills, wallet_address, created_at
		 FROM users WHERE id=$1`, userID,
	).Scan(&id, &name, &email, &bio, &linkedin, &skillsRaw, &wallet, &createdAt)

	if err != nil {
		return nil, err
	}

	var skills []string
	if len(skillsRaw) > 0 {
		// UpdateUser modifies user profile fields
		//
		// Supported fields in updates map:
		// - "name": string - User's full name
		// - "bio": string - User biography/description
		// - "linkedin_url": string - LinkedIn profile URL
		// - "wallet_address": string - Ethereum wallet address (Sepolia)
		// - "skills": []string - Array of skill tags
		//
		// Process:
		// 1. Validate that fields are correct Go types
		// 2. Build SQL UPDATE statement dynamically (safe with parameterized queries)
		// 3. Execute update on database
		// 4. Only provided fields are updated (partial updates supported)
		//
		// Parameters:
		// - userID: UUID string of user to update
		// - updates: Map with field names as keys and new values
		//
		// Returns:
		// - nil on success
		// - Error if database operation fails
		//
		// Usage: Called by PUT /profile endpoint
		//
		// Example:
		//   updates := map[string]interface{}{
		//     "bio": "Software engineer with 5 years experience",
		//     "skills": []string{"React", "Go", "PostgreSQL"},
		//     "wallet_address": "0x1234567890abcdef..."
		//   }
		//   UpdateUser(userID, updates)
		_ = json.Unmarshal(skillsRaw, &skills)
	}

	u := &models.User{
		ID:            id,
		Name:          name,
		Email:         email,
		Bio:           safeStr(bio),
		LinkedinURL:   safeStr(linkedin),
		Skills:        skills,
		WalletAddress: safeStr(wallet),
		CreatedAt:     createdAt,
	}
	return u, nil
}

func UpdateUser(userID string, updates map[string]interface{}) error {
	// Build update dynamically but safely.
	// Allowed fields: name, bio, linkedin_url, skills ([]string), wallet_address
	args := []interface{}{}
	setClauses := []string{}
	argIdx := 1

	if v, ok := updates["name"].(string); ok {
		setClauses = append(setClauses, `name = $`+itoa(argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["bio"].(string); ok {
		setClauses = append(setClauses, `bio = $`+itoa(argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["linkedin_url"].(string); ok {
		setClauses = append(setClauses, `linkedin_url = $`+itoa(argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["wallet_address"].(string); ok {
		setClauses = append(setClauses, `wallet_address = $`+itoa(argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["skills"].([]string); ok {
		// marshal to JSON and set
		skillsBytes, _ := json.Marshal(v)
		setClauses = append(setClauses, `skills = $`+itoa(argIdx))
		args = append(args, skillsBytes)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // nothing to update
	}

	// Append user id
	args = append(args, userID)
	query := `UPDATE users SET ` + join(setClauses, ", ") + ` WHERE id = $` + itoa(argIdx)

	_, err := db.Pool.Exec(context.Background(), query, args...)
	return err
}

// small helpers
func safeStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
func itoa(i int) string { return fmt.Sprintf("%d", i) }
func join(arr []string, sep string) string {
	out := ""
	for i, s := range arr {
		if i != 0 {
			out += sep
		}
		out += s
	}
	return out
}
