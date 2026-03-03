package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// Connect establishes a connection to the database
func Connect() (*pgx.Conn, error) {
	// Extract the connection string from .env
	connStr := os.Getenv("DB_URL")

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return conn, nil
}
