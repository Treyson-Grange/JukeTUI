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

const SPOTIFY_GREEN = "#1DB954"
const UI_LIBRARY_SPACE = 7 // Space to subtract from total to get library space
const CHARACTERS = 7       // Characters we have to account for when truncating
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

// Text formatting functions

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
func getUiElements(m Model, boxWidth int) (string, string, string, string) {
	return getLibText(m, boxWidth), getPlayBack(m), getReccDetails(m), getVisualQueue(m, boxWidth)
}

// Get the library text for display
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
			libText += fmt.Sprintf("%s%s - %s%s\n", favorite, item.name, item.artist, play)
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
	shuffle := "!Shuffle"
	if m.state.ShuffleState {
		shuffle = "Shuffle"
	}

	if m.state.Item.Artists != nil {
		return bracketWrap(m.state.Item.Name+" | "+m.state.Item.Artists[0].Name) + bracketWrap(lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN)).Render(status)) + bracketWrap(msToMinSec(m.progressMs)+" / "+msToMinSec(m.state.Item.DurationMs)) + bracketWrap(shuffle)
	} else {
		return "No Playback Data. Please start a playback session on your device"
	}
}

// Get the reccomendation details for display
func getReccDetails(m Model) string {
	recommendationDetails := "Press '" + keybinds["Recommendation"] + "' for a recommendation!\n\n" + m.image + "\n"
	if len(m.reccomendation.Tracks) > 0 {
		recommendationDetails += fmt.Sprintf(
			"%s - %s\n '%s' to add to your queue!", m.reccomendation.Tracks[0].Name, m.reccomendation.Tracks[0].Artists[0].Name, keybinds["Add to Queue"],
		)
	}
	return recommendationDetails
}

func getVisualQueue(m Model, boxWidth int) string {
	queue := "Queue:\n"// Queue is now a list of queueitems
	for _, item := range m.queue.Queue {
		nameLen := len(item.Name)
		artistLen := len(item.Artists[0].Name)

		if 3 + nameLen + artistLen > boxWidth {
			item.Name = truncate(item.Name, boxWidth - 3 - artistLen)
		}
		queue += fmt.Sprintf("%s - %s\n", item.Name, item.Artists[0].Name)
	}
	return queue
}