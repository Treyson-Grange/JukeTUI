package spotify

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/treyson-grange/JukeTUI/src/types"

	tea "github.com/charmbracelet/bubbletea"
)

// ===============================================================
// ===== spotifyUtil.go | Utilities to deal with Spotify API =====
// ===============================================================

const SPOTIFY_API = "https://api.spotify.com/v1"

// CreateEndpoint creates a full endpoint URL with query parameters.
//
// Parameters:
// - endpoint: the endpoint to fetch data from
// - queryParams: the query parameters to include in the request
//
// Returns:
// - string: the full endpoint URL, with query parameters if any.
func CreateEndpoint(endpoint string, queryParams map[string]string) string { // TODO: Test this lol.
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

// CheckTokenExpiryCmd refreshes the token if it has expired.
//
// Parameters:
// - m: the Model to check
//
// Returns:
// - tea.Cmd: a command to refresh the token if it has expired
func CheckTokenExpiryCmd(m *types.Model) tea.Cmd {
	if time.Now().After(m.TokenExpiresAt) {
		return RefreshSpotifyTokenCmd(m.RefreshToken, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	}
	return nil
}
