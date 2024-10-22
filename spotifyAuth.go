package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// SpotifyAuth.go
// This file contains the functions for authenticating with Spotify.
// The functions in this file are used to open the Spotify login page in the
// user's default browser, get the authorization code from the callback URL,
// and exchange the authorization code for an access token.

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
		fmt.Println()
		err = exec.Command("open", authURL).Start() // macOS
		if err != nil {
			log.Fatalf("Failed to open login page: %v", err)
		}
	}
}

// Gets the authorization code from the callback URL. Makes a script that closes the window afterwards.
func GetCodeFromCallback() string {
	var code string
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code = r.URL.Query().Get("code")
		// Add JavaScript to automatically close the window
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
		time.Sleep(500 * time.Millisecond) // Wait for code
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SpotifyTokenResponse{}, err
	}

	var tokenResp SpotifyTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return SpotifyTokenResponse{}, err
	}

	return tokenResp, nil
}
