package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GenerateUsername creates a unique username based on the provided name.
func GenerateUsername(name string) string {
	// Remove spaces and lowercase the name.
	base := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	// Append a random string to ensure uniqueness.
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if random generation fails.
		return fmt.Sprintf("%s%d", base, time.Now().Unix())
	}
	return fmt.Sprintf("%s%s", base, hex.EncodeToString(b))
}

// Google OAuth2 configuration.
var googleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// GetGoogleOAuthURL returns the URL for Google OAuth login.
func GetGoogleOAuthURL() string {
	state := generateState() // Optionally generate a random state string for security.
	return googleOAuthConfig.AuthCodeURL(state)
}

// generateState creates a random state string.
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GoogleUserInfo holds data returned from Google.
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GetGoogleUserInfo exchanges code for a token and fetches user info from Google.
func GetGoogleUserInfo(code string) (*GoogleUserInfo, error) {
	token, err := googleOAuthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := googleOAuthConfig.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}
