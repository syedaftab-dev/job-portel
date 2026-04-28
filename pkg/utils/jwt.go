// Package utils provides JWT token generation and validation.
// JWT tokens are used for stateless authentication throughout the API.
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT creates a signed JWT token with a 72-hour expiration.
//
// Parameters:
// - userID: User's UUID as string (stored in token claims)
// - secret: Secret key for HMAC-SHA256 signing
//
// Returns:
// - token: Signed JWT string (can be sent to client)
// - error: Any error during token creation
//
// Token Claims:
// - user_id: The authenticated user's UUID
// - exp: Token expiration time (current time + 72 hours)
// - iat: Token issued-at time
//
// Usage: token, err := GenerateJWT(userID, cfg.JWTSecret)
func GenerateJWT(userID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates and parses a JWT token, returning the user_id claim.
// Verifies the token signature and expiration.
//
// Parameters:
// - tokenStr: The JWT token string to parse
// - secret: Secret key used to sign the token (must match)
//
// Returns:
// - userID: The user_id from token claims (as string UUID)
// - error: Returns nil only if token is valid and not expired
//
// Error conditions:
// - "unexpected signing method": Token uses wrong algorithm
// - "invalid token claims": Claims missing or invalid format
// - "token is invalid": Signature doesn't match or token expired
//
// Usage: userID, err := ParseToken(tokenStr, cfg.JWTSecret)
func ParseToken(tokenStr, secret string) (string, error) {
	parser := &jwt.Parser{}
	token, err := parser.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Validate alg
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["user_id"].(string); ok {
			return uid, nil
		}
	}

	return "", errors.New("invalid token claims")
}
