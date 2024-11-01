package main

import (
	"time"
)

// Models.go
// This supplies all models that JukeTUI will use in operation, specifically
// on get requests. These will allow us to parse our json correctly, and
// access our data quicker.

type Model struct {
	//Playback state, including track info, playback status, etc.
	//For specifics, see PlaybackState struct in models.go
	state PlaybackState

	//Spotify web API access token. Lasts for 1 hour.
	token string

	//Spotify web API refresh token. Used to get a new access token when the current one is close to expiration.
	refreshToken string

	//Time when the current access token expires.
	tokenExpiresAt time.Time

	//Error message, if any
	errMsg string

	//Whether or not we're currently fetching access token initially
	loading bool

	//Current recommendation, if any.
	reccomendation SpotifyRecommendations

	//List detail, either "album" or "playlist".
	listDetail string

	//Cursor for the list of albums/playlists.
	cursor int

	//List of albums/playlists.
	libraryList []LibraryItem

	//Height of the list of albums/playlists.
	height int

	//Progress of current track in ms
	progressMs int

	// Album cover image as string
	image string

	// Offset for pagination of albums/playlists
	offset int

	// Total Library Items
	apiTotal int

	// Favorites list
	favorites []LibraryFavorite
}

// playbackMsg tells the update to fetch playback state.
type playbackMsg struct{}

// progressMsg tells the update to update the progress of the current track.
type progressMsg struct{}

// SpotifyUser struct for parsing the user's Spotify profile.
// Currently, unused.
type SpotifyUser struct {
	DisplayName  string `json:"display_name"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Total int `json:"total"`
	} `json:"followers"`
	ID string `json:"id"`
}

// SpotifyTokenResponse struct for parsing the access token response.
type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// SpotifyRecommendations struct for parsing the recommendations response.
type SpotifyRecommendations struct {
	Tracks []struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		URI   string `json:"uri"`
		Album struct {
			Image []struct {
				URL string `json:"url"`
			} `json:"images"`
		} `json:"album"`
	} `json:"tracks"`
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
	Items []struct {
		Name  string `json:"name"`
		URI   string `json:"uri"`
		Owner struct {
			DisplayName string `json:"display_name"`
		}
	}
	Total int `json:"total"`
}

// LibraryItem struct for storing album/playlist information.
type LibraryItem struct {
	name   string
	artist string
	uri    string
	favorite bool
}

type LibraryFavorite struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	URI    string `json:"URI"`
}
