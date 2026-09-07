package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// TopBorderWithLabel draws a box's top border with label at pos. The box needs
// BorderTop(false) or it overdraws this.
func TopBorderWithLabel(width int, label string, pos lipgloss.Position, b lipgloss.Border, style lipgloss.Style) string {
	left, right := fillAround(width-2, label, pos) // 2 for the corner runes

	return style.Render(b.TopLeft+strings.Repeat(b.Top, left)) +
		label +
		style.Render(strings.Repeat(b.Top, right)+b.TopRight)
}

// Rule is a divider spanning width with label centered, "──── 6 files ────".
func Rule(width int, label string, style lipgloss.Style) string {
	left, right := fillAround(width, label, lipgloss.Center)

	return style.Render(strings.Repeat("─", left)) +
		label +
		style.Render(strings.Repeat("─", right))
}

// fillAround splits the leftover space into the runs before and after label.
func fillAround(width int, label string, pos lipgloss.Position) (left, right int) {
	fill := max(0, width-lipgloss.Width(label))
	left = int(float64(fill) * float64(pos))
	return left, fill - left
}
