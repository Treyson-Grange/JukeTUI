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

// Horrid ASCII art of a jukebox based on the screen size.
func GetAsciiJuke(boxWidth, boxHeight int) string {
	var space int
	var vertSpace int
	var recommendationDetails string
	if boxWidth < 50 {
		recommendationDetails = "Press 'r' to get a recommendation!\n\n\n" + `

             @@@@@@@@@             
          @@           @@          
       :@                 @-       
`
		space = 17
		vertSpace = boxHeight / 4
	} else {
		recommendationDetails = `Press 'r' to get a recommendation!\n\n\n
                                             
                            +@@@@@@@@@@@@@@@@@@=                            
                        @@@@@%                %@@@@@                        
                     @@@@=                        -@@@@                     
                   @@@#                              *@@@                   
                 @@@                                    @@@                 
               @@@                                        @@@               
              @@#                                          *@@      	     
`
		space = 47
		vertSpace = boxHeight / 2
	}

	for i := 0; i < vertSpace; i++ {
		recommendationDetails += "@"
		for j := 0; j < space; j++ {
			if j%10 == 0 {
				recommendationDetails += "@"
			} else {
				recommendationDetails += " "
			}
		}
		recommendationDetails += "@\n"
	}

	for i := 0; i < space+2; i++ {
		recommendationDetails += "@"
	}
	recommendationDetails += "\n"

	return recommendationDetails
}

const SPOTIFY_GREEN = "#1DB954"
