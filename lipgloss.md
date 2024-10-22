package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styling using Lip Gloss
var (
	topBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1).
			Margin(0)

	playbackStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			Padding(1).
			Margin(1)
)

type model struct {
	width, height int
	topSection1   string
	topSection2   string
	playback      string
}

// Init initializes the model with the first tick.
func (m model) Init() tea.Cmd {
	return tick()
}

// Update handles incoming messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Capture the current terminal size
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit // Exit on Ctrl+C
		}

	case tickMsg:
		// Update the sections with dynamic content
		m.topSection1 = fmt.Sprintf("Random: %d", rand.Intn(100))
		m.topSection2 = fmt.Sprintf("Time: %s", time.Now().Format("15:04:05"))
		m.playback = "Now Playing: 🎵 Music Track..."

		return m, tick() // Schedule the next tick
	}

	return m, nil
}

// View renders the layout.
func (m model) View() string {
	// Calculate widths and heights dynamically
	halfWidth := m.width / 2
	halfHeight := (m.height - 6) / 2 // Adjust for borders and padding

	// Apply calculated dimensions
	topSection1 := topBoxStyle.Copy().Width(halfWidth).Height(halfHeight).Render(m.topSection1)
	topSection2 := topBoxStyle.Copy().Width(halfWidth).Height(halfHeight).Render(m.topSection2)

	// Render top row side by side
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topSection1, topSection2)

	// Playback section fills the bottom part
	playback := playbackStyle.Copy().Width(m.width).Height(halfHeight).Render(m.playback)

	// Join the two main sections vertically
	return lipgloss.JoinVertical(lipgloss.Top, topRow, playback) + "\n\nPress Ctrl+C to exit."
}

// Custom message type for the tick.
type tickMsg time.Time

// tick schedules a new tick every second.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Ensure the terminal is properly restored on exit
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

