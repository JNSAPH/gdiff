package diffview

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// tabWidth is fixed so the gutter stays aligned whatever the terminal does.
const tabWidth = 4

// content renders the right-hand pane: the file's path, then the viewport.
func (m Model) content() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.contentHeader(m.contentWidth()),
		m.viewport.View(),
	)
}

// sidebarW is the sidebar's width. It gives way as the terminal narrows so the
// diff pane stays usable, and takes nothing at all once the tabs take over.
func (m Model) sidebarW() int {
	if m.narrow() {
		return 0
	}
	return min(sidebarWidth, max(minSidebarWidth, m.width/3))
}

// contentWidth is the pane's total width, sidebar excluded.
func (m Model) contentWidth() int {
	return max(0, m.width-m.sidebarW())
}

// contentHeader shows the file's path, its accent bar lit when focused.
func (m Model) contentHeader(width int) string {
	bar := styles.BorderLine.Render("▏")
	if m.focus == focusContent {
		bar = styles.FocusedBorderLine.Render("▏")
	}

	// Nothing selected: the centered message in the pane below says so.
	file, ok := m.selected()
	if !ok {
		return bar
	}

	// Path left, counts and scroll right; the path truncates into what's left.
	prefix := bar + " " + m.changeGlyph(file.Type) + " "

	// On a narrow pane the file name matters more than the counts.
	status := ""
	if width >= minStatusWidth {
		status = m.diffStats() + "  " + styles.Subtle.Render(m.scrollPercent())
	}

	// Two columns held back: the trailing space, and one before the status.
	room := max(0, width-lipgloss.Width(prefix)-lipgloss.Width(status)-2)
	label := prefix + components.StyledPath(truncateFront(file.Name(), room), styles.RowHeader)

	gap := max(0, width-lipgloss.Width(label)-lipgloss.Width(status)-1)

	return clamp(label+strings.Repeat(" ", gap)+status+" ", width)
}

// clamp truncates s to width columns so a header can't wrap onto a second row.
func clamp(s string, width int) string {
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

// scrollPercent is right-aligned so the counts beside it don't shift.
func (m Model) scrollPercent() string {
	if len(m.diffLines) <= m.viewport.Height() {
		return strings.Repeat(" ", statusWidth)
	}

	return pad(strconv.Itoa(int(m.viewport.ScrollPercent()*100))+"%", statusWidth)
}

// changeGlyph is the one-character marker for a change type, colored.
func (m Model) changeGlyph(t git.ChangeType) string {
	return lipgloss.NewStyle().
		Foreground(components.ChangeTypeColor(t)).
		Bold(true).
		Render(t.Symbol())
}

// diffStats summarizes the current diff as "+12 -3".
func (m Model) diffStats() string {
	var added, deleted int
	for _, l := range m.diffLines {
		switch l.Type {
		case git.LineAdded:
			added++
		case git.LineDeleted:
			deleted++
		}
	}

	if added == 0 && deleted == 0 {
		return ""
	}

	return lipgloss.NewStyle().Foreground(styles.AddedColor).Render("+"+strconv.Itoa(added)) +
		" " +
		lipgloss.NewStyle().Foreground(styles.DeletedColor).Render("-"+strconv.Itoa(deleted))
}

// renderDiff loads m.diffLines into the viewport, padding each to the pane's
// width so its background band spans the row.
func (m Model) renderDiff() Model {
	// No diff — center the message and clear the gutter so it isn't indented.
	if len(m.diffLines) == 0 {
		m.viewport.LeftGutterFunc = nil
		m.viewport.StyleLineFunc = nil
		m.viewport.SetContent(lipgloss.Place(
			m.viewport.Width(), m.viewport.Height(),
			lipgloss.Center, lipgloss.Center, m.message,
		))
		return m
	}

	m.viewport.LeftGutterFunc = m.diffGutter
	m.viewport.StyleLineFunc = m.diffLineStyle

	lines := make([]string, len(m.diffLines))
	longest := 0
	for i, l := range m.diffLines {
		lines[i] = strings.ReplaceAll(l.Text, "\t", strings.Repeat(" ", tabWidth))
		longest = max(longest, lipgloss.Width(lines[i]))
	}

	// Pad to the longest line, not the pane: scrolling sideways moves past the
	// pane's width, and a band that stopped there would run out.
	width := max(m.viewport.Width()-m.gutterWidth(), longest)
	for i, text := range lines {
		if pad := width - lipgloss.Width(text); pad > 0 {
			lines[i] = text + strings.Repeat(" ", pad)
		}
	}

	m.viewport.SetContentLines(lines)
	return m
}

// diffLineStyle colors one diff row by its type.
func (m Model) diffLineStyle(i int) lipgloss.Style {
	if i < 0 || i >= len(m.diffLines) {
		return styles.ContextLine
	}

	switch m.diffLines[i].Type {
	case git.LineAdded:
		return styles.AddedLine
	case git.LineDeleted:
		return styles.DeletedLine
	default:
		return styles.ContextLine
	}
}

// diffGutter renders one row's numbers and sign. The viewport calls it with
// a zero context to measure, so every return must be the same width.
func (m Model) diffGutter(ctx viewport.GutterContext) string {
	if len(m.diffLines) == 0 {
		return ""
	}

	// Styling the numbers in the gutter
	nums := func(oldNo, newNo string) string {
		return styles.GutterNum.Render(" "+pad(oldNo, m.lineNumWidth)+" "+pad(newNo, m.lineNumWidth)) +
			styles.GutterSep.Render(" ▏")
	}

	// A soft-wrapped or out-of-range row gets a blank gutter of equal width.
	if ctx.Soft || ctx.Index < 0 || ctx.Index >= len(m.diffLines) {
		return nums("", "") + "  "
	}

	line := m.diffLines[ctx.Index]
	oldNo, newNo := "", ""
	if line.Old > 0 {
		oldNo = strconv.Itoa(line.Old)
	}
	if line.New > 0 {
		newNo = strconv.Itoa(line.New)
	}

	// Decide which sign to use for the gutter
	sign := "  "
	switch line.Type {
	case git.LineAdded:
		sign = styles.GutterAdded.Render("+") + " "
	case git.LineDeleted:
		sign = styles.GutterDeleted.Render("-") + " "
	}

	return nums(oldNo, newNo) + sign
}

// gutterWidth renders an empty gutter, so it can't drift from the real one.
func (m Model) gutterWidth() int {
	return lipgloss.Width(m.diffGutter(viewport.GutterContext{}))
}

// pad right-aligns s in width columns.
func pad(s string, width int) string {
	if n := width - lipgloss.Width(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

// numWidth is the columns needed for the largest line number in lines.
func numWidth(lines []git.DiffLine) int {
	highest := 0
	for _, l := range lines {
		highest = max(highest, l.Old, l.New)
	}

	return max(2, len(strconv.Itoa(highest)))
}
