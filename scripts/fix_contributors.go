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

	realID := "f1e65842-b9c7-4858-b06d-4b680eb6342b" // stevomollo
	typoID := "3699c9a8-fe0d-49ac-a559-2712faf44cfe" // stevomondi (typo)

	// Step 1 — move all entries from typo account to real account
	result, err := conn.Exec(
		context.Background(),
		`UPDATE entries SET contributor_id = $1 WHERE contributor_id = $2`,
		realID, typoID,
	)
	if err != nil {
		log.Fatalf("Error reassigning entries: %v", err)
	}
	fmt.Printf("Reassigned %d entries to stevomollo\n", result.RowsAffected())

	// Step 2 — move any votes from typo account to real account
	_, err = conn.Exec(
		context.Background(),
		`UPDATE votes SET voter_id = $1 WHERE voter_id = $2`,
		realID, typoID,
	)
	if err != nil {
		log.Fatalf("Error reassigning votes: %v", err)
	}

	// Step 3 — delete the typo contributor row
	_, err = conn.Exec(
		context.Background(),
		`DELETE FROM contributors WHERE id = $1`,
		typoID,
	)
	if err != nil {
		log.Fatalf("Error deleting typo contributor: %v", err)
	}

	fmt.Println("Typo contributor deleted.")
	fmt.Println("All entries and votes now attributed to stevomollo.")
}
