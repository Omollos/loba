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

	contributorID := "f1e65842-b9c7-4858-b06d-4b680eb6342b"

	entryIDs := []string{
		"b6da4b56-3fe8-4d45-a83e-3d2e6ec5f23c", // Apilo
		"56d64473-55cd-47b6-9867-2fd83e046009", // Mikai
		"fe8ffe06-9a33-4113-8617-3aad8e0ad7bc", // Ywak ogwal
	}

	for _, id := range entryIDs {
		_, err := conn.Exec(
			context.Background(),
			`UPDATE entries SET contributor_id = $1 WHERE id = $2`,
			contributorID, id,
		)
		if err != nil {
			log.Printf("Failed to update entry %s: %v", id, err)
			continue
		}
		fmt.Printf("Updated entry %s with contributor %s\n", id, contributorID)
	}

	fmt.Println("Backfill complete.")
}
