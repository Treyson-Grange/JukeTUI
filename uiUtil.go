package main

import (
	"fmt"

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
//
// Parameters:
// - str: the string to truncate
// - width: the width to truncate the string to
//
// Returns:
// - string: the truncated string
func truncate(str string, width int) string {
	if len(str) > width {
		return str[:width-3] + "..."
	}
	return str
}

// Turn ms to 5:30 format
func msToMinSec(ms int) string {
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

const SPOTIFY_GREEN = "#1DB954"
