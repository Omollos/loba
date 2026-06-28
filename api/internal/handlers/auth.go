package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Omollos/loba/api/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

// GitHubUser represents the response from GitHub's user API
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
}

// Claims represents the JWT token payload
type Claims struct {
	ContributorID string `json:"contributor_id"`
	Username      string `json:"username"`
	IsReviewer    bool   `json:"is_reviewer"`
	jwt.RegisteredClaims
}

// GitHubLogin redirects the user to GitHub's OAuth login page
func GitHubLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&scope=user:email",
		clientID,
	)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// GitHubCallback handles the callback from GitHub after login
func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Exchange the code for an access token
	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use the access token to get the user's GitHub profile
	githubUser, err := getGitHubUser(accessToken)
	if err != nil {
		http.Error(w, "failed to get GitHub user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get or create the contributor in our database
	var contributorID string
	var isReviewer bool

	err = db.DB.QueryRow(
		context.Background(),
		`INSERT INTO contributors (username, email, github_handle, github_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (github_id) DO UPDATE
		SET username = EXCLUDED.username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url
		RETURNING id, is_reviewer`,
		githubUser.Login,
		githubUser.Email,
		githubUser.Login,
		githubUser.ID,
		githubUser.AvatarURL,
	).Scan(&contributorID, &isReviewer)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a JWT session token
	token, err := generateToken(contributorID, githubUser.Login, isReviewer)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// During local dev the frontend is opened as a file
	// After deployment this becomes the Vercel URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5500"
	}

	if isReviewer {
		http.Redirect(w, r,
			fmt.Sprintf("%s/review.html#token=%s", frontendURL, token),
			http.StatusTemporaryRedirect,
		)
	} else {
		http.Redirect(w, r,
			fmt.Sprintf("%s/dictionary.html#not-reviewer", frontendURL),
			http.StatusTemporaryRedirect,
		)
	}
}

// GetMe returns the current logged-in user's info
// Used by the review page to verify the session is valid
func GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := validateTokenFromHeader(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contributor_id": claims.ContributorID,
		"username":       claims.Username,
		"is_reviewer":    claims.IsReviewer,
	})
}

// RequireReviewer is middleware that protects sensitive endpoints
// It checks for a valid JWT token with is_reviewer = true
func RequireReviewer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := validateTokenFromHeader(r)
		if err != nil || !claims.IsReviewer {
			http.Error(w, "reviewer access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// helper functions
func exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	url := fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientID, clientSecret, code,
	)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Error != "" {
		return "", fmt.Errorf("github error: %s", result.Error)
	}
	return result.AccessToken, nil
}

func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user GitHubUser
	json.Unmarshal(body, &user)
	return &user, nil
}

func generateToken(contributorID, username string, isReviewer bool) (string, error) {
	secret := os.Getenv("SESSION_SECRET")
	claims := Claims{
		ContributorID: contributorID,
		Username:      username,
		IsReviewer:    isReviewer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateTokenFromHeader(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 {
		return nil, fmt.Errorf("missing token")
	}

	tokenString := authHeader[7:] // strip "Bearer "
	secret := os.Getenv("SESSION_SECRET")

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
