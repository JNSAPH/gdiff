package diffview

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// sidebarHeaderHeight is how many rows sit above the file list: the sort
// state and the file-count divider.
const sidebarHeaderHeight = 2

// sidebar renders the left panel: the header, then the visible window of the
// file list with a scrollbar down its right edge.
func (m Model) sidebar(height int) string {
	width := m.sidebarW() - components.SidebarBorderWidth
	rows := max(0, height-sidebarHeaderHeight)

	return components.Sidebar(
		m.sidebarW(),
		height,
		lipgloss.JoinVertical(lipgloss.Left, m.sidebarHeader(width), m.fileRows(width, rows)),
		m.focus == focusSidebar,
	)
}

// fileRows renders the visible slice of the file list, one row per line,
// with the scrollbar occupying the last column.
func (m Model) fileRows(width, rows int) string {
	if rows <= 0 {
		return ""
	}

	end := min(m.listOffset+rows, len(m.files))
	window := m.files[min(m.listOffset, len(m.files)):end]

	list := fileList(window, m.listOffset, m.cursor, m.scrollOffset, width-1)

	// Pair each row with its scrollbar cell, padding out any rows the file
	// list didn't fill so the scrollbar still spans the panel.
	lines := make([]string, rows)
	listLines := strings.Split(list, "\n")

	for i := range lines {
		row := ""
		if i < len(window) && i < len(listLines) {
			row = listLines[i]
		}
		lines[i] = lipgloss.NewStyle().Width(width - 1).Render(row)
	}

	return strings.Join(lines, "\n")
}

// sidebarHeader renders the sort state, then the file count centered in a
// divider.
func (m Model) sidebarHeader(width int) string {
	count := " " + strconv.Itoa(len(m.files)) + " files "
	if len(m.files) == 1 {
		count = " 1 file "
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.sortLabel(width),
		components.Rule(width, styles.Muted.Render(count), styles.BorderLine),
	)
}
