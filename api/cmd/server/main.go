package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// load environment variables form .env files
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// read the port from environment variables
	// if none is set, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// a simple first route; confirms the server is alive and responding
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
