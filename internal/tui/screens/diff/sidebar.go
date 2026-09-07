package diffview

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// sidebar renders the left panel: the header, then the visible window of the
// file list.
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
// padded out to the panel's full height.
func (m Model) fileRows(width, rows int) string {
	if rows <= 0 {
		return ""
	}

	end := min(m.listOffset+rows, len(m.files))
	window := m.files[min(m.listOffset, len(m.files)):end]

	list := fileList(window, m.listOffset, m.cursor, m.scrollOffset, width-1)

	// Every row is padded to the same width, and rows the file list didn't
	// fill are left blank, so a short list still fills the panel.
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

// sidebarHeader renders the sort state or filter, then the file count.
func (m Model) sidebarHeader(width int) string {
	count := " " + strconv.Itoa(len(m.files)) + " files "
	if len(m.files) == 1 {
		count = " 1 file "
	}

	// One row, two things that want it: the filter wins while it's in play.
	first := m.sortLabel(width)
	if m.filterActive() {
		first = m.filterRow(width)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		first,
		components.Rule(width, styles.Muted.Render(count), styles.BorderLine),
	)
}

// listPosition is the cursor's place in the file list, e.g. "3/32".
func (m Model) listPosition() string {
	if len(m.files) == 0 {
		return ""
	}
	return styles.Subtle.Render(strconv.Itoa(m.cursor+1) + "/" + strconv.Itoa(len(m.files)))
}

// spread pushes left and right to opposite ends of a row width columns wide.
func spread(left, right string, width int) string {
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}
