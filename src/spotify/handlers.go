package spotify

import (
	"fmt"
	"math"

	"github.com/treyson-grange/JukeTUI/src/types"
	"github.com/treyson-grange/JukeTUI/src/util"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================================================
// ===== spotifyHandlers.go | Various requests to Spotify API =====
// ================================================================

// HandleGenericFetch handles and error checks a generic fetch
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
func HandleGenericFetch[T any](endpoint, accessToken string, queryParams, bodyArgs map[string]string) T {
	data, err := GenericFetch[T](endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		util.ErrorLogger.Printf("Failed to fetch data from %s: %v", endpoint, err)
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
func HandleGenericPut(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := GenericPut(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		util.ErrorLogger.Printf("Failed to put data to %s: %v", endpoint, err)
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
func HandleGenericPost(endpoint, accessToken string, queryParams, bodyArgs map[string]string) (int, error) {
	statusCode, err := GenericPost(endpoint, accessToken, queryParams, bodyArgs)
	if err != nil {
		util.ErrorLogger.Printf("Failed to post data to %s: %v", endpoint, err)
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
func HandleFetchPlayback(token string) tea.Cmd {
	return func() tea.Msg {
		state := HandleGenericFetch[types.PlaybackState]("/me/player", token, nil, nil)
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
func HandleFetchLibrary(favorites []types.LibraryFavorite, token string, listDetail string, height, offset int) tea.Cmd {
	return func() tea.Msg {
		height = int(math.Min(float64(height), 50))
		if listDetail == "album" {
			albums := HandleGenericFetch[types.SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			removed := 0

			for _, item := range albums.Items {
				for _, favorite := range favorites {
					if item.Album.URI == favorite.URI {
						removed++
					}
				}
			}
			albums = HandleGenericFetch[types.SpotifyAlbum]("/me/albums", token, map[string]string{"limit": fmt.Sprintf("%d", height+removed), "offset": fmt.Sprintf("%d", offset)}, nil)
			filteredItems := make([]struct{ types.SpotifyAlbumItem }, 0, len(albums.Items))
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
			playlist := HandleGenericFetch[types.SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": fmt.Sprintf("%d", height), "offset": fmt.Sprintf("%d", offset)}, nil)
			filteredItems := make([]types.SpotifyPlaylistItem, 0, len(playlist.Items))
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

// handleGetLibraryTotal fetches albums or playlists from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
// - listDetail: The type of playlist to fetch (album or playlist).
//
// Returns:
// - The fetched playlist or album.
func HandleGetLibraryTotal(token string, listDetail string) tea.Cmd {
	return func() tea.Msg {
		if listDetail == "album" {
			albums := HandleGenericFetch[types.SpotifyAlbum]("/me/albums", token, map[string]string{"limit": "1"}, nil)
			return albums.Total
		} else {
			playlist := HandleGenericFetch[types.SpotifyPlaylist]("/me/playlists", token, map[string]string{"limit": "1"}, nil)
			return playlist.Total
		}
	}
}

// HandleGetQueue fetches the user's queue from the Spotify API.
//
// Parameters:
// - token: Spotify access token.
func HandleGetQueue(token string) tea.Cmd {
	return func() tea.Msg {
		queue := HandleGenericFetch[types.Queue]("/me/player/queue", token, nil, nil)
		return queue
	}
}
