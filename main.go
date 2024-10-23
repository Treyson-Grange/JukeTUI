package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

type Model struct {
	/*
	 * Playback state, including track info, playback status, etc.
	 * For specifics, see PlaybackState struct in models.go
	 */
	state          PlaybackState
	/*
	 * Spotify web API access token. Lasts for 1 hour.
	 */
	token          string
	/*
	 * Spotify web API refresh token. Used to get a new access token when the current one is close to expiration.
	 */
	refreshToken   string
	/*
	 * Time when the current access token expires.
	 */
	tokenExpiresAt time.Time
	/*
	 * Error message, if any
	 */
	errMsg         string
	/*
	 * Whether or not we're currently fetching access token initially
	 */
	loading        bool
	/*
	 * Current recommendation, if any.
	 */
	reccomendation SpotifyRecommendations
	/*
	 * List detail, either "album" or "playlist".
	 */
	listDetail     string
}

var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Align(lipgloss.Center)

	horizontalGap = lipgloss.NewStyle().Padding(0, 1)
)

type tickMsg struct{}

func initialModel(token, listDetail string) Model {
	return Model{
		token:      token,
		loading:    true,
		listDetail: listDetail,
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
	return fetchPlaybackStateCmd(m.token)
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
			return m, fetchPlaybackStateCmd(m.token)
		case "n", "N":
			handleGenericPost("/me/player/next", m.token, nil, nil)
			return m, fetchPlaybackStateCmd(m.token)
		case "r", "R":
			data := handleGenericFetch[SpotifyRecommendations]("/recommendations", m.token, map[string]string{"seed_tracks": m.state.Item.ID, "limit": "1"}, nil)
			m.reccomendation = data
			return m, fetchPlaybackStateCmd(m.token)
		case "c", "C":
			if len(m.reccomendation.Tracks) > 0 {
				handleGenericPost("/me/player/queue", m.token, map[string]string{"uri": m.reccomendation.Tracks[0].URI}, nil)
			}
			return m, fetchPlaybackStateCmd(m.token)
		case "t", "T":
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
		return m, tea.Batch(scheduleNextFetch(3*time.Second), CheckTokenExpiryCmd(m))

	case SpotifyTokenResponse:
		m.token = msg.AccessToken
		m.tokenExpiresAt = time.Now().Add(time.Duration(msg.ExpiresIn) * time.Second)
		return m, nil

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
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}
	boxWidth := width / 2 - 2
	boxHeight := height - 10
	playBackHeight := 1
	playBackWidth := width - 2

	if m.loading {
		return "Loading playback state...\n"
	}
	if m.errMsg != "" {
		return fmt.Sprintf("Error: %s\n", m.errMsg)
	}

	status := "▶"
	if m.state.IsPlaying {
		status = "◼"
	}

	var recommendationDetails string
	if len(m.reccomendation.Tracks) > 0 {
		recommendationDetails = fmt.Sprintf(
			"Recommendations: %s - %s\n", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name,
		)
	}

	library := boxStyle.Width(boxWidth).Height(boxHeight).Render(fmt.Sprintf(""))
	jukebox := boxStyle.Width(boxWidth).Height(boxHeight).Render(recommendationDetails)
	playbackBar := boxStyle.Width(playBackWidth).Height(playBackHeight).Render("Now playing: " + m.state.Item.Name + " - " + m.state.Item.Artists[0].Name + " (" + status + " )")

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
		return tickMsg{}
	}
}

// Fetch the playback state.
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
	fmt.Println("Press 'p' to Play/Pause, 'n' to Skip, 'q' to Quit, 'r' to get recommendations, 'c' to add recommendation to queue")

	model := initialModel(token.AccessToken, listDetail)
	model.refreshToken = token.RefreshToken
	model.tokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}