package util

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // These aren't used directly, but are required for image.Decode to work
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"strings"

	"github.com/treyson-grange/JukeTUI/src/types"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/image/draw"

	moji "github.com/Treyson-Grange/go-moji-ui"
)

const SPOTIFY_GREEN = "#1DB954"
const UI_LIBRARY_SPACE = 7 // Space to subtract from total to get library space
const CHARACTERS = 8       // Characters we have to account for when truncating
const LIBRARY_SPACING = 13 // TODO: Make this dynamic | Bigger number, less displayed

var (
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1).
			Align(lipgloss.Center)

	LibraryStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0).Align(lipgloss.Left)

	bold = lipgloss.NewStyle().Bold(true)

	green = lipgloss.NewStyle().Foreground(lipgloss.Color(SPOTIFY_GREEN))

	highlight = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: SPOTIFY_GREEN}

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
)

// =======================
// === Text Formatting ===
// =======================

// Truncate a string to fit any width
func Truncate(str string, width int) string {
	if len(str) > width {
		if width > 5 {
			return str[:width-5] + "..."
		}
	}
	return str
}

// Wrap a string in brackets
func BracketWrap(str string) string {
	return fmt.Sprintf(" [ %s ] ", str)
}

// Turn ms to MM:SS format
func MsToMinSec(ms int) string {
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// =====================================
// ===== Album Cover Functionality =====
// =====================================

// Given a color, return the ANSI color code
func BgAnsiColor(c color.Color) string {
	r, g, b, _ := c.RGBA()                                      // no alpha
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8) // 16-bit color to 8-bit
}

// Resize an image to a given width and height
func ResizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if img == nil || dst == nil {
		return nil
	}
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// Print an image to display
func PrintImage(img image.Image) string {
	bounds := img.Bounds()
	var result strings.Builder

	// For every pixel in the image, get the color and print it
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var line strings.Builder
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := img.At(x, y)
			line.WriteString(fmt.Sprintf("%s  \x1b[0m", BgAnsiColor(color)))
		}
		result.WriteString(line.String() + "\n")
	}
	return result.String()
}

// Simple fetch for an image given a URL
func FetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// Handler for fetching an image, resizing it, and returning it as a string
func MakeNewImage(url string, width, height int) string {
	img, err := FetchImage(url)
	if err != nil {
		return "Error fetching image"
	}

	albumSize := math.Min(float64(width), float64(height))
	return PrintImage(ResizeImage(img, int(albumSize*11/10), int(albumSize)))
}

// =======================
// ===== UI Elements =====
// =======================

// Get the UI elements for display
func GetUiElements(m *types.Model, boxWidth, maxQueueHeight int) (string, string, string, string) {
	return GetLibText(m, boxWidth), GetPlayBack(m, boxWidth), m.Image, GetVisualQueue(m, boxWidth, maxQueueHeight)
}

// Generate the library text for display
func GetLibText(m *types.Model, boxWidth int) string {
	libText := ""

	libText += GetTabs(m, boxWidth) + "\n"

	if m.LibraryList == nil {
		return "Loading Library Data..."
	}
	totalPages := int(math.Ceil(float64(m.ApiTotal)/float64(m.Height-UI_LIBRARY_SPACE-len(GetFavorites(m))))) + 1
	currentPage := int(math.Ceil(float64(m.Offset)/float64(m.Height-UI_LIBRARY_SPACE-len(GetFavorites(m))))) + 1
	libText += bold.Render(fmt.Sprintf("Page %d of %d", currentPage, totalPages))
	if m.Loading {
		libText += "  Loading..."
	}
	libText += "\n"
	if m.LibraryList != nil {
		for i, item := range m.LibraryList {
			if i == m.Cursor {
				item = types.LibraryItem{
					Name:     green.Render("> " + bold.Render(Truncate(item.Name, boxWidth-len(item.Artist)-CHARACTERS))),
					Artist:   item.Artist,
					URI:      item.URI,
					Favorite: item.Favorite,
				}
			} else {
				item = types.LibraryItem{
					Name:     "  " + Truncate(item.Name, boxWidth-len(item.Artist)-CHARACTERS),
					Artist:   item.Artist,
					URI:      item.URI,
					Favorite: item.Favorite,
				}
			}
			play := map[bool]string{true: " 🔊", false: ""}[m.State.Context.URI == item.URI]
			favorite := map[bool]string{true: "♥ ", false: "  "}[item.Favorite]
			libText += fmt.Sprintf("%s%s - %s%s\n", favorite, moji.FilterEmojisBySize(item.Name, 2), item.Artist, play)
		}
	}
	return libText
}

// Generate the playback text for display
func GetPlayBack(m *types.Model, width int) string {
	if m.State.Item.Artists == nil {
		return "No Playback Data. Please start a playback session on your device"
	}
	status := "▶ "
	if m.State.IsPlaying {
		status = "▮▮"
	}
	shuffle := "!Shuffle"
	if m.State.ShuffleState {
		shuffle = "Shuffle"
	}

	progress := bold.Render(MsToMinSec(m.ProgressMs) + " / " + MsToMinSec(m.State.Item.DurationMs))
	statusRendered := green.Render(status)

	return BracketWrap(Truncate(m.State.Item.Name+" | "+m.State.Item.Artists[0].Name, width)) +
		BracketWrap(statusRendered) +
		BracketWrap(progress) +
		BracketWrap(shuffle)
}

// Generate the visual queue for display
func GetVisualQueue(m *types.Model, boxWidth, maxItems int) string {
	queue := bold.Render("Queue:\n")
	queueLen := len(m.Queue.Queue)
	if queueLen > maxItems {
		queueLen = maxItems
	}
	SEP := " - "
	for i := 0; i < queueLen; i++ {
		item := m.Queue.Queue[i]
		nameLen := len(item.Name)
		artistLen := len(item.Artists[0].Name)
		queueItem := ""
		if len(SEP)+nameLen+artistLen > boxWidth {
			queueItem = Truncate(item.Name, boxWidth-artistLen-len(SEP)) + SEP + item.Artists[0].Name
		} else {
			queueItem = item.Name + SEP + item.Artists[0].Name
		}
		queue += fmt.Sprintf("%s", queueItem)
		if i < queueLen-1 {
			queue += "\n"
		}
	}
	return queue
}

// Generate the tabs for the library with the active tab highlighted
func GetTabs(m *types.Model, boxWidth int) string {
	currentTab := m.ListDetail
	albumTab := activeTab
	playlistTab := tab

	if currentTab == "playlist" {
		albumTab, playlistTab = tab, activeTab
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		albumTab.Render("1) Albums"),
		playlistTab.Render("2) Playlists"),
	)
	gap := tabGap.Render(strings.Repeat(" ", max(0, boxWidth-lipgloss.Width(row)-2)))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

}
