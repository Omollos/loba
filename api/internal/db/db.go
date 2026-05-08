package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the connections tool - shared across the entrire app
var DB *pgxpool.Pool

// Connect opens a connections pool to the Neon PostgreSQL database
// it reads the DATABASE_URL from the enviroment variables
func Connect() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL envrionment variable is not set")
	}

	// pgxpool creates a pool of connections rather than a single connection
	// this implies that multiple requests can hit the database simultaneously
	// without waiting for each other, this is key for a community platform
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// ping the database to confirm the connection actually works
	err = pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to ping the database: %w", err)
	}

	DB = pool
	fmt.Println("Connected to the Neon database successfully")
	return nil
}