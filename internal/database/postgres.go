package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool" // Standard for high-performance Go apps
)

// Connect establishes a connection pool to the PostgreSQL database
// It returns a *pgxpool.Pool which is safe for concurrent use by multiple goroutines.
func Connect() (*pgxpool.Pool, error) {
	// Construct the connection string using environment variables for security
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Create a connection pool configuration
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %v", err)
	}

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Define the database schema
	// Added CHECK constraint to status to ensure data integrity
	query := `
    -- Create users table if it does not exist
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    -- Create tasks table with status validation for Kanban columns
    CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        title TEXT NOT NULL,
        -- Status must be one of the Kanban board columns
        status TEXT NOT NULL DEFAULT 'todo' 
            CHECK (status IN ('todo', 'in_progress', 'done')),
        started_at TIMESTAMP WITH TIME ZONE,
        finished_at TIMESTAMP WITH TIME ZONE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	// Execute schema creation
	_, err = pool.Exec(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return pool, nil
}
