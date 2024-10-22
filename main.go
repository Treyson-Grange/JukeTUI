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
}

type tickMsg struct{}

func initialModel(token string) Model {
	return Model{
		token:   token,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchPlaybackStateCmd(m.token)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "p":
			if m.state.IsPlaying {
				handleGenericPut("/me/player/pause", m.token, nil)
			} else {
				handleGenericPut("/me/player/play", m.token, nil)
			}
			return m, fetchPlaybackStateCmd(m.token)
		case "n":
			handleGenericPost("/me/player/next", m.token, nil)
			return m, fetchPlaybackStateCmd(m.token)
		case "r":
			data := handleGenericFetch[SpotifyRecommendations]("/recommendations", m.token, map[string]string{"seed_tracks": m.state.Item.ID, "limit": "1"})
			m.reccomendation = data

			return m, fetchPlaybackStateCmd(m.token)
		case "c": // play reccomendation
			fmt.Println(m.reccomendation.Tracks[0].URI)
			handleGenericPut("/me/player/play", m.token, nil)
			// we need to put a list of uri's in the BODY.
			// Or what we can do, is add to the queue, and then skip. 
			// So we keep whatever album/playlist is playing, and then skip to the reccomendation.
			// Then after teh reccomenation is done, we skip back to the original album/playlist.
			// Issues: 
				// what if the user already has a queue? It goes to end. Issue
			// K heres my decision for now. We will add to queue but not skip.
			// we will tell them it was added to the queue. 
			return m, fetchPlaybackStateCmd(m.token)
			
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
	if len(m.reccomendation.Tracks) > 0 {
		return fmt.Sprintf(
			"Reccomendations: %s %s %s ", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name, m.reccomendation.Tracks[0].URI,
		)
	}
	return fmt.Sprintf(
		"Track: %s\nStatus: %s\n\nPress 'p' to Play/Pause, 'n' to Skip, 'q' to Quit.",
		m.state.Item.Name, status,
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
		state := handleGenericFetch[PlaybackState]("/me/player", token, nil)
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

	fmt.Println("Opening login page...")
	OpenLoginPage(clientID)
	code := GetCodeFromCallback()
	token, err := GetSpotifyToken(context.Background(), clientID, clientSecret, code)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Login successful! Access token retrieved.")

	p := tea.NewProgram(initialModel(token.AccessToken))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
	test()
}
