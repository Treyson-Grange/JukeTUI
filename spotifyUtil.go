package main

import (
	"fmt"
	"net/url"
)

// SpotifyUtil.go
// This file holds any utils that interact with the inner workings of
// our spotify system.

const SPOTIFY_API = "https://api.spotify.com/v1"

// Edit endpoint with queryparams
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
