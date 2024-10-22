package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// Render.go
// This file is really just for testing, I'm not sure if I want to use
// A library to render everything, or just do it myself.

func GetTerminalWidth() (int, error) {
	cmd := exec.Command("tput", "cols") // Run tput command to get the columns
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	width, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, err
	}
	return width, nil
}

func GetTerminalHeight() (int, error) {
	fd := int(os.Stdin.Fd())
	_, height, err := term.GetSize(fd)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func test() {
	width, err := GetTerminalWidth()
	if err != nil {
		fmt.Println("Error fetching terminal width:", err)
		return
	}
	height, err := GetTerminalHeight()
	if err != nil {
		fmt.Println("Error fetching terminal width:", err)
		return
	}
	fmt.Println(strings.Repeat("▁", width))
	for i := 0; i < height; i++ {
		fmt.Println("|")
	}

}
