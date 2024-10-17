package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	SPOTIFY_AUTH_URL  = "https://accounts.spotify.com/authorize"
	SPOTIFY_TOKEN_URL = "https://accounts.spotify.com/api/token"
	REDIRECT_URI      = "http://localhost:8080/callback"
)

var (
	clientID, clientSecret string
	accessToken            string
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	clientID = os.Getenv("SPOTIFY_ID")
	clientSecret = os.Getenv("SPOTIFY_SECRET")

	openLoginPage()

	http.HandleFunc("/callback", callbackHandler)

	fmt.Println("Server running at http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func openLoginPage() {
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-private user-read-email",
		SPOTIFY_AUTH_URL, clientID, url.QueryEscape(REDIRECT_URI),
	)

	// Open the URL in the default browser
	err := exec.Command("xdg-open", authURL).Start() // For Linux
	if err != nil {
		err = exec.Command("open", authURL).Start() // For macOS
		if err != nil {
			log.Fatalf("Failed to open login page: %v", err)
		}
	}
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in callback", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	token, err := getSpotifyToken(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	accessToken = token.AccessToken
	user, err := getSpotifyUser(ctx, accessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("User Data: %+v\n", user)

	fmt.Fprintln(w, `<script>window.close();</script>`)
}

type SpotifyUser struct {
	DisplayName  string `json:"display_name"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Total int `json:"total"`
	} `json:"followers"`
	ID string `json:"id"`
}

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func getSpotifyToken(ctx context.Context, code string) (SpotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", REDIRECT_URI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SPOTIFY_TOKEN_URL, strings.NewReader(data.Encode()))
	if err != nil {
		return SpotifyTokenResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return SpotifyTokenResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SpotifyTokenResponse{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SpotifyTokenResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResponse SpotifyTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return SpotifyTokenResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return tokenResponse, nil
}

func getSpotifyUser(ctx context.Context, accessToken string) (*SpotifyUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var user SpotifyUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &user, nil
}
