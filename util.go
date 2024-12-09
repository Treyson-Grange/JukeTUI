package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Util.go contains utility functions that are used in the main.go file
// Kinda a catch-all for functions that don't fit anywhere else

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
				fmt.Println("Keybinds:")
				keys := make([]string, 0, len(keybinds))
				for key := range keybinds {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					fmt.Printf("\t%s: %s\n", key, keybinds[key])
				}
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
