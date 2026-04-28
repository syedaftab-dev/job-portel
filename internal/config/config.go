// Package config handles environment configuration loading for the Job Portal backend.
//
// This package loads all required environment variables from the system or .env file.
// For deployment, ensure all required variables are set in the deployment environment.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all server configuration loaded from environment variables.
//
// Fields:
// - Port: HTTP server port (default: 8080)
// - DatabaseURL: PostgreSQL connection string (required)
// - JWTSecret: Secret key for JWT token signing/validation (required)
// - FrontendURL: Frontend application URL for CORS (default: http://localhost:5173)
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	FrontendURL string
}

func LoadConfig() *Config {
	// Attempt to load .env silently if exists
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required in env")
	}

	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		log.Fatal("JWT_SECRET is required in env")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   jwt,
		FrontendURL: frontendURL,
	}
}
