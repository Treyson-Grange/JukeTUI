package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Align(lipgloss.Center)

	horizontalGap = lipgloss.NewStyle().Padding(0, 1)
)