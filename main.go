package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

func initialModel(token, listDetail string) Model {
	return Model{
		token:      token,
		loading:    true,
		listDetail: listDetail,
		height:     10,
		progressMs: 0,
	}
}

var errorLogger = func() *log.Logger {
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}()

func (m Model) Init() tea.Cmd {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}
	return tea.Batch(
		handleFetchPlayback(m.token),
		scheduleProgressInc(1*time.Second),
		handleFetchLibrary(m.token, m.listDetail, height-11),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q":
			return m, tea.Quit

		case "p", "P":
			if m.state.IsPlaying {
				handleGenericPut("/me/player/pause", m.token, nil, nil)
			} else {
				handleGenericPut("/me/player/play", m.token, nil, map[string]string{"device_id": m.state.Device.ID})
			}
			return m, handleFetchPlayback(m.token)

		case "n", "N":
			handleGenericPost("/me/player/next", m.token, nil, nil)
			return m, handleFetchPlayback(m.token)

		case "r", "R":
			data := handleGenericFetch[SpotifyRecommendations]("/recommendations", m.token, map[string]string{"seed_tracks": m.state.Item.ID, "limit": "1"}, nil)
			m.reccomendation = data
			return m, handleFetchPlayback(m.token)

		case "c", "C":
			if len(m.reccomendation.Tracks) > 0 {
				handleGenericPost("/me/player/queue", m.token, map[string]string{"uri": m.reccomendation.Tracks[0].URI}, nil)
			}
			return m, handleFetchPlayback(m.token)

		case "up":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.libraryList) - 1
			}

		case "down":
			if m.cursor < len(m.libraryList)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case "enter":
			if m.state.IsPlaying {
				if m.libraryList != nil {
					if m.listDetail == "album" {
						handleGenericPut("/me/player/shuffle", m.token, map[string]string{"state": "false"}, nil)
					} else {
						handleGenericPut("/me/player/shuffle", m.token, map[string]string{"state": "true"}, nil)
					}
					handleGenericPut("/me/player/play", m.token, map[string]string{"device_id": m.state.Device.ID}, map[string]string{"context_uri": m.libraryList[m.cursor].uri})
					return m, handleFetchPlayback(m.token)
				}
			}
		}

	case PlaybackState:
		m.state = msg
		m.loading = false
		if math.Abs(float64(m.progressMs-msg.ProgressMs)) > 1000 { // Don't bother unless we are more then a second off
			m.progressMs = msg.ProgressMs
		}
		return m, tea.Batch(scheduleNextFetch(3*time.Second), CheckTokenExpiryCmd(m))

	case SpotifyTokenResponse:
		m.token = msg.AccessToken
		m.tokenExpiresAt = time.Now().Add(time.Duration(msg.ExpiresIn) * time.Second)
		return m, nil

	case SpotifyAlbum:
		m.libraryList = nil
		for _, album := range msg.Items {
			m.libraryList = append(m.libraryList, LibraryItem{name: album.Album.Name, artist: album.Album.Artists[0].Name, uri: album.Album.URI})
		}

	case SpotifyPlaylist:
		m.libraryList = nil
		for _, playlist := range msg.Items {
			m.libraryList = append(m.libraryList, LibraryItem{name: playlist.Name, artist: playlist.Owner.DisplayName, uri: playlist.URI})
		}

	case error:
		m.errMsg = msg.Error()
		m.loading = false
		return m, scheduleNextFetch(3 * time.Second)

	case playbackMsg:
		return m, handleFetchPlayback(m.token)

	case progressMsg:
		m.progressMs += 1000
		return m, scheduleProgressInc(1 * time.Second)
	}

	return m, nil
}

func (m Model) View() string {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}
	boxWidth := width/2 - 2
	boxHeight := height - 10
	playBackHeight := 1
	playBackWidth := width - 2

	if m.loading {
		return "Loading playback state...\n"
	}
	if m.errMsg != "" {
		return fmt.Sprintf("Error: %s\n", m.errMsg)
	}

	status := "▶ "
	if m.state.IsPlaying {
		status = "▮▮"
	}

	recommendationDetails := GetAsciiJuke(boxWidth, boxHeight, false)

	if len(m.reccomendation.Tracks) > 0 {
		recommendationDetails += fmt.Sprintf(
			"Recommendation: %s - %s\n 'c' to add to your queue!", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name,
		)
	}

	libText := ""
	if m.libraryList != nil {
		const CHARACTERS = 6 // Characters we have to account for when truncating
		for i, item := range m.libraryList {
			if i == m.cursor {
				item = LibraryItem{
					name:   lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render("> " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS)),
					artist: item.artist,
					uri:   item.uri,
				}
			} else {
				item = LibraryItem{
					name:   "  " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS),
					artist: item.artist,
					uri:   item.uri,
				}
			} 
			play := map[bool]string{true: " 🔊", false: ""}[m.state.Context.URI == item.uri]
			libText += fmt.Sprintf("%s - %s%s\n", item.name, item.artist, play)
		}
	}

	var playback string
	if m.state.Item.Artists != nil {
		playback = "🎵 [ " + m.state.Item.Name + " | " + m.state.Item.Artists[0].Name + " ]  " + lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render(status) + "  [ " + msToMinSec(m.progressMs) + " / " + msToMinSec(m.state.Item.DurationMs) + " ]"
	} else {
		playback = "No Playback Data. Please start a playback session on your device"
	}

	library := libraryStyle.Width(boxWidth).Height(boxHeight).Render(libText)
	jukebox := boxStyle.Width(boxWidth).Height(boxHeight).Render(recommendationDetails)
	playbackBar := boxStyle.Width(playBackWidth).Height(playBackHeight).Render(playback)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, library, jukebox),
		playbackBar,
	)
}

// Schedule the next fetch of the playback state.
func scheduleNextFetch(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return playbackMsg{}
	}
}

func scheduleProgressInc(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return progressMsg{}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")
	listDetail := os.Getenv("SPOTIFY_PREFERENCE")

	fmt.Println("Opening login page...")
	OpenLoginPage(clientID)
	code := GetCodeFromCallback()
	token, err := GetSpotifyToken(context.Background(), clientID, clientSecret, code)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Login successful! Access token retrieved.")
	fmt.Println("Press 'p' to Play/Pause, 'n' to Skip, 'q' to Quit")

	model := initialModel(token.AccessToken, listDetail)
	model.refreshToken = token.RefreshToken
	model.tokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
