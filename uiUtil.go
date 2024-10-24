package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1).
			Align(lipgloss.Center)

	libraryStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0).Align(lipgloss.Left)
)

// Truncate a string to fit any width
func truncate(str string, width int) string {
	if len(str) > width {
		return str[:width-3] + "..."
	}
	return str
}

const SPOTIFY_GREEN = "#1DB954"
