package main

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================================================
// ===== spotifyHandlers.go | Various requests to Spotify API =====
// ================================================================

// handleGenericFetch handles and error checks a generic fetch
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

// handleGenericPut handles and error checks a generic PUT request.
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

// handleGenericPost handles and error checks a generic POST request.
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

// handleFetchPlayback handles fetching and error checking of the playback state.
//
// Parameters:
// - token: Spotify access token.
//
// Returns:
// - The playback state.
func handleFetchPlayback(token string) tea.Cmd {
	return func() tea.Msg {
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
func handleFetchLibrary(favorites []LibraryFavorite, token string, listDetail string, height, offset int) tea.Cmd {
	return func() tea.Msg {
		height = int(math.Min(float64(height), 50))
		if listDetail == "album" {
			albums := handleGenericFetch[SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			removed := 0

			for _, item := range albums.Items {
				for _, favorite := range favorites {
					if item.Album.URI == favorite.URI {
						removed++
					}
				}
			}
			albums = handleGenericFetch[SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height+removed), "offset": fmt.Sprintf("%d", offset)}, nil)
			filteredItems := make([]struct{ SpotifyAlbumItem }, 0, len(albums.Items))
			for _, item := range albums.Items {
				isFavorite := false
				for _, favorite := range favorites {
					if item.Album.URI == favorite.URI {
						isFavorite = true
						break
					}
				}
				if !isFavorite {
					filteredItems = append(filteredItems, item)
				}
			}
			albums.Items = filteredItems
			return albums
		} else {
			playlist := handleGenericFetch[SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			filteredItems := make([]SpotifyPlaylistItem, 0, len(playlist.Items))
			favoriteURIs := make(map[string]struct{})
			for _, favorite := range favorites {
				favoriteURIs[favorite.URI] = struct{}{}
			}
			for _, item := range playlist.Items {
				if _, found := favoriteURIs[item.URI]; !found {
					filteredItems = append(filteredItems, item)
				}
			}
			playlist.Items = filteredItems
			return playlist
		}
	}
}

// handleFetchPlaylist fetches a playlist from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
// - listDetail: The type of playlist to fetch (album or playlist).
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

// handleFetchPlaylist fetches a playlist from the Spotify API.
//
// Parameters:
// - token: Spotify access token.

func handleGetQueue(token string) tea.Cmd {
	return func() tea.Msg {
		queue := handleGenericFetch[Queue]("/me/player/queue", token, nil, nil)
		return queue
	}
}
