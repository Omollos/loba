package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Omollos/loba/api/internal/db"
	"github.com/Omollos/loba/api/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to Neon database
	err = db.Connect()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.DB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ── Routes ───────────────────────────────────────────
	// Health check
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Loba API is running.")
	})

	// Entries
	http.HandleFunc("/api/v1/entries", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateEntry(w, r)
		case http.MethodGet:
			handlers.ListEntries(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Approve / flag — Go 1.22 path patterns with {id}
	http.HandleFunc("PUT /api/v1/entries/{id}/approve", handlers.UpdateEntryStatus("approved"))
	http.HandleFunc("PUT /api/v1/entries/{id}/flag", handlers.UpdateEntryStatus("flagged"))

	log.Printf("Loba API starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
