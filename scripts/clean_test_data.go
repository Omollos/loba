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

	// Delete the test contributor created during development
	// Only safe to run before real contributors exist
	_, err = conn.Exec(
		context.Background(),
		`DELETE FROM contributors WHERE username = 'test-contributor'`,
	)
	if err != nil {
		log.Fatalf("Error deleting test contributor: %v", err)
	}

	fmt.Println("Test contributor removed. Database is clean.")
}
