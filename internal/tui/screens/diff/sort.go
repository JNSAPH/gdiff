package diffview

import (
	"path"
	"sort"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// sortMode identifies how the file list is ordered.
type sortMode int

const (
	sortByType sortMode = iota
	sortByName
	sortByPath
)

// sortOption is one state the sort key cycles through. Adding a sort state
// is just adding an entry here.
type sortOption struct {
	mode      sortMode
	ascending bool
	label     string
}

var sortOptions = []sortOption{
	{sortByType, true, "Type ▲"},
	{sortByName, true, "Name ▲"},
	{sortByName, false, "Name ▼"},
	{sortByPath, true, "Path "},
}

// cycleSort steps to the next sort state, or the previous one if reverse,
// wrapping at either end.
func (m Model) cycleSort(reverse bool) Model {
	step := 1
	if reverse {
		step = -1
	}

	count := len(sortOptions)
	m.sortIndex = (m.sortIndex + step + count) % count

	return m.refreshList()
}

// applySort reorders m.allFiles in place; applyFilter derives the display
// list from it. The cursor indexes into that order, so the selection moves.
func (m Model) applySort() Model {
	opt := sortOptions[m.sortIndex]

	less := func(i, j int) bool {
		if opt.mode == sortByName {
			return baseName(m.allFiles[i]) < baseName(m.allFiles[j])
		}
		return m.allFiles[i].Type < m.allFiles[j].Type
	}

	sort.SliceStable(m.allFiles, func(i, j int) bool {
		if opt.ascending {
			return less(i, j)
		}
		return less(j, i)
	})

	return m
}

// sortLabel puts the current sort state on the left and the cursor's
// position in the list on the right, e.g. "Type ▲            3/32".
func (m Model) sortLabel(width int) string {
	sort := styles.Subtle.Render("sort ") + styles.Muted.Render(sortOptions[m.sortIndex].label)

	return spread(sort, m.listPosition(), width)
}

// baseName returns just the final segment of a change's path, for
// name-based sorting.
func baseName(c git.FileChange) string {
	return path.Base(c.Name())
}
