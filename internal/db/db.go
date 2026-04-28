// Package db provides PostgreSQL database connection management.
//
// This package initializes and maintains a connection pool to PostgreSQL
// for storing user profiles, job listings, and application data.
package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global PostgreSQL connection pool.
// Used by all database operations throughout the application.
// Initialize with Connect() and close with Close().
var Pool *pgxpool.Pool

// Connect establishes a connection pool to PostgreSQL.
//
// Parameters:
// - databaseURL: PostgreSQL connection string
//
// Connection string format:
// postgres://username:password@localhost:5432/database_name
//
// This function will fail fatally if connection cannot be established.
func Connect(databaseURL string) {
	var err error
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	log.Println("Connected to PostgreSQL")
}

// Close gracefully closes the database connection pool.
// Should be called in main() with defer to ensure cleanup.
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
