package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ===========================================
// ===== util.go | General Program Utils =====
// ===========================================

// Schedule the next fetch of the playback state.
func scheduleNextFetch(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return playbackMsg{}
	}
}

// Schedule the next increment of the progress bar.
func scheduleProgressInc(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return progressMsg{}
	}
}

// Check if the user has passed in any arguments
func checkArguments() {
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if arg == "-h" || arg == "--help" {
				order := []string{
					"Skip",
					"Play/Pause",
					"Select",
					"Shuffle",
					"Favorites",
					"Next Page",
					"Previous Page",
					"Cursor Up",
					"Cursor Down",
					"Quit",
				}
				fmt.Println("Keybinds:")
				for _, key := range order {
					fmt.Printf("\t%s: %s\n", key, keybinds[key])
				}
				os.Exit(0)
			}
			if arg == "-v" || arg == "--version" {
				fmt.Println("JukeTUI v1.0.0")
				os.Exit(0)
			}
		}
	}
}

// Query an environment variable, returning a default value if it is not set
func queryEnv(envKey, defaultValue string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}

// Set the keybinds for the application
func setKeybinds() {
	keybinds = map[string]string{
		"Quit":          queryEnv("QUIT", "q"),
		"Play/Pause":    queryEnv("PLAYPAUSE", "p"),
		"Skip":          queryEnv("SKIP", "n"),
		"Shuffle":       queryEnv("SHUFFLE", "s"),
		"Favorites":     queryEnv("FAVORITES", "f"),
		"Cursor Up":     "up",
		"Cursor Down":   "down",
		"Next Page":     "right",
		"Previous Page": "left",
		"Select":        "enter",
	}
}
