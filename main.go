package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

// Main.go
// This is the main file. It is responsible for setting up the program and running it.
// Uses bubbletea to create a simple terminal UI

type Model struct {
	state          PlaybackState
	token          string
	errMsg         string
	loading        bool
	reccomendation SpotifyRecommendations
	listDetail     string
}

type tickMsg struct{}

func initialModel(token, listDetail string) Model {
	return Model{
		token:      token,
		loading:    true,
		listDetail: listDetail,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchPlaybackStateCmd(m.token)
}

var errorLogger = func() *log.Logger { //IDK where to put this
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}()

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "p":
			if m.state.IsPlaying {
				handleGenericPut("/me/player/pause", m.token, nil, nil)
			} else {
				handleGenericPut("/me/player/play", m.token, nil, nil)
			}
			return m, fetchPlaybackStateCmd(m.token)
		case "n":
			handleGenericPost("/me/player/next", m.token, nil, nil)
			return m, fetchPlaybackStateCmd(m.token)
		case "r":
			data := handleGenericFetch[SpotifyRecommendations]("/recommendations", m.token, map[string]string{"seed_tracks": m.state.Item.ID, "limit": "1"}, nil)
			m.reccomendation = data

			return m, fetchPlaybackStateCmd(m.token)
		case "c": // add to queuer
			handleGenericPost("/me/player/queue", m.token, map[string]string{"uri": m.reccomendation.Tracks[0].URI}, nil)
			return m, fetchPlaybackStateCmd(m.token)

		case "t": // General Test command

			if m.listDetail == "album" {
				test := handleGenericFetch[SpotifyAlbum]("/me/albums", m.token, map[string]string{"limit": "20"}, nil)
				for _, album := range test.Items {
					fmt.Println(album.Album.Name)
				}
			} else {
				test := handleGenericFetch[SpotifyPlaylist]("/me/playlists", m.token, map[string]string{"limit": "20"}, nil)
				for _, playlist := range test.Items {
					fmt.Println(playlist.Name)
				}
			}

		}

	case PlaybackState:
		m.state = msg
		m.loading = false
		return m, scheduleNextFetch(3 * time.Second)

	case error:
		m.errMsg = msg.Error()
		m.loading = false
		return m, scheduleNextFetch(3 * time.Second)

	case tickMsg:
		return m, fetchPlaybackStateCmd(m.token)
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return "Loading playback state...\n"
	}
	if m.errMsg != "" {
		return fmt.Sprintf("Error: %s\n", m.errMsg)
	}
	status := "Paused"
	if m.state.IsPlaying {
		status = "Playing"
	}

	var recommendationDetails string
	if len(m.reccomendation.Tracks) > 0 {
		recommendationDetails = fmt.Sprintf(
			"Recommendations: %s - %s\n", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name,
		)
	}

	return fmt.Sprintf(
		"Track: %s\nStatus: %s\n%s\nPress 'p' to Play/Pause, 'n' to Skip, 'q' to Quit, 'r' to get recommendations, 'c' to add recc to queue\n",
		m.state.Item.Name, status, recommendationDetails,
	)
}

func scheduleNextFetch(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return tickMsg{}
	}
}

func fetchPlaybackStateCmd(token string) tea.Cmd {
	return func() tea.Msg {
		state := handleGenericFetch[PlaybackState]("/me/player", token, nil, nil)
		return state
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

	p := tea.NewProgram(initialModel(token.AccessToken, listDetail))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
