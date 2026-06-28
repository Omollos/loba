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
	// Load .env file if it exists (local development)
	// In production (Railway) environment variables are injected directly
	// so missing .env file is not an error
	godotenv.Load()

	// Connect to Neon database
	err := db.Connect()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.DB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Routes
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

	// Auth routes
	http.HandleFunc("GET /auth/github", handlers.GitHubLogin)
	http.HandleFunc("GET /auth/callback", handlers.GitHubCallback)
	http.HandleFunc("GET /auth/me", handlers.GetMe)

	// Public entry routes
	http.HandleFunc("POST /api/v1/entries/{id}/vote", handlers.CastVote)
	http.HandleFunc("GET /api/v1/stats", handlers.GetStats)
	http.HandleFunc("POST /api/v1/contributors", handlers.GetOrCreateContributor)
	http.HandleFunc("GET /api/v1/leaderboard", handlers.GetLeaderboard)
	http.HandleFunc("GET /api/v1/export/jsonl", handlers.ExportJSONL)
	http.HandleFunc("GET /api/v1/export/csv", handlers.ExportCSV)
	http.HandleFunc("GET /api/v1/languages", handlers.GetLanguages)

	// Protected routes — reviewer only
	http.HandleFunc("PUT /api/v1/entries/{id}/approve", handlers.RequireReviewer(handlers.UpdateEntryStatus("approved")))
	http.HandleFunc("PUT /api/v1/entries/{id}/flag", handlers.RequireReviewer(handlers.UpdateEntryStatus("flagged")))
	http.HandleFunc("DELETE /api/v1/entries/{id}", handlers.RequireReviewer(handlers.DeleteEntry))
	http.HandleFunc("PUT /api/v1/entries/{id}/definition", handlers.RequireReviewer(handlers.UpdateEntryDefinition))

	log.Printf("Loba API starting on port %s", port)
	if err := http.ListenAndServe(":"+port, withCORS(http.DefaultServeMux)); err != nil {
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Browsers send an OPTIONS request first to check permissions
		// before the real request — respond OK immediately for these
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
