package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthHandler struct {
	userRepo    *postgres.UserRepository
	jwtManager  *auth.JWTManager
	googleOAuth *oauth2.Config
	stateStore  map[string]time.Time // Simple state storage (use Redis in production)
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewOAuthHandler(userRepo *postgres.UserRepository, jwtManager *auth.JWTManager, clientID, clientSecret, redirectURL string) *OAuthHandler {
	return &OAuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		googleOAuth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		stateStore: make(map[string]time.Time),
	}
}

// generateStateToken generates a random state token for CSRF protection
func (h *OAuthHandler) generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state with expiration (5 minutes)
	h.stateStore[state] = time.Now().Add(5 * time.Minute)

	// Clean up expired states
	go h.cleanupExpiredStates()

	return state, nil
}

// cleanupExpiredStates removes expired state tokens
func (h *OAuthHandler) cleanupExpiredStates() {
	now := time.Now()
	for state, expiry := range h.stateStore {
		if now.After(expiry) {
			delete(h.stateStore, state)
		}
	}
}

// verifyState verifies the state token
func (h *OAuthHandler) verifyState(state string) bool {
	expiry, exists := h.stateStore[state]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(h.stateStore, state)
		return false
	}

	delete(h.stateStore, state)
	return true
}

// GoogleLogin godoc
// @Summary Initiate Google OAuth login
// @Description Redirects user to Google OAuth consent page
// @Tags oauth
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Failure 500 {object} map[string]string
// @Router /auth/google/login [get]
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	state, err := h.generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state token"})
		return
	}

	// Store state in cookie for additional security
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	url := h.googleOAuth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Description Processes the OAuth callback from Google and creates/logs in user
// @Tags oauth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State token"
// @Success 200 {object} types.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/google/callback [get]
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state parameter
	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")

	// Debug logging
	fmt.Printf("OAuth Callback - state from URL: %s\n", state)
	fmt.Printf("OAuth Callback - state from cookie: %s (error: %v)\n", cookieState, err)

	if err != nil || state == "" || state != cookieState {
		fmt.Printf("State validation failed! URL state: %s, Cookie state: %s, Error: %v\n", state, cookieState, err)
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "invalid_state")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	if !h.verifyState(state) {
		fmt.Println("State verification failed - state not in store or expired")
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "expired_state")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	fmt.Println("State verification passed")

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Exchange authorization code for token
	code := c.Query("code")
	if code == "" {
		fmt.Println("No authorization code in request")
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "missing_code")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	fmt.Printf("Exchanging authorization code with Google...\n")
	token, err := h.googleOAuth.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Token exchange failed: %v\n", err)
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "token_exchange_failed")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	fmt.Println("Token exchange successful")

	// Get user info from Google
	fmt.Println("Fetching user info from Google...")
	userInfo, err := h.getUserInfo(token.AccessToken)
	if err != nil {
		fmt.Printf("Failed to get user info: %v\n", err)
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "user_info_failed")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	fmt.Printf("User info retrieved: %s (%s)\n", userInfo.Email, userInfo.Name)

	// Check if email is verified
	if !userInfo.VerifiedEmail {
		fmt.Println("Email not verified")
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "email_not_verified")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	// Check if user exists
	fmt.Printf("Checking if user exists: %s\n", userInfo.Email)
	user, err := h.userRepo.GetByEmail(c.Request.Context(), userInfo.Email)
	if err != nil {
		// User doesn't exist, create new user
		fmt.Println("User not found, creating new user...")
		user = &types.User{
			Email:        userInfo.Email,
			FullName:     userInfo.Name,
			Role:         types.UserRoleViewer, // Default role
			IsActive:     true,
			PasswordHash: "", // No password for OAuth users
		}

		if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
			fmt.Printf("Failed to create user: %v\n", err)
			errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "user_creation_failed")
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}
		fmt.Printf("User created successfully: %s\n", user.ID)
	} else {
		fmt.Printf("Existing user found: %s\n", user.ID)
	}

	// Generate JWT token
	fmt.Println("Generating JWT token...")
	jwtToken, err := h.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		fmt.Printf("Failed to generate JWT: %v\n", err)
		errorURL := fmt.Sprintf("http://localhost:3000/login?error=%s", "token_generation_failed")
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	fmt.Printf("JWT token generated successfully (length: %d)\n", len(jwtToken))

	// Clear password hash before returning
	user.PasswordHash = ""

	// Redirect to frontend with token
	// Frontend will extract token from URL and store it
	frontendURL := fmt.Sprintf("http://localhost:3000/auth/google/callback?token=%s", jwtToken)
	fmt.Printf("Redirecting to frontend: %s\n", frontendURL)
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
	fmt.Println("OAuth callback completed successfully")
}

// getUserInfo fetches user information from Google
func (h *OAuthHandler) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}
