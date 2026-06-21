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

	var id string
	err = conn.QueryRow(
		context.Background(),
		`INSERT INTO contributors (username, email, github_handle)
		VALUES ('stevomondi', 'omollosteve022@gmail.com', 'Omollos')
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
		RETURNING id`,
	).Scan(&id)
	if err != nil {
		log.Fatalf("Error inserting contributor: %v", err)
	}

	fmt.Println("Contributor created with ID:", id)
}
