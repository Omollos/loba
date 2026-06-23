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

	rows, err := conn.Query(
		context.Background(),
		`SELECT id, username, joined_at FROM contributors ORDER BY joined_at`,
	)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("ID | Username | Joined")
	fmt.Println("---")
	for rows.Next() {
		var id, username, joined string
		rows.Scan(&id, &username, &joined)
		fmt.Printf("%s | %s | %s\n", id, username, joined)
	}

}
