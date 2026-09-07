package diffview

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// tabBarHeight is how many rows the tab strip takes: the tabs and the
// divider that carries the sort state, mirroring the sidebar's header.
const tabBarHeight = 2

// Overflow markers, shown when the strip can't hold every tab.
const (
	moreLeft  = "‹"
	moreRight = "›"
)

// tabChrome is the columns a tab spends on everything but the name: the
// accent bar, the change glyph and the space either side of it.
const tabChrome = 4

// layoutMode is how the file list is laid out. Auto follows the window's
// width and is where the screen starts; the other two are the toggle key
// pinning one layout.
type layoutMode int

const (
	layoutAuto layoutMode = iota
	layoutTabs
	layoutSidebar
)

// narrow reports whether the file list shows as a tab strip on top rather
// than a sidebar beside the diff.
func (m Model) narrow() bool {
	// Tabs regardless of the mode: there isn't room here for a sidebar and
	// a readable diff next to it.
	if m.width < minSidebarWidth+minStatusWidth {
		return true
	}

	switch m.layout {
	case layoutTabs:
		return true
	case layoutSidebar:
		return false
	default:
		return m.width < narrowWidth
	}
}

// tabsHeight is the rows the tab strip costs the body, zero when the sidebar
// is showing instead.
func (m Model) tabsHeight() int {
	if m.narrow() {
		return tabBarHeight
	}
	return 0
}

// tabs renders the file list as a strip above the diff: the tabs themselves,
// then the divider closing them off.
func (m Model) tabs(width int) string {
	return lipgloss.JoinVertical(lipgloss.Left, m.tabStrip(width), m.tabsDivider(width))
}

// tabStrip renders one row of tabs, windowed so the selected file is always
// among them, with markers on either side for the ones that didn't fit.
func (m Model) tabStrip(width int) string {
	if len(m.files) == 0 {
		return clamp(styles.Muted.Render(" no files"), width)
	}

	// A name wider than the strip would be sliced mid-band by the clamp
	// below, so it's cut to what's left once the chrome and both overflow
	// markers have taken their columns.
	room := max(1, width-tabChrome-2)

	labels := make([]string, len(m.files))
	for i, f := range m.files {
		labels[i] = tabLabel(f, i == m.cursor, room)
	}

	// The markers cost two columns, so the window is re-cut to make room
	// once it's clear the tabs don't all fit.
	start, end := tabWindow(labels, m.cursor, width)
	if start > 0 || end < len(labels) {
		start, end = tabWindow(labels, m.cursor, width-2)
	}

	return clamp(marker(moreLeft, start > 0)+
		strings.Join(labels[start:end], " ")+
		marker(moreRight, end < len(labels)), width)
}

// marker is an overflow arrow, or the blank column it would occupy, so the
// tabs sit at the same offset either way.
func marker(glyph string, show bool) string {
	if !show {
		return " "
	}
	return styles.Subtle.Render(glyph)
}

// tabLabel is one tab: the change-type glyph and the file's base name,
// wearing the selected row's band and accent bar while it's the current file.
func tabLabel(file git.FileChange, selected bool, room int) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	glyph := s.Glyph.Foreground(components.ChangeTypeColor(file.Type)).Render(file.Type.Symbol())

	return bar + glyph + s.Row.Render(" ") + s.Name.Render(truncateTail(baseName(file), room)) + s.Row.Render(" ")
}

// tabWindow is the run of tabs that fits in width with the cursor's tab in
// it. It fills forward from the cursor first, then backwards with the rest.
func tabWindow(labels []string, cursor, width int) (start, end int) {
	used := lipgloss.Width(labels[cursor])
	start, end = cursor, cursor+1

	for end < len(labels) {
		w := lipgloss.Width(labels[end]) + 1 // the joining space
		if used+w > width {
			break
		}
		used += w
		end++
	}

	for start > 0 {
		w := lipgloss.Width(labels[start-1]) + 1
		if used+w > width {
			break
		}
		used += w
		start--
	}

	return start, end
}

// tabsDivider separates the tabs from the diff and carries the sort state and
// cursor position the sidebar's header would show. It takes the accent color
// when the tabs have focus, standing in for the sidebar's focused border.
func (m Model) tabsDivider(width int) string {
	label := " " + styles.Subtle.Render("sort ") + styles.Muted.Render(sortOptions[m.sortIndex].label)
	if len(m.files) > 0 {
		label += styles.BorderLine.Render(" · ") +
			styles.Subtle.Render(strconv.Itoa(m.cursor+1)+"/"+strconv.Itoa(len(m.files)))
	}
	label += " "

	line := styles.BorderLine
	if m.focus == focusSidebar {
		line = styles.FocusedBorderLine
	}

	return components.Rule(width, label, line)
}
