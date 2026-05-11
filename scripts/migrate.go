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
	// load environment variables
	err := godotenv.Load("../api/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to Neon
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unabl to connect to database: %v", err)
	}
	defer conn.Close(context.Background())
	fmt.Println("Connected to Neon database")

	// Read the schema file
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Unable to read schema file: %v", err)
	}

	// Execute the schema file
	_, err = conn.Exec(context.Background(), string(schema))
	if err != nil {
		log.Fatalf("Error running schema: %v", err)
	}

	fmt.Println("Schema successfully created")
	fmt.Println("Dholuo language row inserted")
	fmt.Println("Loba database is ready")
}
