package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/treyson-grange/JukeTUI/src/spotify"
	"github.com/treyson-grange/JukeTUI/src/types"
	"github.com/treyson-grange/JukeTUI/src/util"
	"golang.org/x/term"
)

// ==========================================
// ===== main.go | Entry point and loop =====
// ==========================================

// Model embeds types.Model for Bubbletea methods
type Model struct {
	*types.Model
}

func initialModel(token, listDetail string, favoriteAlbums, favoritePlaylists []types.LibraryFavorite) Model {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}

	return Model{
		Model: &types.Model{
			Token:             token,
			ListDetail:        listDetail,
			Height:            height,
			FavoriteAlbums:    favoriteAlbums,
			FavoritePlaylists: favoritePlaylists,
		},
	}
}

const FETCH_TIMER = 2

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		spotify.HandleFetchPlayback(m.Token),
		spotify.HandleGetLibraryTotal(m.Token, m.ListDetail),
		util.ScheduleProgressInc(1*time.Second),
		spotify.HandleFetchLibrary(util.GetFavorites(m.Model), m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(util.GetFavorites(m.Model)), 0),
		spotify.HandleGetQueue(m.Token),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case util.Keybinds["Quit"]:
			return m, tea.Quit

		case util.Keybinds["Play/Pause"]:
			if m.State.IsPlaying {
				spotify.HandleGenericPut("/me/player/pause", m.Token, nil, nil)
			} else {
				spotify.HandleGenericPut("/me/player/play", m.Token, nil, map[string]string{"device_id": m.State.Device.ID})
			}
			return m, nil

		case util.Keybinds["Skip"]:
			spotify.HandleGenericPost("/me/player/next", m.Token, nil, nil)
			return m, nil

		case util.Keybinds["Shuffle"]:
			spotify.HandleGenericPut("/me/player/shuffle", m.Token, map[string]string{"state": fmt.Sprintf("%t", !m.State.ShuffleState)}, nil)
			return m, nil

		case util.Keybinds["Favorites"]:
			file := fmt.Sprintf("data/favorites/%ss.json", m.ListDetail)
			favorites := util.GetFavorites(m.Model)
			if favorites != nil && m.ListDetail == "album" {
				for _, fav := range m.FavoriteAlbums {
					if fav.URI == m.LibraryList[m.Cursor].URI {
						util.RemoveFromJSON(file, fav)
						m.FavoriteAlbums, _ = util.ReadJSON(file)
						return m, spotify.HandleFetchLibrary(m.FavoriteAlbums, m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(m.FavoriteAlbums), m.Offset)
					}
				}
				util.WriteJSONFile(file, types.LibraryFavorite{m.LibraryList[m.Cursor].Name, m.LibraryList[m.Cursor].Artist, m.LibraryList[m.Cursor].URI})
				m.FavoriteAlbums, _ = util.ReadJSON(file)

			} else if favorites != nil && m.ListDetail == "playlist" {
				for _, fav := range m.FavoritePlaylists {
					if fav.URI == m.LibraryList[m.Cursor].URI {
						util.RemoveFromJSON(file, fav)
						m.FavoritePlaylists, _ = util.ReadJSON(file)
						return m, spotify.HandleFetchLibrary(m.FavoritePlaylists, m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(m.FavoritePlaylists), m.Offset)
					}
				}
				util.WriteJSONFile(file, types.LibraryFavorite{m.LibraryList[m.Cursor].Name, m.LibraryList[m.Cursor].Artist, m.LibraryList[m.Cursor].URI})
				m.FavoritePlaylists, _ = util.ReadJSON(file)
			}
			favorites, _ = util.ReadJSON(file)
			return m, spotify.HandleFetchLibrary(favorites, m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(favorites), m.Offset)

		case util.Keybinds["Cursor Up"]:
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				m.Cursor = len(m.LibraryList) - 1
			}
			return m, nil

		case util.Keybinds["Cursor Down"]:
			if m.Cursor < len(m.LibraryList)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}
			return m, nil

		case util.Keybinds["Next Page"]:
			m.Loading = true
			favorites := util.GetFavorites(m.Model)
			if m.Offset+m.Height-util.LIBRARY_SPACING-len(favorites) < m.ApiTotal {
				if m.Offset == 0 {
					m.Offset += m.Height - (util.LIBRARY_SPACING + len(favorites))
				} else {
					m.Offset += m.Height - (util.LIBRARY_SPACING + len(favorites)) - 3
				}
			} else {
				m.Offset = 0
			}
			return m, spotify.HandleFetchLibrary(favorites, m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(favorites), m.Offset)

		case util.Keybinds["Previous Page"]:
			m.Loading = true
			favorites := util.GetFavorites(m.Model)
			page := m.Offset/(m.Height-(util.LIBRARY_SPACING+len(favorites))) + 1
			if page > 1 {
				if m.Offset == m.Height-(util.LIBRARY_SPACING+len(favorites)) {
					m.Offset -= m.Height - (util.LIBRARY_SPACING + len(favorites))
				} else {
					m.Offset -= m.Height - (util.LIBRARY_SPACING + len(favorites)) - 3
				}
			} else {
				m.Offset = m.ApiTotal - (m.ApiTotal % (m.Height - (util.LIBRARY_SPACING + len(favorites))))
			}
			return m, spotify.HandleFetchLibrary(favorites, m.Token, m.ListDetail, m.Height-util.LIBRARY_SPACING-len(favorites), m.Offset)

		case util.Keybinds["Select"]:
			if m.State.IsPlaying {
				if m.LibraryList != nil {
					if m.ListDetail == "album" {
						spotify.HandleGenericPut("/me/player/shuffle", m.Token, map[string]string{"state": "false"}, nil)
					} else {
						spotify.HandleGenericPut("/me/player/shuffle", m.Token, map[string]string{"state": "true"}, nil)
					}
					spotify.HandleGenericPut("/me/player/play", m.Token, map[string]string{"device_id": m.State.Device.ID}, map[string]string{"context_uri": m.LibraryList[m.Cursor].URI})
					return m, nil
				}
			}
			return m, nil
		case util.Keybinds["First Tab"]:
			if m.ListDetail == "album" {
				return m, nil
			}
			m.ListDetail = "album"
			m.Loading = true
			m.Offset = 0
			favorites := util.GetFavorites(m.Model)
			return m, spotify.HandleFetchLibrary(favorites, m.Token, "album", m.Height-util.LIBRARY_SPACING-len(favorites), 0)
		case util.Keybinds["Second Tab"]:
			if m.ListDetail == "playlist" {
				return m, nil
			}
			m.ListDetail = "playlist"
			m.Loading = true
			m.Offset = 0
			favorites := util.GetFavorites(m.Model)
			return m, spotify.HandleFetchLibrary(favorites, m.Token, "playlist", m.Height-util.LIBRARY_SPACING-len(favorites), 0)
		}

	case types.PlaybackState:
		if len(msg.Item.Album.Images) > 0 {
			if m.State.Item.Name != msg.Item.Name {
				width, height, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					log.Fatalf("Failed to get terminal size: %v", err)
				}
				m.Image = util.MakeNewImage(msg.Item.Album.Images[0].URL, width/4-3, height/2)
				m.State = msg
				return m, tea.Batch(util.ScheduleNextFetch(FETCH_TIMER*time.Second), spotify.CheckTokenExpiryCmd(m.Model), spotify.HandleGetQueue(m.Token))
			}
		}
		m.State = msg
		if math.Abs(float64(m.ProgressMs-msg.ProgressMs)) > 1000 { // Don't bother unless we are more then a second off
			m.ProgressMs = msg.ProgressMs
		}
		return m, tea.Batch(util.ScheduleNextFetch(FETCH_TIMER*time.Second), spotify.CheckTokenExpiryCmd(m.Model))

	case types.SpotifyTokenResponse:
		m.Token = msg.AccessToken
		m.TokenExpiresAt = time.Now().Add(time.Duration(msg.ExpiresIn) * time.Second)
		return m, nil

	case types.SpotifyAlbum:
		m.LibraryList = nil
		favorites := util.GetFavorites(m.Model)
		for _, album := range favorites {
			m.LibraryList = append(m.LibraryList, types.LibraryItem{Name: album.Title, Artist: album.Author, URI: album.URI, Favorite: true})
		}
		for _, album := range msg.Items {
			m.LibraryList = append(m.LibraryList, types.LibraryItem{Name: album.Album.Name, Artist: album.Album.Artists[0].Name, URI: album.Album.URI, Favorite: false})
			m.ApiTotal = msg.Total - len(favorites)
			m.Loading = false
		}

	case types.SpotifyPlaylist:
		m.LibraryList = nil
		favorites := util.GetFavorites(m.Model)
		for _, playlist := range favorites {
			m.LibraryList = append(m.LibraryList, types.LibraryItem{Name: playlist.Title, Artist: playlist.Author, URI: playlist.URI, Favorite: true})
		}
		for _, playlist := range msg.Items {
			m.LibraryList = append(m.LibraryList, types.LibraryItem{Name: playlist.Name, Artist: playlist.Owner.DisplayName, URI: playlist.URI, Favorite: false})
			m.ApiTotal = msg.Total
			m.Loading = false
		}

	case types.Queue:
		// Store the full queue - will be truncated in View() with real calculated height
		m.Queue = msg
		return m, nil

	case error:
		m.ErrMsg = msg.Error()
		m.Loading = false
		return m, util.ScheduleNextFetch(FETCH_TIMER * time.Second)

	case types.PlaybackMsg:
		return m, spotify.HandleFetchPlayback(m.Token)

	case types.ProgressMsg:
		if m.State.IsPlaying {
			m.ProgressMs += 1000
			if m.State.Item.DurationMs-m.ProgressMs < 2000 {
				return m, tea.Batch(util.ScheduleProgressInc(1 * time.Second))
			}
		}

		return m, util.ScheduleProgressInc(1 * time.Second)
	}
	return m, nil
}

func (m Model) View() string {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}
	boxWidth := width/2 - 2
	boxHeight := height - util.UI_LIBRARY_SPACE // correct constant for layout (not fetch limit)
	playBackWidth := width - 2

	jukeboxHeight := height / 2
	visQueueHeight := boxHeight - jukeboxHeight - 6 // -4 for two box overheads (border+padding), -2 gap
	if visQueueHeight < 3 {
		visQueueHeight = 3
	}

	libText, playback, image, visQueue := util.GetUiElements(m.Model, boxWidth, visQueueHeight-2) // header + overhead

	library := util.LibraryStyle.Width(boxWidth).Height(boxHeight).Render(libText)
	jukebox := util.BoxStyle.Width(boxWidth).Height(jukeboxHeight).Render(image)
	playbackBar := util.BoxStyle.Width(playBackWidth).Height(1).Render(playback)
	visualQueue := util.BoxStyle.Width(boxWidth).Height(visQueueHeight).Render(visQueue)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, library, lipgloss.JoinVertical(lipgloss.Left, jukebox, visualQueue)),
		playbackBar,
	)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")
	listDetail := "album"

	util.SetKeybinds()

	util.CheckArguments()

	fmt.Println("Opening login page...")
	spotify.OpenLoginPage(clientID)
	code := spotify.GetCodeFromCallback()
	token, err := spotify.GetSpotifyToken(context.Background(), clientID, clientSecret, code)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Login successful! Access token retrieved.\n" + fmt.Sprintf("Press '%s' to Play/Pause, '%s' to Skip, '%s' to Quit", util.Keybinds["Play/Pause"], util.Keybinds["Skip"], util.Keybinds["Quit"]))

	favoriteAlbums, success := util.ReadJSON(fmt.Sprintf("data/favorites/albums.json"))
	if !success {
		fmt.Println("No favorites found. Creating new favorites file.")
		util.CreateEmptyJSONFile(fmt.Sprintf("data/favorites/albums.json"))
	}

	favoritePlaylists, success := util.ReadJSON(fmt.Sprintf("data/favorites/playlists.json"))
	if !success {
		fmt.Println("No favorites found. Creating new favorites file.")
		util.CreateEmptyJSONFile(fmt.Sprintf("data/favorites/playlists.json"))
	}

	model := initialModel(token.AccessToken, listDetail, favoriteAlbums, favoritePlaylists)
	model.RefreshToken = token.RefreshToken
	model.TokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
