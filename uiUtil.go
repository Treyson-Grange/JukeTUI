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
		return str[:width-5] + "..."
	}
	return str
}

// Turn ms to 5:30 format
func msToMinSec(ms int) string {
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

const SPOTIFY_GREEN = "#1DB954"

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

// Print an image to the terminal
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
