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

	// Count total rows
	var total int
	conn.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM contributors`,
	).Scan(&total)
	fmt.Println("Total contributor rows:", total)

	// Show all usernames regardless of scan issues
	rows, _ := conn.Query(context.Background(),
		`SELECT id, username FROM contributors`,
	)
	defer rows.Close()

	fmt.Println("All rows:")
	for rows.Next() {
		var id, username string
		rows.Scan(&id, &username)
		fmt.Printf("  %s | %s\n", id, username)
	}

}
