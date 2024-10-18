package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const (
	SPOTIFY_AUTH_URL  = "https://accounts.spotify.com/authorize"
	SPOTIFY_TOKEN_URL = "https://accounts.spotify.com/api/token"
	REDIRECT_URI      = "http://localhost:8080/callback"
)

func OpenLoginPage(clientID string) {
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s",
		SPOTIFY_AUTH_URL, clientID, url.QueryEscape(REDIRECT_URI),
		url.QueryEscape("user-read-private user-read-email user-read-playback-state"),
	)
	

	err := exec.Command("xdg-open", authURL).Start() // Linux
	if err != nil {
		err = exec.Command("open", authURL).Start() // macOS
		if err != nil {
			log.Fatalf("Failed to open login page: %v", err)
		}
	}
}

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

func GetRecommendations(accessToken, genre string, limit int) (SpotifyRecommendations, error) {
	endpoint := fmt.Sprintf("https://api.spotify.com/v1/recommendations?seed_genres=%s&limit=%d", genre, limit)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return SpotifyRecommendations{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return SpotifyRecommendations{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SpotifyRecommendations{}, err
	}

	var recs SpotifyRecommendations
	if err := json.Unmarshal(body, &recs); err != nil {
		return SpotifyRecommendations{}, err
	}

	return recs, nil
}

// TODO: add possibility to pass in additional query parameters like in reccomendations
func genericFetch[T any](endpoint, accessToken string) (T, error) {
	var result T

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return result, fmt.Errorf("missing required permissions")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}