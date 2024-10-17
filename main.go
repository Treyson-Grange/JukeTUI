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
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const SPOTIFY_TOKEN_URL = "https://accounts.spotify.com/api/token"

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := getSpotifyToken(ctx, SPOTIFY_TOKEN_URL, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Spotify Token Response:", token.AccessToken)
	fmt.Println(getSpotifyUser(ctx, token.AccessToken, "treyson.grange"))
}

type SpotifyUser struct {
	DisplayName  string `json:"display_name"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  interface{} `json:"href"`
		Total int         `json:"total"`
	} `json:"followers"`
	Href   string `json:"href"`
	ID     string `json:"id"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

func getSpotifyUser(ctx context.Context, token, userID string) (*SpotifyUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.spotify.com/v1/users/%s", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func getSpotifyToken(ctx context.Context, tokenURL, clientID, clientSecret string) (SpotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
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
