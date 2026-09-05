package components

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// HeaderHeight is how many rows Header renders — always one, by
// construction (it never contains a newline).
const HeaderHeight = 1

// headerSeparator divides the title bar's segments.
const headerSeparator = " · "

// Header renders the title bar: title plus any non-empty segments, centered
// in the body's top border row. Segments arrive already styled by the caller.
func Header(width int, title string, segments ...string) string {
	parts := make([]string, 0, len(segments)+1)
	if title != "" {
		parts = append(parts, styles.BrandTitle.Render(title))
	}
	for _, s := range segments {
		if s != "" {
			parts = append(parts, s)
		}
	}

	// No content at all: draw an unbroken border rather than a gap in it.
	label := ""
	if len(parts) > 0 {
		label = " " + strings.Join(parts, styles.BorderLine.Render(headerSeparator)) + " "
	}

	return TopBorderWithLabel(width, label, lipgloss.Center, lipgloss.RoundedBorder(), styles.BorderLine)
}
