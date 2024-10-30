package main

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
)

// SpotifyHandlers.go
// This file contains the error handling functions for Spotify API requests.

// handleGenericFetch makes a generic fetch request to the Spotify API.
//
// Parameters:
// - endpoint: The endpoint to fetch data from.
// - accessToken: Spotify access token.
// - queryParams: Query parameters as a map of strings.
// - bodyArgs: Body arguments as a map of strings.
//
// Returns:
// - The data fetched from the endpoint.
//
// Type	parameters:
// - T: The type of data expected to be fetched from the endpoint.
func handleGenericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) T {
	data, err := genericFetch[T](endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to fetch data from %s: %v", endpoint, err)
		var empty T
		return empty
	}
	return data
}

// handleGenericPut
//
// Parameters:
// - endpoint: The endpoint to send data to
// - accessToken: Spotify access token
// - queryParams: Query parameters
// - bodyArgs: Body arguments
//
// Returns:
// - statusCode: The status code of the request
// - err: The error message
func handleGenericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPut(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to put data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}

// handleGenericPost makes a generic POST request to the Spotify API.
//
// Parameters:
// - endpoint: The endpoint to post data to.
// - accessToken: Spotify access token.
// - queryParams: Query parameters as a map of strings.
// - bodyArgs: Body arguments as a map of strings.
//
// Returns:
// - The status code of the request.
// - An error if the request failed.
func handleGenericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPost(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to post data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}

// handleFetchPlayback fetches the playback state from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
//
// Returns:
// - The playback state.
func handleFetchPlayback(token string) tea.Cmd {
	return func() tea.Msg {
		infoLogger.Println("Fetching playback state")
		state := handleGenericFetch[PlaybackState]("/me/player", token, nil, nil)
		return state
	}
}

// handleFetchLibrary fetches the user's library from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
// - listDetail: The type of library to fetch (album or playlist).
// - height: The number of items to fetch.
//
// Returns:
// - The fetched library.
func handleFetchLibrary(token string, listDetail string, height, offset int) tea.Cmd {
	return func() tea.Msg {
		height = int(math.Min(float64(height), 50))
		if listDetail == "album" {
			albums := handleGenericFetch[SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			return albums
		} else {
			playlist := handleGenericFetch[SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			return playlist
		}
	}
}

// handleFetchPlaylist fetches a playlist from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
// - playlistID: The ID of the playlist to fetch.
//
// Returns:
// - The fetched playlist.
func handleGetLibraryTotal(token string, listDetail string) tea.Cmd {
	return func() tea.Msg {
		if listDetail == "album" {
			albums := handleGenericFetch[SpotifyAlbum]("/me/albums", token, map[string]string{"limit": "1"}, nil)
			return albums.Total
		} else {
			playlist := handleGenericFetch[SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": "1"}, nil)
			return playlist.Total
		}
	}
}
