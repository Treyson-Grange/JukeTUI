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
	"golang.org/x/term"
)

func initialModel(token, listDetail string, favorites []LibraryFavorite) Model {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}

	return Model{
		token:      token,
		listDetail: listDetail,
		height:     height,
		favorites:  favorites,
	}
}

var keybinds = map[string]string{}

// spotify api is 180 per minute, counts over a 30 second rolling window. 1 fetch per second will be safe
const FETCH_TIMER = 1

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		handleFetchPlayback(m.token),
		handleGetLibraryTotal(m.token, m.listDetail),
		scheduleProgressInc(1*time.Second),
		handleFetchLibrary(m.favorites, m.token, m.listDetail, m.height-LIBRARY_SPACING-len(m.favorites), 0),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case keybinds["Quit"]:
			return m, tea.Quit

		case keybinds["Play/Pause"]:
			if m.state.IsPlaying {
				handleGenericPut("/me/player/pause", m.token, nil, nil)
			} else {
				handleGenericPut("/me/player/play", m.token, nil, map[string]string{"device_id": m.state.Device.ID})
			}
			return m, nil

		case keybinds["Skip"]:
			handleGenericPost("/me/player/next", m.token, nil, nil)
			return m, nil

		case keybinds["Shuffle"]:
			handleGenericPut("/me/player/shuffle", m.token, map[string]string{"state": fmt.Sprintf("%t", !m.state.ShuffleState)}, nil)
			return m, nil

		case keybinds["Recommendation"]:
			if m.state.Item.ID != "" {
				data := handleGenericFetch[SpotifyRecommendations]("/recommendations", m.token, map[string]string{"seed_tracks": m.state.Item.ID, "limit": "1"}, nil)
				m.reccomendation = data
				m.image = makeNewImage(m.reccomendation.Tracks[0].Album.Image[0].URL)
			}
			return m, nil

		case keybinds["Add to Queue"]:
			if len(m.reccomendation.Tracks) > 0 {
				handleGenericPost("/me/player/queue", m.token, map[string]string{"uri": m.reccomendation.Tracks[0].URI}, nil)
			}
			m.image = makeNewImage(m.state.Item.Album.Images[0].URL) // Reset image to current song.
			return m, nil

		case keybinds["Favorites"]:
			file := fmt.Sprintf("favorites/%ss.json", m.listDetail)
			if m.favorites != nil {
				for _, fav := range m.favorites {
					if fav.URI == m.libraryList[m.cursor].uri {
						removeFromJSON(file, fav)
						m.favorites, _ = readJSON(file)
						return m, handleFetchLibrary(m.favorites, m.token, m.listDetail, m.height-LIBRARY_SPACING-len(m.favorites), m.offset)
					}
				}
				writeJSONFile(file, LibraryFavorite{m.libraryList[m.cursor].name, m.libraryList[m.cursor].artist, m.libraryList[m.cursor].uri})
			}
			m.favorites, _ = readJSON(file)
			return m, handleFetchLibrary(m.favorites, m.token, m.listDetail, m.height-LIBRARY_SPACING-len(m.favorites), m.offset)

		case keybinds["Cursor Up"]:
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.libraryList) - 1
			}

		case keybinds["Cursor Down"]:
			if m.cursor < len(m.libraryList)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case keybinds["Next Page"]:
			m.loading = true
			if m.offset+m.height-LIBRARY_SPACING-len(m.favorites) < m.apiTotal {
				if m.offset == 0 { // IDK why this is necessary, but its the only way i got it working.
					m.offset += m.height - (UI_LIBRARY_SPACE + len(m.favorites))
				} else {
					m.offset += m.height - (UI_LIBRARY_SPACE + len(m.favorites)) - 3
				}
			} else {
				m.offset = 0
			}
			return m, handleFetchLibrary(m.favorites, m.token, m.listDetail, m.height-LIBRARY_SPACING-len(m.favorites), m.offset)

		case keybinds["Previous Page"]:
			m.loading = true
			page := m.offset/(m.height-(UI_LIBRARY_SPACE+len(m.favorites))) + 1
			if page > 1 {
				if m.offset == m.height-(UI_LIBRARY_SPACE+len(m.favorites)) {
					m.offset -= m.height - (UI_LIBRARY_SPACE + len(m.favorites))
				} else {
					m.offset -= m.height - (UI_LIBRARY_SPACE + len(m.favorites)) - 3
				}
			} else {
				m.offset = m.apiTotal - (m.apiTotal % (m.height - (UI_LIBRARY_SPACE + len(m.favorites))))
			}
			return m, handleFetchLibrary(m.favorites, m.token, m.listDetail, m.height-LIBRARY_SPACING-len(m.favorites), m.offset)

		case keybinds["Select"]:
			if m.state.IsPlaying {
				if m.libraryList != nil {
					if m.listDetail == "album" {
						handleGenericPut("/me/player/shuffle", m.token, map[string]string{"state": "false"}, nil)
					} else {
						handleGenericPut("/me/player/shuffle", m.token, map[string]string{"state": "true"}, nil)
					}
					handleGenericPut("/me/player/play", m.token, map[string]string{"device_id": m.state.Device.ID}, map[string]string{"context_uri": m.libraryList[m.cursor].uri})
					return m, nil
				}
			}
		}

	case PlaybackState:
		if len(msg.Item.Album.Images) > 0 {
			if m.state.Item.Name != msg.Item.Name {
				m.image = makeNewImage(msg.Item.Album.Images[0].URL)
			}
		}
		m.state = msg
		if math.Abs(float64(m.progressMs-msg.ProgressMs)) > 1000 { // Don't bother unless we are more then a second off
			m.progressMs = msg.ProgressMs
		}
		return m, tea.Batch(scheduleNextFetch(FETCH_TIMER*time.Second), CheckTokenExpiryCmd(m))

	case SpotifyTokenResponse:
		m.token = msg.AccessToken
		m.tokenExpiresAt = time.Now().Add(time.Duration(msg.ExpiresIn) * time.Second)
		return m, nil

	case SpotifyAlbum:
		m.libraryList = nil
		for _, album := range m.favorites {
			m.libraryList = append(m.libraryList, LibraryItem{name: album.Title, artist: album.Author, uri: album.URI, favorite: true})
		}
		for _, album := range msg.Items {
			m.libraryList = append(m.libraryList, LibraryItem{name: album.Album.Name, artist: album.Album.Artists[0].Name, uri: album.Album.URI, favorite: false})
			m.apiTotal = msg.Total - len(m.favorites)
			m.loading = false
		}

	case SpotifyPlaylist:
		m.libraryList = nil
		for _, playlist := range m.favorites {
			m.libraryList = append(m.libraryList, LibraryItem{name: playlist.Title, artist: playlist.Author, uri: playlist.URI, favorite: true})
		}
		for _, playlist := range msg.Items {
			m.libraryList = append(m.libraryList, LibraryItem{name: playlist.Name, artist: playlist.Owner.DisplayName, uri: playlist.URI, favorite: false})
			m.apiTotal = msg.Total
			m.loading = false
		}

	case error:
		m.errMsg = msg.Error()
		m.loading = false
		return m, scheduleNextFetch(FETCH_TIMER * time.Second)

	case playbackMsg:
		return m, handleFetchPlayback(m.token)

	case progressMsg:
		if m.state.IsPlaying {
			m.progressMs += 1000
			if m.state.Item.DurationMs-m.progressMs < 2000 {
				return m, tea.Batch(scheduleProgressInc(1 * time.Second))
			}
		}

		return m, scheduleProgressInc(1 * time.Second)
	}

	return m, nil
}

func (m Model) View() string {
	//TODO: this is calling GetSize every view. Store it in our model.
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}
	boxWidth := width/2 - 2
	boxHeight := height - UI_LIBRARY_SPACE
	playBackWidth := width - 2

	libText, playback, reccDetails := getUiElements(m, boxWidth)

	library := libraryStyle.Width(boxWidth).Height(boxHeight).Render(libText)
	jukebox := boxStyle.Width(boxWidth).Height(boxHeight).Render(reccDetails)
	playbackBar := boxStyle.Width(playBackWidth).Height(1).Render(playback)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, library, jukebox),
		playbackBar,
	)
}

func main() {
	checkArguments()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")
	listDetail := os.Getenv("SPOTIFY_PREFERENCE")

	setKeybinds()

	fmt.Println("Opening login page...")
	OpenLoginPage(clientID)
	code := GetCodeFromCallback()
	token, err := GetSpotifyToken(context.Background(), clientID, clientSecret, code)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Login successful! Access token retrieved.\n" + fmt.Sprintf("Press '%s' to Play/Pause, '%s' to Skip, '%s' to Quit", keybinds["Play/Pause"], keybinds["Skip"], keybinds["Quit"]))

	favorites, success := readJSON(fmt.Sprintf("favorites/%ss.json", listDetail))
	if !success {
		fmt.Println("No favorites found. Creating new favorites file.")
		createEmptyJSONFile(fmt.Sprintf("favorites/%ss.json", listDetail))
	}

	model := initialModel(token.AccessToken, listDetail, favorites)
	model.refreshToken = token.RefreshToken
	model.tokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
