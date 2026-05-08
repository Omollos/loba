package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/Omollos/loba/api/internal/db"
)

func main() {
	// load environment variables form .env files
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to Neon database
	err = db.Connect
	if err != nil {
		log.Fatal("Database connection failed: %v", err)
	}
	defer db.DB.Close()


	// read the port from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// health check route; confirms the server is alive and responding
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Loba API is running.")
	})

	// start the server
	log.Printf("Starting server on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
