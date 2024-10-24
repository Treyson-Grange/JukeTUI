package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SpotifyUtil.go
// This file holds any utils that interact with the inner workings of
// our spotify system.

const SPOTIFY_API = "https://api.spotify.com/v1"

// TODO: Test this lol.
func createEndpoint(endpoint string, queryParams map[string]string) string {
	endpoint = fmt.Sprintf("%s%s", SPOTIFY_API, endpoint)

	if len(queryParams) == 0 {
		return endpoint
	}

	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	return fmt.Sprintf("%s?%s", endpoint, query.Encode())
}

// Check if the token is expired
func (s *SpotifyTokenResponse) IsExpired() bool {
	return time.Now().After(time.Now().Add(time.Duration(s.ExpiresIn) * time.Second))
}

// CheckTokenExpiryCmd refreshes the token if it has expired.
func CheckTokenExpiryCmd(m Model) tea.Cmd {
	if time.Now().After(m.tokenExpiresAt) {
		return refreshSpotifyTokenCmd(m.refreshToken, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	}
	return nil
}
