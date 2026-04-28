// Package utils provides utility functions for password hashing and JWT operations.
package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword creates a bcrypt hash of the provided password.
// Uses bcrypt.DefaultCost (currently 10) for security/performance balance.
//
// Parameters:
// - password: Plain text password to hash
//
// Returns:
// - hash: The bcrypted password hash (safe to store in database)
// - error: Any error during hashing (usually only memory or system errors)
//
// Usage: hash, err := HashPassword("myPassword123")
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a plain text password against a bcrypt hash.
// Returns true only if the password matches the hash.
//
// Parameters:
// - password: Plain text password to verify
// - hash: The bcrypt hash to check against
//
// Returns:
// - true: Password matches the hash
// - false: Password does not match, or invalid hash
//
// Usage: if CheckPassword("myPassword123", storedHash) { /* valid */ }
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
