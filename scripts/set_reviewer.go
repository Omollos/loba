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

	// Set Omollos as reviewer
	_, err = conn.Exec(
		context.Background(),
		`UPDATE contributors SET is_reviewer = true WHERE github_handle = 'Omollos'`,
	)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Println("Omollos set as reviewer.")

	// Also merge the old stevomollo entries to the Omollos account
	// so leaderboard attribution is correct
	_, err = conn.Exec(
		context.Background(),
		`UPDATE entries
		SET contributor_id = '01928a5e-7004-4209-8bf0-17d389052821'
		WHERE contributor_id = 'f1e65842-b9c7-4858-b06d-4b680eb6342b'`,
	)
	if err != nil {
		log.Fatalf("Entry migration failed: %v", err)
	}
	fmt.Println("Entries migrated from stevomollo to Omollos.")

	// Delete the old stevomollo row
	_, err = conn.Exec(
		context.Background(),
		`DELETE FROM contributors WHERE id = 'f1e65842-b9c7-4858-b06d-4b680eb6342b'`,
	)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Old stevomollo row removed.")
	fmt.Println("Done — one clean contributor record for Omollos with reviewer access.")
}
