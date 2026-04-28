package models

import ("github.com/google/uuid"
	"time"
)

// Post represents a user's blog post or career update in the social feed.
//
// Fields:
// - ID: Unique post identifier (UUID)
// - UserID: Reference to the user who created the post
// - Content: The post text content (career advice, updates, etc.)
// - CreatedAt: Timestamp when the post was created
type Post struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	// User details included for feed display
	UserName string `json:"user_name,omitempty"`
	UserBio  string `json:"user_bio,omitempty"`
}
