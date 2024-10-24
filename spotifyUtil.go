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

const SPOTIFY_API = "https://api.spotify.com/v1" //TODO: where to put this const?

// createEndpoint creates a full endpoint URL with query parameters.
//
// Parameters:
// - endpoint: the endpoint to fetch data from
// - queryParams: the query parameters to include in the request
//
// Returns:
// - string: the full endpoint URL, with query parameters if any.
func createEndpoint(endpoint string, queryParams map[string]string) string { // TODO: Test this lol.
	endpoint = fmt.Sprintf("%s%s", SPOTIFY_API, endpoint)

	if len(queryParams) == 0 { // Empty map is not nil
		return endpoint
	}

	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	return fmt.Sprintf("%s?%s", endpoint, query.Encode())
}

// Check if the token is expired
//
// Parameters:
// - s: the SpotifyTokenResponse to check
//
// Returns:
// - bool: true if the token is expired, false otherwise
func (s *SpotifyTokenResponse) IsExpired() bool {
	return time.Now().After(time.Now().Add(time.Duration(s.ExpiresIn) * time.Second))
}

// CheckTokenExpiryCmd refreshes the token if it has expired.
//
// Parameters:
// - m: the Model to check
//
// Returns:
// - tea.Cmd: a command to refresh the token if it has expired
func CheckTokenExpiryCmd(m Model) tea.Cmd {
	if time.Now().After(m.tokenExpiresAt) {
		return refreshSpotifyTokenCmd(m.refreshToken, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	}
	return nil
}
