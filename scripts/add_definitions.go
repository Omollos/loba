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
		`ALTER TABLE entries
		ADD COLUMN IF NOT EXISTS definition_source TEXT,
		ADD COLUMN IF NOT EXISTS definition_english TEXT`,
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("definition_source and definition_english columns added.")
}
