package branches

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// Fixed-width columns, so a name never shifts as the numbers beside it do.
const (
	currentWidth = 2
	trackWidth   = 10
	dateWidth    = 8
)

// list renders the window of branches at offset; cursor indexes the full list.
func list(branches []git.Branch, offset, cursor, width int) string {
	rows := make([]string, len(branches))
	for i, b := range branches {
		rows[i] = row(b, width, offset+i == cursor)
	}
	return strings.Join(rows, "\n")
}

// row renders one branch: bar and name, then marker, tracking and date.
func row(b git.Branch, width int, selected bool) string {
	s := styles.Row
	bar := " "
	if selected {
		s = styles.RowSelected
		bar = s.Accent.Render("▌")
	}

	left := bar + s.Row.Render(" ") + components.StyledPath(b.Name, s)
	right := components.PadLeft(currentMark(b, s), currentWidth, s.Row) +
		components.PadLeft(trackState(b, s), trackWidth, s.Row) +
		components.PadLeft(s.Dir.Render(components.Relative(b.Updated)), dateWidth, s.Row)

	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	row := left + s.Row.Render(strings.Repeat(" ", gap)) + right

	if pad := width - lipgloss.Width(row); pad > 0 {
		row += s.Row.Render(strings.Repeat(" ", pad))
	}

	return row
}

// currentMark flags the checked-out branch, the one accent on this screen.
func currentMark(b git.Branch, s styles.RowStyles) string {
	if !b.Current {
		return ""
	}
	return s.Accent.Render("●")
}

// trackState summarizes the branch against its upstream, or why it can't.
func trackState(b git.Branch, s styles.RowStyles) string {
	switch {
	case !b.Local:
		return s.Dir.Render("remote")
	case b.Upstream == "":
		return s.Dir.Render("local")
	case b.Ahead == 0 && b.Behind == 0:
		return ""
	}

	var parts []string
	if b.Ahead > 0 {
		parts = append(parts, s.Dir.Render("↑")+s.Name.Render(strconv.Itoa(b.Ahead)))
	}
	if b.Behind > 0 {
		parts = append(parts, s.Dir.Render("↓")+s.Name.Render(strconv.Itoa(b.Behind)))
	}

	return strings.Join(parts, s.Row.Render(" "))
}
