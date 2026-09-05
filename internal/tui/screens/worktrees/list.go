package worktrees

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// list renders one row per entry in entries, which is the window of
// m.entries starting at offset — cursor is still an index into the full list.
func list(entries []entry, offset, cursor, width int) string {
	rows := make([]string, len(entries))
	for i, e := range entries {
		rows[i] = row(e, width, offset+i == cursor)
	}
	return strings.Join(rows, "\n")
}

// row renders one worktree: accent bar, its directory name (path dimmed),
// then its branch and change counts right-aligned.
func row(e entry, width int, selected bool) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	left := bar + s.Row.Render(" ") + entryName(e, s)
	right := entrySummary(e)

	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	row := left + s.Row.Render(strings.Repeat(" ", gap)) + right

	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}

// entryName renders the worktree's path with its parent directory dimmed,
// same treatment as a file's path in the diff view's file list.
func entryName(e entry, s styles.RowStyles) string {
	cut := strings.LastIndex(e.Path, "/")
	if cut < 0 {
		return s.Name.Render(e.Path)
	}
	return s.Dir.Render(e.Path[:cut+1]) + s.Name.Render(e.Path[cut+1:])
}

// entrySummary renders the branch and, if the worktree has uncommitted
// changes, the same "+N ~N -N" counts the diff view's header uses.
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
