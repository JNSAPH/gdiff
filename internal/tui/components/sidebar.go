package components

import (
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SidebarBorderWidth is how many of Sidebar's width columns its own border
// eats. Size content to width-SidebarBorderWidth or it wraps.
const SidebarBorderWidth = 1

// Sidebar wraps content in a left-hand panel exactly width columns wide,
// its right-edge divider taking the accent color when the panel has focus.
func Sidebar(width, height int, content string, focused bool) string {
	divider := styles.BorderColor
	if focused {
		divider = styles.FocusedBorderColor
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(divider).
		Render(content)
}
