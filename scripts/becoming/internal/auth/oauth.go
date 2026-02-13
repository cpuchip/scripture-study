package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// OAuthConfig holds Google OAuth configuration.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string // e.g., "https://ibeco.me/auth/google/callback"
}

// OAuthConfigFromEnv loads Google OAuth config from environment variables.
// Returns nil if not configured (Google sign-in will be disabled).
func OAuthConfigFromEnv() *OAuthConfig {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		return nil
	}
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/google/callback"
	}
	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

// stateStore holds CSRF state tokens with expiry.
var stateStore = struct {
	sync.Mutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	stateStore.Lock()
	defer stateStore.Unlock()

	// Clean expired states
	now := time.Now()
	for k, exp := range stateStore.tokens {
		if now.After(exp) {
			delete(stateStore.tokens, k)
		}
	}

	stateStore.tokens[state] = now.Add(5 * time.Minute)
	return state, nil
}

func validateState(state string) bool {
	stateStore.Lock()
	defer stateStore.Unlock()

	exp, ok := stateStore.tokens[state]
	if !ok {
		return false
	}
	delete(stateStore.tokens, state)
	return time.Now().Before(exp)
}

// GoogleLogin handles GET /auth/google/login — redirects to Google's consent screen.
func (h *Handlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.OAuth == nil {
		http.Error(w, "Google sign-in is not configured", http.StatusNotFound)
		return
	}

	state, err := generateState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	params := url.Values{
		"client_id":     {h.OAuth.ClientID},
		"redirect_uri":  {h.OAuth.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}

	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), http.StatusFound)
}

// GoogleCallback handles GET /auth/google/callback — exchanges code for tokens and creates session.
func (h *Handlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.OAuth == nil {
		http.Error(w, "Google sign-in is not configured", http.StatusNotFound)
		return
	}

	// Validate state
	state := r.URL.Query().Get("state")
	if !validateState(state) {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	// Check for error from Google
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		log.Printf("google oauth error: %s", errParam)
		http.Redirect(w, r, "/login?error=oauth_denied", http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenResp, err := exchangeCode(h.OAuth, code)
	if err != nil {
		log.Printf("google oauth: code exchange failed: %v", err)
		http.Redirect(w, r, "/login?error=oauth_failed", http.StatusFound)
		return
	}

	// Get user info from Google
	userInfo, err := getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("google oauth: failed to get user info: %v", err)
		http.Redirect(w, r, "/login?error=oauth_failed", http.StatusFound)
		return
	}

	// Create or link user
	user, err := h.DB.CreateOAuthUser(userInfo.Email, userInfo.Name, userInfo.Picture, "google", userInfo.Sub)
	if err != nil {
		log.Printf("google oauth: failed to create/link user: %v", err)
		http.Redirect(w, r, "/login?error=oauth_failed", http.StatusFound)
		return
	}

	h.DB.TouchUserLogin(user.ID)

	// Create session
	session, err := h.DB.CreateSession(user.ID, r.UserAgent(), clientIP(r))
	if err != nil {
		log.Printf("google oauth: failed to create session: %v", err)
		http.Redirect(w, r, "/login?error=session_failed", http.StatusFound)
		return
	}

	h.setSessionCookie(w, session.ID)

	// Redirect to the app
	http.Redirect(w, r, "/today", http.StatusFound)
}

// --- Google API helpers ---

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func exchangeCode(cfg *OAuthConfig, code string) (*googleTokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"redirect_uri":  {cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}
	return &tokenResp, nil
}

func getGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo failed (%d): %s", resp.StatusCode, body)
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing userinfo: %w", err)
	}
	return &info, nil
}
