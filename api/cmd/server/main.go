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
	http.HandleFunc("POST /api/v1/entries/{id}/vote", handlers.CastVote)
	http.HandleFunc("GET /api/v1/stats", handlers.GetStats)
	http.HandleFunc("POST /api/v1/contributors", handlers.GetOrCreateContributor)
	http.HandleFunc("GET /api/v1/leaderboard", handlers.GetLeaderboard)

	log.Printf("Loba API starting on port %s", port)
	err = http.ListenAndServe(":"+port, withCORS(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}

// withCORS wraps every request with headers that allow the
// frontend (running on a different port) to call this API.
// In production this will be restricted to loba.dev specifically —
// for local development we allow any origin.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Browsers send an OPTIONS request first to check permissions
		// before the real request — respond OK immediately for these
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
