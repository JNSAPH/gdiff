package components

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// ChangeTypeColor is shared, so a change type looks the same wherever it shows.
func ChangeTypeColor(t git.ChangeType) color.Color {
	switch t {
	case git.ChangeTypeNew:
		return styles.AddedColor
	case git.ChangeTypeDeleted:
		return styles.DeletedColor
	case git.ChangeTypeRenamed:
		return styles.RenamedColor
	default:
		return styles.ModifiedColor
	}
}

// ChangeCounts is "+3 ~12 -1", leaving out any type that isn't present.
func ChangeCounts(counts map[git.ChangeType]int) string {
	out := ""
	for _, t := range []git.ChangeType{git.ChangeTypeNew, git.ChangeTypeModified, git.ChangeTypeDeleted, git.ChangeTypeRenamed} {
		if counts[t] == 0 {
			continue
		}
		if out != "" {
			out += " "
		}
		out += lipgloss.NewStyle().
			Foreground(ChangeTypeColor(t)).
			Render(t.Symbol() + strconv.Itoa(counts[t]))
	}
	return out
}
