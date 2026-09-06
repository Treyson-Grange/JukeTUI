package types

import (
	"time"
)

// Model for the application state
type Model struct {
	//Playback state, including track info, playback status, etc.
	State PlaybackState

	//Spotify web API access token. Lasts for 1 hour.
	Token string

	//Spotify web API refresh token. Used to get a new access token when the current one is close to expiration.
	RefreshToken string

	//Time when the current access token expires.
	TokenExpiresAt time.Time

	//Error message, if any
	ErrMsg string

	//Whether or not we're currently fetching access token initially
	Loading bool

	//List detail, either "album" or "playlist".
	ListDetail string

	//Cursor for the list of albums/playlists.
	Cursor int

	//List of albums/playlists.
	LibraryList []LibraryItem

	//Height of the list of albums/playlists.
	Height int

	//Progress of current track in ms
	ProgressMs int

	// Album cover image as string
	Image string

	// Offset for pagination of albums/playlists
	Offset int

	// Total Library Items
	ApiTotal int

	// Favorites list (albums)
	FavoriteAlbums []LibraryFavorite

	// Favorites list (playlists)
	FavoritePlaylists []LibraryFavorite

	// Queue list
	Queue Queue //This isnt what itll be
}

// SpotifyTokenResponse struct for parsing the access token response.
type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// PlaybackState struct for parsing the playback state response.
type PlaybackState struct {
	Device struct {
		ID string `json:"id"`
	} `json:"device"`
	ShuffleState bool   `json:"shuffle_state"`
	RepeatState  string `json:"repeat_state"`
	Context      struct {
		URI string `json:"uri"`
	} `json:"context"`
	ProgressMs int `json:"progress_ms"`
	Item       struct {
		Album struct {
			AlbumType string `json:"album_type"`
			Artists   []struct {
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"album"`
		Artists []struct {
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		DurationMs int    `json:"duration_ms"`
		Href       string `json:"href"`
		ID         string `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		URI        string `json:"uri"`
	} `json:"item"`
	IsPlaying bool `json:"is_playing"`
}

// SpotifyAlbum struct for parsing the albums response.
type SpotifyAlbum struct {
	Items []struct {
		SpotifyAlbumItem
	}
	Total int `json:"total"`
}

// SpotifyAlbumItem struct for parsing the album item response.
type SpotifyAlbumItem struct {
	Album struct {
		Name    string `json:"name"`
		URI     string `json:"uri"`
		Artists []struct {
			Name string `json:"name"`
		}
	}
}

// SpotifyPlaylist struct for parsing the playlists response.
type SpotifyPlaylist struct {
	Items []SpotifyPlaylistItem `json:"items"`
	Total int                   `json:"total"`
}

type SpotifyPlaylistItem struct {
	Name  string `json:"name"`
	URI   string `json:"uri"`
	Owner struct {
		DisplayName string `json:"display_name"`
	}
}

// LibraryItem struct for storing album/playlist information.
type LibraryItem struct {
	Name     string
	Artist   string
	URI      string
	Favorite bool
}

// LibraryFavorite struct for storing favorite album/playlist information.
type LibraryFavorite struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	URI    string `json:"URI"`
}

// Queue struct for storing the queue of songs.
type Queue struct {
	Queue []QueueItem `json:"queue"`
}

// QueueItem struct for storing the queue item information.
type QueueItem struct {
	Href    string `json:"href"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	URI     string `json:"uri"`
	IsLocal bool   `json:"is_local"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

// IsExpired checks if the token is expired
func (s *SpotifyTokenResponse) IsExpired() bool {
	return time.Now().After(time.Now().Add(time.Duration(s.ExpiresIn) * time.Second))
}

// PlaybackMsg tells the update to fetch playback state.
type PlaybackMsg struct{}

// ProgressMsg tells the update to update the progress of the current track.
type ProgressMsg struct{}
