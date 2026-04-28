package models

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a job posting in the Job Portal platform
//
// Fields:
// - ID: Unique identifier (UUID), primary key in database
// - Title: Job position title (e.g., "Senior React Engineer")
// - Description: Full job description including responsibilities
// - Skills: Array of required skills for the job
// - Salary: Compensation range or "Competitive"
// - Location: Job location (remote, office address, etc.)
// - UserID: UUID of user who posted the job
// - PaymentTxHash: Sepolia ETH transaction hash proving payment
// - CreatedAt: Job posting timestamp
//
// Database Table: jobs
// - Skills stored as JSON array in database
// - PaymentTxHash format: "0x" + 64 hex characters (66 chars total)
// - Foreign key: UserID references users(id)
// - Indexed on: created_at (for listing), user_id (for user's jobs)
//
// Blockchain Payment:
// - PaymentTxHash is the Sepolia transaction hash from job creation
// - Value: 0.001 Sepolia ETH sent to ADMIN_WALLET
// - Serves as proof of payment / audit trail
// - NOT verified on-chain (future enhancement)
//
// API Usage:
// - Returned by GET /jobs, GET /jobs/:id, POST /jobs
// - Match score computed by AI (not in this model, added in response)
// - Only users can POST jobs, anyone can GET (list/details)
type Job struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Skills        []string  `json:"skills,omitempty"`
	Salary        string    `json:"salary,omitempty"`
	Location      string    `json:"location,omitempty"`
	UserID        uuid.UUID `json:"user_id"`
	PaymentTxHash string    `json:"payment_tx_hash,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}
