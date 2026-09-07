package diffview

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// fileList renders the window of files at offset; cursor indexes the full list.
func fileList(files []git.FileChange, offset, cursor, scrollOffset, width int) string {
	rows := make([]string, len(files))
	for i, c := range files {
		rows[i] = fileListItem(c, width, offset+i == cursor, scrollOffset)
	}

	return strings.Join(rows, "\n")
}

// fileListItem renders one row: bar, glyph, then the path with its directory
// dimmed. An overlong selected row scrolls; others truncate.
func fileListItem(file git.FileChange, width int, selected bool, scrollOffset int) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	glyph := s.Glyph.Foreground(components.ChangeTypeColor(file.Type)).Render(file.Type.Symbol())

	// The bar, glyph and their two spaces take a fixed 4 columns.
	available := max(0, width-4)

	name := file.Name()
	if selected {
		name = scrollWindow(name, scrollOffset, available)
	} else {
		name = truncateFront(name, available)
	}

	row := bar + s.Row.Render(" ") + glyph + s.Row.Render(" ") + components.StyledPath(name, s)

	// Pad to the full width so a selected row's band reaches the edge.
	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}

// scrollWindow is the maxWidth slice of name at scrollOffset; long names loop.
func scrollWindow(name string, scrollOffset, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= maxWidth {
		return name
	}

	const gap = "      "
	loop := append(runes, []rune(gap)...)

	start := scrollOffset % len(loop)
	window := make([]rune, maxWidth)
	for i := range window {
		window[i] = loop[(start+i)%len(loop)]
	}

	return string(window)
}

// truncateFront drops characters from the front, keeping a path's tail.
func truncateFront(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= maxWidth {
		return name
	}

	const ellipsis = "…"
	keep := maxWidth - 1 // the ellipsis takes one column
	if keep <= 0 {
		return ellipsis
	}

	return ellipsis + string(runes[len(runes)-keep:])
}

// truncateTail drops them from the end, for a name whose front identifies it.
func truncateTail(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= maxWidth {
		return name
	}

	const ellipsis = "…"
	keep := maxWidth - 1 // the ellipsis takes one column
	if keep <= 0 {
		return ellipsis
	}

	return string(runes[:keep]) + ellipsis
}
