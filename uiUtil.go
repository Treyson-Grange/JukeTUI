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

	moji "github.com/Treyson-Grange/go-moji-ui"
)

const SPOTIFY_GREEN = "#1DB954"
const UI_LIBRARY_SPACE = 7 // Space to subtract from total to get library space
const CHARACTERS = 8       // Characters we have to account for when truncating
const LIBRARY_SPACING = 10

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1).
			Align(lipgloss.Center)

	libraryStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0).Align(lipgloss.Left)
)

// =======================
// === Text Formatting ===
// =======================

// Truncate a string to fit any width
func truncate(str string, width int) string {
	if len(str) > width {
		if width > 5 {
			return str[:width-5] + "..."
		}
	}
	return str
}

// Wrap a string in brackets
func bracketWrap(str string) string {
	return fmt.Sprintf(" [ %s ] ", str)
}

// Turn ms to 5:30 format
func msToMinSec(ms int) string {
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// =====================================
// ===== Album Cover Functionality =====
// =====================================

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
		return "Error fetching image"
	}
	const targetWidth, targetHeight = 20, 20

	return printImage(resizeImage(img, targetWidth, targetHeight))
}

// =======================
// ===== UI Elements =====
// =======================

// Get the UI elements for display
func getUiElements(m Model, boxWidth int) (string, string, string, string) {
	return getLibText(m, boxWidth), getPlayBack(m), m.image, getVisualQueue(m, boxWidth)
}

// Generate the library text for display
func getLibText(m Model, boxWidth int) string {
	libText := ""
	if m.libraryList == nil {
		return "Loading Library Data..."
	}
	libText += fmt.Sprintf("Page %d of %d", m.offset/(m.height-UI_LIBRARY_SPACE-len(m.favorites))+1, m.apiTotal/(m.height-UI_LIBRARY_SPACE)+1)
	if m.loading {
		libText += "  Loading..."
	}
	libText += "\n"
	if m.libraryList != nil {
		for i, item := range m.libraryList {
			if i == m.cursor {
				item = LibraryItem{
					name:     lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render("> " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS)),
					artist:   item.artist,
					uri:      item.uri,
					favorite: item.favorite,
				}
			} else {
				item = LibraryItem{
					name:     "  " + truncate(item.name, boxWidth-len(item.artist)-CHARACTERS),
					artist:   item.artist,
					uri:      item.uri,
					favorite: item.favorite,
				}
			}
			play := map[bool]string{true: " 🔊", false: ""}[m.state.Context.URI == item.uri]
			favorite := map[bool]string{true: "♥ ", false: "  "}[item.favorite]
			libText += fmt.Sprintf("%s%s - %s%s\n", favorite, moji.FilterEmojisBySize(item.name, 2), item.artist, play)
		}
	}
	return libText
}

// Generate the playback text for display
func getPlayBack(m Model) string {
	if m.state.Item.Artists == nil {
		return "No Playback Data. Please start a playback session on your device"
	}
	status := "▶ "
	if m.state.IsPlaying {
		status = "▮▮"
	}
	shuffle := "!Shuffle"
	if m.state.ShuffleState {
		shuffle = "Shuffle"
	}

	progress := msToMinSec(m.progressMs) + " / " + msToMinSec(m.state.Item.DurationMs)
	statusRendered := lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render(status)

	return bracketWrap(m.state.Item.Name + " | " + m.state.Item.Artists[0].Name) +
		bracketWrap(statusRendered) +
		bracketWrap(progress) +
		bracketWrap(shuffle)

}

// Generate the visual queue for display
func getVisualQueue(m Model, boxWidth int) string {
	queue := "Queue:\n"
	queueLen := len(m.queue.Queue)
	SEP := " - "
	for i, item := range m.queue.Queue {
		nameLen := len(item.Name)
		artistLen := len(item.Artists[0].Name)
		if len(SEP)+nameLen+artistLen > boxWidth {
			item.Name = truncate(item.Name, boxWidth-3-artistLen)
		}
		queue += fmt.Sprintf("%s - %s", item.Name, item.Artists[0].Name)
		if i < queueLen-1 {
			queue += "\n"
		}
	}
	return queue
}
