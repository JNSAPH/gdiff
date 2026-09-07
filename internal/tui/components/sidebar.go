package components

import (
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SidebarBorderWidth is what Sidebar's border eats; size content to less.
const SidebarBorderWidth = 1

// Sidebar wraps content in a panel width columns wide, its divider accented
// when the panel has focus.
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
