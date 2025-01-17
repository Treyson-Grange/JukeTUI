package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================================================
// ===== spotifyAuth.go | Login and authenticate with Spotify =====
// ================================================================

const (
	SPOTIFY_AUTH_URL  = "https://accounts.spotify.com/authorize"
	SPOTIFY_TOKEN_URL = "https://accounts.spotify.com/api/token"
	REDIRECT_URI      = "http://localhost:8080/callback"
)

var SPOTIFY_PERMS = []string{
	"user-read-private",
	"user-read-email",
	"user-read-playback-state",
	"user-modify-playback-state",
	"playlist-read-private",
	"user-library-read",
}

// Opens the login page on the users primary browser, prompting for login.
func OpenLoginPage(clientID string) {
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s",
		SPOTIFY_AUTH_URL, clientID, url.QueryEscape(REDIRECT_URI),
		url.QueryEscape(strings.Join(SPOTIFY_PERMS, " ")),
	)

	err := exec.Command("xdg-open", authURL).Start() // Linux
	if err != nil {
		err = exec.Command("open", authURL).Start() // macOS
		if err != nil {
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Start() // Windows
			if err != nil {
				log.Fatalf("Failed to open login page: %v", err)
			}
		}
	}
}

// Gets the authorization code from the callback URL. Makes a script that closes the window afterwards.
func GetCodeFromCallback() string {
	var code string
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code = r.URL.Query().Get("code")
		htmlResponse := `
            <html>
                <body>
                    <p>Login successful! You can close this window now.</p>
                    <script type="text/javascript">
                        window.onload = function() {
                            window.open('','_self').close();
                        }
                    </script>
                </body>
            </html>
        `
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, htmlResponse)
	})

	go http.ListenAndServe(":8080", nil)

	for code == "" {
		time.Sleep(500 * time.Millisecond)
	}
	return code
}

// Given the client ID, client secret, and authorization code, returns the Spotify token response.
func GetSpotifyToken(ctx context.Context, clientID, clientSecret, code string) (SpotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", REDIRECT_URI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SPOTIFY_TOKEN_URL, strings.NewReader(data.Encode()))
	if err != nil {
		return SpotifyTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return SpotifyTokenResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SpotifyTokenResponse{}, err
	}

	var tokenResp SpotifyTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return SpotifyTokenResponse{}, err
	}

	return tokenResp, nil
}

// refreshSpotifyTokenCmd returns a command that refreshes the Spotify token.
func refreshSpotifyTokenCmd(refreshToken, clientID, clientSecret string) tea.Cmd {
	return func() tea.Msg {
		newToken, err := RefreshSpotifyToken(refreshToken, clientID, clientSecret)
		if err != nil {
			return fmt.Errorf("failed to refresh token: %v", err)
		}
		return newToken
	}
}

// RefreshSpotifyToken refreshes the Spotify token using the refresh token.
func RefreshSpotifyToken(refreshToken, clientID, clientSecret string) (SpotifyTokenResponse, error) {
	reqBody := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}.Encode()

	req, _ := http.NewRequest("POST", SPOTIFY_TOKEN_URL, strings.NewReader(reqBody))
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SpotifyTokenResponse{}, err
	}
	defer resp.Body.Close()

	var tokenRes SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return SpotifyTokenResponse{}, err
	}
	return tokenRes, nil
}
