package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Omollos/loba/api/internal/db"
	"github.com/Omollos/loba/api/internal/models"
)

// CreateEntry handles POST /api/v1/entries
// It reads a JSON body, validates the required fields,
// inserts a new entry into the database, and returns
// the created entry as JSON.
func CreateEntry(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests on this handler
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON request body into a NewEntryRequest struct
	var req models.NewEntryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields — source_text and english are mandatory
	// Everything else is optional but encouraged
	if req.SourceText == "" || req.English == "" {
		http.Error(w, "source_text and english are required", http.StatusBadRequest)
		return
	}

	// If no language_id is provided, default to Dholuo
	// We will make this dynamic once more languages are added
	if req.LanguageID == "" {
		var dhuoloID string
		err := db.DB.QueryRow(
			context.Background(),
			"SELECT id FROM languages WHERE code = 'luo'",
		).Scan(&dhuoloID)
		if err != nil {
			http.Error(w, "could not find default language", http.StatusInternalServerError)
			return
		}
		req.LanguageID = dhuoloID
	}

	// Insert the entry into the database
	// Status defaults to 'pending' — no entry goes live without review
	var entry models.Entry
	// Convert empty UUID strings to nil so PostgreSQL receives NULL
	// rather than an empty string, which is invalid for UUID columns
	var contributorID *string
	if req.ContributorID != "" {
		contributorID = &req.ContributorID
	}

	err = db.DB.QueryRow(
		context.Background(),
		`INSERT INTO entries (
			language_id, source_text, english, part_of_speech,
			explanation, english_equivalent, example_source,
			example_english, notes, category, dialect,
			contributor_id, source, source_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		RETURNING id, language_id, source_text, english,
			part_of_speech, explanation, english_equivalent,
			example_source, example_english, notes, category,
			dialect, status, vote_score, created_at`,
		req.LanguageID, req.SourceText, req.English, req.PartOfSpeech,
		req.Explanation, req.EnglishEquivalent, req.ExampleSource,
		req.ExampleEnglish, req.Notes, req.Category, req.Dialect,
		contributorID, req.Source, req.SourceURL,
	).Scan(
		&entry.ID, &entry.LanguageID, &entry.SourceText, &entry.English,
		&entry.PartOfSpeech, &entry.Explanation, &entry.EnglishEquivalent,
		&entry.ExampleSource, &entry.ExampleEnglish, &entry.Notes,
		&entry.Category, &entry.Dialect, &entry.Status,
		&entry.VoteScore, &entry.CreatedAt,
	)
	if err != nil {
		http.Error(w, "could not create entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created

	// Encode the new entry as JSON and write it to the response
	json.NewEncoder(w).Encode(entry)
}

// ListEntries handles GET /api/v1/entries
// Returns all approved entries, optionally filtered by category
func ListEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read optional query parameters for filtering
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "approved" // default to approved entries only
	}

	// Build the query — filter by status always, category optionally
	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Err() error
	}
	var err error

	if category != "" {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT id, language_id, source_text, english, part_of_speech,
				explanation, english_equivalent, example_source, example_english,
				notes, category, dialect, status, vote_score, created_at
			FROM entries
			WHERE status = $1 AND category = $2
			ORDER BY created_at DESC`,
			status, category,
		)
	} else {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT id, language_id, source_text, english, part_of_speech,
				explanation, english_equivalent, example_source, example_english,
				notes, category, dialect, status, vote_score, created_at
			FROM entries
			WHERE status = $1
			ORDER BY created_at DESC`,
			status,
		)
	}

	if err != nil {
		http.Error(w, "could not fetch entries", http.StatusInternalServerError)
		return
	}

	// Collect all entries into a slice
	var entries []models.Entry
	for rows.Next() {
		var e models.Entry
		var reviewedAt *time.Time
		err := rows.Scan(
			&e.ID, &e.LanguageID, &e.SourceText, &e.English,
			&e.PartOfSpeech, &e.Explanation, &e.EnglishEquivalent,
			&e.ExampleSource, &e.ExampleEnglish, &e.Notes,
			&e.Category, &e.Dialect, &e.Status,
			&e.VoteScore, &e.CreatedAt,
		)
		if err != nil {
			http.Error(w, "error reading entries", http.StatusInternalServerError)
			return
		}
		e.ReviewedAt = reviewedAt
		entries = append(entries, e)
	}

	if rows.Err() != nil {
		http.Error(w, "error iterating entries", http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null if no entries found
	if entries == nil {
		entries = []models.Entry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// UpdateEntryStatus handles PUT /api/v1/entries/{id}/approve
// and PUT /api/v1/entries/{id}/flag
// The status to set is determined by which route called it.
func UpdateEntryStatus(newStatus string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract the {id} from the URL path pattern
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "entry id is required", http.StatusBadRequest)
			return
		}

		// Update the entry's status and set reviewed_at to now
		var entry models.Entry
		err := db.DB.QueryRow(
			context.Background(),
			`UPDATE entries
			SET status = $1, reviewed_at = now()
			WHERE id = $2
			RETURNING id, language_id, source_text, english,
				part_of_speech, explanation, english_equivalent,
				example_source, example_english, notes, category,
				dialect, status, vote_score, created_at, reviewed_at`,
			newStatus, id,
		).Scan(
			&entry.ID, &entry.LanguageID, &entry.SourceText, &entry.English,
			&entry.PartOfSpeech, &entry.Explanation, &entry.EnglishEquivalent,
			&entry.ExampleSource, &entry.ExampleEnglish, &entry.Notes,
			&entry.Category, &entry.Dialect, &entry.Status,
			&entry.VoteScore, &entry.CreatedAt, &entry.ReviewedAt,
		)
		if err != nil {
			http.Error(w, "entry not found or update failed: "+err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)
	}
}
