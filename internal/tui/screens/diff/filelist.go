package diffview

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// fileList renders one row per file in files, which is the window of
// m.files starting at offset — cursor is still an index into the full list.
func fileList(files []git.FileChange, offset, cursor, scrollOffset, width int) string {
	rows := make([]string, len(files))
	for i, c := range files {
		rows[i] = fileListItem(c, width, offset+i == cursor, scrollOffset)
	}

	return strings.Join(rows, "\n")
}

// fileListItem renders one row: accent bar, change glyph, then the path with
// its directory dimmed. An overlong selected row scrolls; others truncate.
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

	row := bar + s.Row.Render(" ") + glyph + s.Row.Render(" ") + styledPath(name, s)

	// Pad to the full width so a selected row's band reaches the edge.
	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}

// styledPath renders a path with its directory dimmed and its final segment
// bright, so the eye lands on the file name.
func styledPath(path string, s styles.RowStyles) string {
	cut := strings.LastIndex(path, "/")
	if cut < 0 {
		return s.Name.Render(path)
	}

	return s.Dir.Render(path[:cut+1]) + s.Name.Render(path[cut+1:])
}

// scrollWindow returns the maxWidth-wide slice of name visible at
// scrollOffset. A name that fits is returned as-is; a longer one loops.
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

// truncateFront shortens name to maxWidth by dropping characters from the
// front, keeping the tail — usually the identifying part of a path.
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
