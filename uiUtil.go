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
	libraryStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0).Align(lipgloss.Left).PaddingLeft(2)
)