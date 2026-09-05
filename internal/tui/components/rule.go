package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// TopBorderWithLabel renders a box's top border row with label embedded at
// pos. The box's own style needs BorderTop(false) or it overdraws this.
func TopBorderWithLabel(width int, label string, pos lipgloss.Position, b lipgloss.Border, style lipgloss.Style) string {
	left, right := fillAround(width-2, label, pos) // 2 for the corner runes

	return style.Render(b.TopLeft+strings.Repeat(b.Top, left)) +
		label +
		style.Render(strings.Repeat(b.Top, right)+b.TopRight)
}

// Rule renders a plain horizontal divider spanning width, with label
// centered in it, e.g. "──────── 6 files ────────".
func Rule(width int, label string, style lipgloss.Style) string {
	left, right := fillAround(width, label, lipgloss.Center)

	return style.Render(strings.Repeat("─", left)) +
		label +
		style.Render(strings.Repeat("─", right))
}

// fillAround splits the space left over after label into the runs of fill
// that go before and after it, placing label at pos.
func fillAround(width int, label string, pos lipgloss.Position) (left, right int) {
	fill := max(0, width-lipgloss.Width(label))
	left = int(float64(fill) * float64(pos))
	return left, fill - left
}
