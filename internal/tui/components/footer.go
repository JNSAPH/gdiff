package components

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// NewHelp builds a screen's footer help bar in the app's palette.
func NewHelp() help.Model {
	h := help.New()
	h.Styles.ShortKey = styles.Text.Bold(true)
	h.Styles.ShortDesc = styles.Muted
	h.Styles.ShortSeparator = styles.Subtle
	h.Styles.FullKey = styles.Text.Bold(true)
	h.Styles.FullDesc = styles.Muted
	h.Styles.FullSeparator = styles.Subtle
	h.Styles.Ellipsis = styles.Subtle
	h.ShortSeparator = "   "
	return h
}

// Footer renders a screen's bottom help bar: a top border, then h's view of
// keys.
func Footer(width int, h help.Model, keys help.KeyMap) string {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Render(h.View(keys))
}
