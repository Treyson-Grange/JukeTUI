package util

import (
	"fmt"
	"os"
	"time"

	"github.com/treyson-grange/JukeTUI/src/types"

	tea "github.com/charmbracelet/bubbletea"
)

// ===========================================
// ===== util.go | General Program Utils =====
// ===========================================

// Keybinds stores the application keybinds
var Keybinds map[string]string

// Schedule the next fetch of the playback state.
func ScheduleNextFetch(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return types.PlaybackMsg{}
	}
}

// Schedule the next increment of the progress bar.
func ScheduleProgressInc(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return types.ProgressMsg{}
	}
}

// Check if the user has passed in any arguments
func CheckArguments() {
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
					"First Tab",
					"Second Tab",
				}
				fmt.Println("Keybinds:")
				for _, key := range order {
					fmt.Printf("\t%s: %s\n", key, Keybinds[key])
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
func QueryEnv(envKey, defaultValue string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}

// Set the keybinds for the application
func SetKeybinds() {
	Keybinds = map[string]string{
		"Quit":          QueryEnv("QUIT", "q"),
		"Play/Pause":    QueryEnv("PLAYPAUSE", "p"),
		"Skip":          QueryEnv("SKIP", "n"),
		"Shuffle":       QueryEnv("SHUFFLE", "s"),
		"Favorites":     QueryEnv("FAVORITES", "f"),
		"Cursor Up":     "up",
		"Cursor Down":   "down",
		"Next Page":     "right",
		"Previous Page": "left",
		"Select":        "enter",
		"First Tab":     "1",
		"Second Tab":    "2",
	}
}

// Given a model, return the list of favorites corresponding to the current list detail
func GetFavorites(m *types.Model) []types.LibraryFavorite {
	if m.ListDetail == "album" {
		return m.FavoriteAlbums
	} else {
		return m.FavoritePlaylists
	}
}
