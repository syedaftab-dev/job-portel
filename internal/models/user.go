package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the Job Portal platform
//
// Fields:
// - ID: Unique identifier (UUID), primary key in database
// - Name: User's full name
// - Email: Unique email address, used for login
// - Bio: Optional biography/description
// - LinkedinURL: Optional LinkedIn profile URL
// - Skills: Array of skill tags extracted from resume/bio
// - WalletAddress: Optional Ethereum wallet address (for job posting)
// - CreatedAt: Account creation timestamp
//
// Database Table: users
// - Password hash stored separately for security (not in this model)
// - Skills stored as JSON array in database
// - All optional fields can be empty strings
//
// API Usage:
// - Returned by GET /me, GET /profile/:id, POST /auth/login, PUT /profile
// - Skills array used for job matching algorithm
// - WalletAddress indicates if user has blockchain payments enabled
type User struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Bio           string    `json:"bio,omitempty"`
	LinkedinURL   string    `json:"linkedin_url,omitempty"`
	Skills        []string  `json:"skills,omitempty"`
	WalletAddress string    `json:"wallet_address,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}
