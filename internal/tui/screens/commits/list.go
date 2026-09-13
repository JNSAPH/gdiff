package commits

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// Fixed-width columns, so the subject beside them never shifts.
const (
	hashWidth   = 7
	authorWidth = 14
	dateWidth   = 6
)

// list renders the window of commits at offset; cursor indexes the full list.
func list(commits []git.Commit, offset, cursor, width int) string {
	rows := make([]string, len(commits))
	for i, c := range commits {
		rows[i] = row(c, width, offset+i == cursor)
	}
	return strings.Join(rows, "\n")
}

// row renders one commit: bar, hash and subject, then author and date.
func row(c git.Commit, width int, selected bool) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	right := components.PadLeft(s.Dir.Render(components.TruncateTail(c.Author, authorWidth-1)), authorWidth, s.Row) +
		components.PadLeft(s.Dir.Render(components.Relative(c.When)), dateWidth, s.Row)

	// The subject takes whatever the fixed columns leave, one column held back
	// so it never touches the author.
	left := bar + s.Row.Render(" ") + components.PadLeft(s.Dir.Render(c.Short), hashWidth, s.Row) + s.Row.Render(" ")
	room := max(0, width-lipgloss.Width(left)-lipgloss.Width(right)-1)
	left += s.Name.Render(components.TruncateTail(c.Subject, room))

	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	row := left + s.Row.Render(strings.Repeat(" ", gap)) + right

	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}
