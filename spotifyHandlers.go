package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// SpotifyHandlers.go
// This file contains the error handling functions for Spotify API requests.

// Generic fetch/get function for Spotify API requests.
func handleGenericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) T {
	data, err := genericFetch[T](endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to fetch data from %s: %v", endpoint, err)
		var empty T
		return empty
	}
	return data
}

// Generic PUT function for Spotify API requests.
func handleGenericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPut(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to put data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}

// Generic POST function for Spotify API requests.
func handleGenericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := genericPost(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		errorLogger.Printf("Failed to post data to %s: %v", endpoint, err)
		return statusCode, err
	}
	return statusCode, nil
}

// Handler to fetch the playback state.
func handleFetchPlayback(token string) tea.Cmd {
	return func() tea.Msg {
		state := handleGenericFetch[PlaybackState]("/me/player", token, nil, nil)
		return state
	}
}

// Handler to fetch the library.
func handleFetchLibrary(token string, listDetail string, height int) tea.Cmd {
	return func() tea.Msg {
		if listDetail == "album" {
			albums := handleGenericFetch[SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height)}, nil)
			return albums
		} else {
			playlist := handleGenericFetch[SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": fmt.Sprintf("%d", height)}, nil)
			return playlist
		}
	}
}
