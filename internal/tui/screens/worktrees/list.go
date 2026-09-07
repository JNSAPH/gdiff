package worktrees

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// list renders the window of entries at offset; cursor indexes the full list.
func list(entries []entry, offset, cursor, width int, active string) string {
	rows := make([]string, len(entries))
	for i, e := range entries {
		rows[i] = row(e, width, offset+i == cursor, e.Path == active)
	}
	return strings.Join(rows, "\n")
}

// row renders one worktree: bar and directory, then branch and counts right.
func row(e entry, width int, selected, active bool) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	left := bar + s.Row.Render(" ") + components.StyledPath(e.Path, s)
	right := activeMark(active, s) + entrySummary(e)

	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	row := left + s.Row.Render(strings.Repeat(" ", gap)) + right

	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}

// activeMark flags the worktree the diff view is on, or holds its column so
// the branches stay aligned.
func activeMark(active bool, s styles.RowStyles) string {
	if !active {
		return s.Row.Render("  ")
	}
	return s.Accent.Render("●") + s.Row.Render(" ")
}

// entrySummary is the branch plus, if dirty, the header's "+N ~N -N" counts.
func entrySummary(e entry) string {
	branch := styles.Muted.Render(e.Branch)

	if e.loadErr != nil {
		return branch + "  " + styles.Error.Render("error reading worktree")
	}

	if counts := components.ChangeCounts(e.counts); counts != "" {
		return branch + "  " + counts
	}
	return branch
}
