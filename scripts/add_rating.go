package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../api/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer conn.Close(context.Background())

	// Add rating column to entries table
	// DEFAULT 'general' means all existing entries are
	// automatically classified as safe — no manual backfill needed
	_, err = conn.Exec(
		context.Background(),
		`ALTER TABLE entries
		ADD COLUMN IF NOT EXISTS rating VARCHAR(20)
		DEFAULT 'general'
		CHECK (rating IN ('general', 'mature', 'restricted'))`,
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Rating column added successfully.")
	fmt.Println("All existing entries defaulted to 'general'.")
}