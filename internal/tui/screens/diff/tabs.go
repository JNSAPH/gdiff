package diffview

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// tabBarHeight is the tabs plus the divider carrying the sort state.
const tabBarHeight = 2

// Overflow markers, shown when the strip can't hold every tab.
const (
	moreLeft  = "‹"
	moreRight = "›"
)

// tabChrome is what a tab spends on everything but the name.
const tabChrome = 4

// layoutMode is how the file list is laid out. Auto follows the window's width;
// the other two are the toggle key pinning one.
type layoutMode int

const (
	layoutAuto layoutMode = iota
	layoutTabs
	layoutSidebar
)

// narrow reports whether the list shows as a tab strip rather than a sidebar.
func (m Model) narrow() bool {
	// No room here for a sidebar and a readable diff beside it.
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

// tabsHeight is what the strip costs the body, zero when the sidebar shows.
func (m Model) tabsHeight() int {
	if m.narrow() {
		return tabBarHeight
	}
	return 0
}

// tabs renders the strip, then the divider closing it off.
func (m Model) tabs(width int) string {
	return lipgloss.JoinVertical(lipgloss.Left, m.tabStrip(width), m.tabsDivider(width))
}

// tabStrip renders one row of tabs, windowed around the selected file,
// with markers for the ones that didn't fit.
func (m Model) tabStrip(width int) string {
	if len(m.files) == 0 {
		return clamp(styles.Muted.Render(" no files"), width)
	}

	// A name wider than the strip would be sliced mid-band by the clamp below,
	// so it's cut to what the chrome and markers leave.
	room := max(1, width-tabChrome-2)

	labels := make([]string, len(m.files))
	for i, f := range m.files {
		labels[i] = tabLabel(f, i == m.cursor, room)
	}

	// The markers cost two columns, so the window is re-cut for them.
	start, end := tabWindow(labels, m.cursor, width)
	if start > 0 || end < len(labels) {
		start, end = tabWindow(labels, m.cursor, width-2)
	}

	return clamp(marker(moreLeft, start > 0)+
		strings.Join(labels[start:end], " ")+
		marker(moreRight, end < len(labels)), width)
}

// marker is an overflow arrow, or the blank column it would occupy.
func marker(glyph string, show bool) string {
	if !show {
		return " "
	}
	return styles.Subtle.Render(glyph)
}

// tabLabel is the glyph and base name, in the selected band when current.
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

// tabWindow is the run of tabs that fits, filling forward from the cursor.
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

// tabsDivider carries the sort state and position the sidebar's header would
// show, taking the accent color when the tabs have focus.
func (m Model) tabsDivider(width int) string {
	// Only two rows here, so the divider gives way to the input while typing.
	if m.filtering {
		return clamp(m.filter.View(), width)
	}

	label := " " + styles.Subtle.Render("sort ") + styles.Muted.Render(sortOptions[m.sortIndex].label)
	if m.filterActive() {
		label = " " + styles.BrandTitle.Render("/") + styles.Text.Render(m.filter.Value())
	}
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
