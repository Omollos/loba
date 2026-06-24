package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"fmt"

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
			dialect, contributor_id, source, source_url,
			status, vote_score, created_at`,
		req.LanguageID, req.SourceText, req.English, req.PartOfSpeech,
		req.Explanation, req.EnglishEquivalent, req.ExampleSource,
		req.ExampleEnglish, req.Notes, req.Category, req.Dialect,
		contributorID, req.Source, req.SourceURL,
	).Scan(
		&entry.ID, &entry.LanguageID, &entry.SourceText, &entry.English,
		&entry.PartOfSpeech, &entry.Explanation, &entry.EnglishEquivalent,
		&entry.ExampleSource, &entry.ExampleEnglish, &entry.Notes,
		&entry.Category, &entry.Dialect, &entry.ContributorID,
		&entry.Source, &entry.SourceURL,
		&entry.Status, &entry.VoteScore, &entry.CreatedAt,
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
	lang := r.URL.Query().Get("lang")
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
			`SELECT e.id, e.language_id, e.source_text, e.english, e.part_of_speech,
				e.explanation, e.english_equivalent, e.example_source, e.example_english,
				e.notes, e.category, e.dialect, e.contributor_id, e.status, e.vote_score, e.created_at
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = $1
			AND e.category = $2
			AND ($3 = '' OR l.code = $3)
			ORDER BY e.created_at DESC`,
			status, category, lang,
		)
	} else {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT e.id, e.language_id, e.source_text, e.english, e.part_of_speech,
				e.explanation, e.english_equivalent, e.example_source, e.example_english,
				e.notes, e.category, e.dialect, e.contributor_id, e.status, e.vote_score, e.created_at
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = $1
			AND ($2 = '' OR l.code = $2)
			ORDER BY e.created_at DESC`,
			status, lang,
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
			&e.Category, &e.Dialect, &e.ContributorID, &e.Status,
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

// CastVote handles POST /api/v1/entries/{id}/vote
// Records a vote from a contributor and updates the entry's
// running vote_score. One vote per contributor per entry —
// enforced by the database's PRIMARY KEY on (entry_id, voter_id).
func CastVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entryID := r.PathValue("id")
	if entryID == "" {
		http.Error(w, "entry id is required", http.StatusBadRequest)
		return
	}

	// Expected body: { "voter_id": "uuid", "vote": 1 }
	var req struct {
		VoterID string `json:"voter_id"`
		Vote    int    `json:"vote"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.VoterID == "" {
		http.Error(w, "voter_id is required", http.StatusBadRequest)
		return
	}
	if req.Vote != 1 && req.Vote != -1 {
		http.Error(w, "vote must be 1 or -1", http.StatusBadRequest)
		return
	}

	// Use a transaction — both the vote insert and the score
	// update must succeed together, or neither happens.
	// This prevents the vote_score from ever drifting out of
	// sync with the actual votes recorded.
	tx, err := db.DB.Begin(context.Background())
	if err != nil {
		http.Error(w, "could not start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background()) // safe no-op if Commit succeeds

	// Insert the vote. If this contributor already voted on this
	// entry, ON CONFLICT updates their existing vote instead of
	// erroring or creating a duplicate.
	_, err = tx.Exec(
		context.Background(),
		`INSERT INTO votes (entry_id, voter_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT (entry_id, voter_id)
		DO UPDATE SET vote = $3, voted_at = now()`,
		entryID, req.VoterID, req.Vote,
	)
	if err != nil {
		http.Error(w, "could not record vote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Recalculate vote_score as the sum of all votes for this entry
	// Recalculating rather than incrementing means the score is
	// always accurate even if a contributor changes their vote
	var newScore int
	err = tx.QueryRow(
		context.Background(),
		`SELECT COALESCE(SUM(vote), 0) FROM votes WHERE entry_id = $1`,
		entryID,
	).Scan(&newScore)
	if err != nil {
		http.Error(w, "could not calculate vote score", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(
		context.Background(),
		`UPDATE entries SET vote_score = $1 WHERE id = $2`,
		newScore, entryID,
	)
	if err != nil {
		http.Error(w, "could not update vote score", http.StatusInternalServerError)
		return
	}

	// Commit the transaction — both changes are now permanent together
	err = tx.Commit(context.Background())
	if err != nil {
		http.Error(w, "could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry_id":   entryID,
		"vote_score": newScore,
	})
}

// GetStats handles GET /api/v1/stats
// Returns corpus-wide totals used to power dashboards
// like the progress bars and stats strip on the landing page.
func GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type Stats struct {
		TotalApproved     int            `json:"total_approved"`
		TotalPending      int            `json:"total_pending"`
		TotalFlagged      int            `json:"total_flagged"`
		TotalContributors int            `json:"total_contributors"`
		ByCategory        map[string]int `json:"by_category"`
	}

	var stats Stats
	stats.ByCategory = make(map[string]int)

	// Count entries by status
	err := db.DB.QueryRow(
		context.Background(),
		`SELECT
			COUNT(*) FILTER (WHERE status = 'approved'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'flagged')
		FROM entries`,
	).Scan(&stats.TotalApproved, &stats.TotalPending, &stats.TotalFlagged)
	if err != nil {
		http.Error(w, "could not fetch entry stats", http.StatusInternalServerError)
		return
	}

	// Count total contributors
	err = db.DB.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM contributors`,
	).Scan(&stats.TotalContributors)
	if err != nil {
		http.Error(w, "could not fetch contributor count", http.StatusInternalServerError)
		return
	}

	// Count approved entries grouped by category
	rows, err := db.DB.Query(
		context.Background(),
		`SELECT category, COUNT(*)
		FROM entries
		WHERE status = 'approved' AND category IS NOT NULL
		GROUP BY category
		ORDER BY COUNT(*) DESC`,
	)
	if err != nil {
		http.Error(w, "could not fetch category breakdown", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			http.Error(w, "error reading category stats", http.StatusInternalServerError)
			return
		}
		stats.ByCategory[category] = count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetOrCreateContributor handles POST /api/v1/contributors
// Looks up a contributor by username. If they don't exist yet,
// creates them. This is a simple stand-in for proper auth —
// it will be replaced by GitHub OAuth in Phase 2.
func GetOrCreateContributor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	var id string
	err = db.DB.QueryRow(
		context.Background(),
		`INSERT INTO contributors (username)
		VALUES ($1)
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
		RETURNING id`,
		req.Username,
	).Scan(&id)
	if err != nil {
		http.Error(w, "could not get or create contributor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":       id,
		"username": req.Username,
	})
}

// GetLeaderboard handles GET /api/v1/leaderboard
// Returns contributors ranked by their number of approved entries.
func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type LeaderboardEntry struct {
		Username   string `json:"username"`
		EntryCount int    `json:"entry_count"`
	}

	rows, err := db.DB.Query(
		context.Background(),
		`SELECT c.username, COUNT(e.id) as entry_count
		FROM contributors c
		JOIN entries e ON e.contributor_id = c.id
		WHERE e.status = 'approved'
		GROUP BY c.username
		ORDER BY entry_count DESC
		LIMIT 20`,
	)
	if err != nil {
		http.Error(w, "could not fetch leaderboard", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leaderboard []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.EntryCount); err != nil {
			http.Error(w, "error reading leaderboard", http.StatusInternalServerError)
			return
		}
		leaderboard = append(leaderboard, entry)
	}

	if leaderboard == nil {
		leaderboard = []LeaderboardEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

// ExportJSONL handles GET /api/v1/export/jsonl
// Returns all approved entries as JSONL (one JSON object per line)
// This is the format used for LLM fine-tuning and NLP research.
func ExportJSONL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional language filter e.g. ?lang=luo
	lang := r.URL.Query().Get("lang")

	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Err() error
		Close()
	}
	var err error

	if lang != "" {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT e.source_text, e.english, e.explanation,
				e.english_equivalent, e.example_source,
				e.example_english, e.category, e.dialect,
				e.part_of_speech, e.notes, l.code, l.name
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = 'approved' AND l.code = $1
			ORDER BY e.created_at ASC`,
			lang,
		)
	} else {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT e.source_text, e.english, e.explanation,
				e.english_equivalent, e.example_source,
				e.example_english, e.category, e.dialect,
				e.part_of_speech, e.notes, l.code, l.name
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = 'approved'
			ORDER BY e.created_at ASC`,
		)
	}

	if err != nil {
		http.Error(w, "could not fetch entries", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// JSONL — one JSON object per line, no wrapping array
	// Each line is independently parseable — important for
	// streaming large datasets into training pipelines
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", "attachment; filename=loba-corpus.jsonl")

	encoder := json.NewEncoder(w)
	for rows.Next() {
		var rec struct {
			SourceText        string `json:"source_text"`
			English           string `json:"english"`
			Explanation       string `json:"explanation,omitempty"`
			EnglishEquivalent string `json:"english_equivalent,omitempty"`
			ExampleSource     string `json:"example_source,omitempty"`
			ExampleEnglish    string `json:"example_english,omitempty"`
			Category          string `json:"category,omitempty"`
			Dialect           string `json:"dialect,omitempty"`
			PartOfSpeech      string `json:"part_of_speech,omitempty"`
			Notes             string `json:"notes,omitempty"`
			LanguageCode      string `json:"language_code"`
			LanguageName      string `json:"language_name"`
		}
		err := rows.Scan(
			&rec.SourceText, &rec.English, &rec.Explanation,
			&rec.EnglishEquivalent, &rec.ExampleSource,
			&rec.ExampleEnglish, &rec.Category, &rec.Dialect,
			&rec.PartOfSpeech, &rec.Notes,
			&rec.LanguageCode, &rec.LanguageName,
		)
		if err != nil {
			continue
		}
		encoder.Encode(rec)
	}
}

// ExportCSV handles GET /api/v1/export/csv
// Returns all approved entries as a CSV file for
// spreadsheet analysis and academic research use.
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lang := r.URL.Query().Get("lang")

	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Err() error
		Close()
	}
	var err error

	if lang != "" {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT e.source_text, e.english, e.explanation,
				e.english_equivalent, e.example_source,
				e.example_english, e.category, e.dialect,
				e.part_of_speech, e.notes, l.code, l.name
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = 'approved' AND l.code = $1
			ORDER BY e.created_at ASC`,
			lang,
		)
	} else {
		rows, err = db.DB.Query(
			context.Background(),
			`SELECT e.source_text, e.english, e.explanation,
				e.english_equivalent, e.example_source,
				e.example_english, e.category, e.dialect,
				e.part_of_speech, e.notes, l.code, l.name
			FROM entries e
			JOIN languages l ON l.id = e.language_id
			WHERE e.status = 'approved'
			ORDER BY e.created_at ASC`,
		)
	}

	if err != nil {
		http.Error(w, "could not fetch entries", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=loba-corpus.csv")

	// Write header row
	w.Write([]byte("source_text,english,explanation,english_equivalent,example_source,example_english,category,dialect,part_of_speech,notes,language_code,language_name\n"))

	for rows.Next() {
		var sourceText, english, explanation, englishEquivalent string
		var exampleSource, exampleEnglish, category, dialect string
		var partOfSpeech, notes, langCode, langName string

		err := rows.Scan(
			&sourceText, &english, &explanation,
			&englishEquivalent, &exampleSource,
			&exampleEnglish, &category, &dialect,
			&partOfSpeech, &notes, &langCode, &langName,
		)
		if err != nil {
			continue
		}

		// Escape any commas or quotes within fields
		row := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			csvEscape(sourceText), csvEscape(english),
			csvEscape(explanation), csvEscape(englishEquivalent),
			csvEscape(exampleSource), csvEscape(exampleEnglish),
			csvEscape(category), csvEscape(dialect),
			csvEscape(partOfSpeech), csvEscape(notes),
			csvEscape(langCode), csvEscape(langName),
		)
		w.Write([]byte(row))
	}
}

// csvEscape wraps a field in quotes if it contains
// commas, quotes, or newlines — standard CSV escaping
func csvEscape(s string) string {
	if s == "" {
		return ""
	}
	// Escape any existing double quotes by doubling them
	escaped := `"` + replaceAll(s, `"`, `""`) + `"`
	return escaped
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

// GetLanguages handles GET /api/v1/languages
// Returns all languages in the corpus with their entry counts.
// Powers the language selector and the about page.
func GetLanguages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(
		context.Background(),
		`SELECT l.id, l.code, l.name, l.script, l.region,
			COUNT(e.id) FILTER (WHERE e.status = 'approved') as entry_count
		FROM languages l
		LEFT JOIN entries e ON e.language_id = l.id
		GROUP BY l.id, l.code, l.name, l.script, l.region
		ORDER BY entry_count DESC`,
	)
	if err != nil {
		http.Error(w, "could not fetch languages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Language struct {
		ID         string `json:"id"`
		Code       string `json:"code"`
		Name       string `json:"name"`
		Script     string `json:"script"`
		Region     string `json:"region"`
		EntryCount int    `json:"entry_count"`
	}

	var languages []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.ID, &l.Code, &l.Name, &l.Script, &l.Region, &l.EntryCount); err != nil {
			continue
		}
		languages = append(languages, l)
	}

	if languages == nil {
		languages = []Language{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(languages)
}
