package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // These aren't used directly, but are required for image.Decode to work
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/image/draw"
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
		if width > 5 {
			return str[:width-5] + "..."
		}
	}
	return str
}

// Turn ms to 5:30 format
func msToMinSec(ms int) string {
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

const SPOTIFY_GREEN = "#1DB954"
const UI_LIBRARY_SPACE = 8 // Space to subtract from total to get library space
const CHARACTERS = 6       // Characters we have to account for when truncating

// Album Cover Functionality

// given a color, return the ANSI color code
func bgAnsiColor(c color.Color) string {
	r, g, b, _ := c.RGBA()                                      // no alpha
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8) // 16-bit color to 8-bit
}

// Resize an image to a given width and height
func resizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if img == nil || dst == nil {
		return nil
	}
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// Print an image to display
func printImage(img image.Image) string {
	bounds := img.Bounds()
	var result strings.Builder

	// For every pixel in the image, get the color and print it
	// Intensive operation, MAKE this only happen once.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var line strings.Builder
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := img.At(x, y)
			line.WriteString(fmt.Sprintf("%s  \x1b[0m", bgAnsiColor(color)))
		}
		result.WriteString(line.String() + "\n")
	}
	return result.String()
}

// Simple fetch for an image given a URL
func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// Handler for fetching an image, resizing it, and returning it as a string
func makeNewImage(url string) string {
	img, err := fetchImage(url)
	if err != nil {
		fmt.Println("Error:", err)
	}
	const targetWidth, targetHeight = 15, 15

	return printImage(resizeImage(img, targetWidth, targetHeight))
}

// UI Elements

// Get the UI elements for display
// Returns the library text, playback text, and recommendation details
func getUiElements(m Model, boxWidth int) (string, string, string) {
	libText := getLibText(m, boxWidth)
	playback := getPlayBack(m)
	reccDetails := getReccDetails(m)
	return libText, playback, reccDetails
}

// Get the library text for display
func getLibText(m Model, boxWidth int) string {
	libText := ""
	page := m.offset / (m.height - UI_LIBRARY_SPACE)
	totalPage := m.apiTotal / (m.height - UI_LIBRARY_SPACE)
	libText += fmt.Sprintf("Page %d of %d", page+1, totalPage+1)
	if m.loading {
		libText += "  Loading..."
	}
	libText += "\n"
	if m.libraryList != nil {
		for i, item := range m.libraryList {
			if i == m.cursor {
				item = LibraryItem{
					name:   lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render("> " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS)),
					artist: item.artist,
					uri:    item.uri,
				}
			} else {
				item = LibraryItem{
					name:   "  " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS),
					artist: item.artist,
					uri:    item.uri,
				}
			}
			play := map[bool]string{true: " 🔊", false: ""}[m.state.Context.URI == item.uri]
			libText += fmt.Sprintf("%s - %s%s\n", item.name, item.artist, play)
		}
	}
	return libText
}

// Get the playback text for display
func getPlayBack(m Model) string {
	status := "▶ "
	if m.state.IsPlaying {
		status = "▮▮"
	}
	playback := ""
	if m.state.Item.Artists != nil {
		playback = "🎵 [ " + m.state.Item.Name + " | " + m.state.Item.Artists[0].Name + " ]  " + lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render(status) + "  [ " + msToMinSec(m.progressMs) + " / " + msToMinSec(m.state.Item.DurationMs) + " ]"
	} else {
		playback = "No Playback Data. Please start a playback session on your device"
	}
	return playback
}

// Get the reccomendation details for display
func getReccDetails(m Model) string {
	recommendationDetails := "Press 'r' for a recommendation!\n\n\n" + m.image + "\n"
	if len(m.reccomendation.Tracks) > 0 {
		recommendationDetails += fmt.Sprintf(
			"Recommendation: %s - %s\n 'c' to add to your queue!", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name,
		)
	}
	return recommendationDetails
}
