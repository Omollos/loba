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

	_, err = conn.Exec(
		context.Background(),
		`ALTER TABLE contributors
		ADD COLUMN IF NOT EXISTS github_id BIGINT UNIQUE,
		ADD COLUMN IF NOT EXISTS avatar_url TEXT,
		ADD COLUMN IF NOT EXISTS is_reviewer BOOLEAN DEFAULT false;

		-- Make me a reviewer immediately
		UPDATE contributors
		SET is_reviewer = true
		WHERE username = 'stevomollo';`,
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("github_id, avatar_url, is_reviewer columns added.")
	fmt.Println("stevomollo set as reviewer.")
}
