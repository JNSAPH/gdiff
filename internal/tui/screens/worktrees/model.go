package worktrees

import (
	"charm.land/bubbles/v2/help"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
)

// entry pairs a worktree with a summary of its own uncommitted changes —
// each worktree has an independent working tree and index.
type entry struct {
	git.Worktree
	counts  map[git.ChangeType]int
	loadErr error
}

// headerHeight is how many rows sit above the list: the title rule.
const headerHeight = 2

// Model holds the worktrees screen's state: the listed worktrees and the
// cursor into them.
type Model struct {
	gitPath string

	entries    []entry
	cursor     int
	listOffset int

	loadErr error // failure listing the worktrees themselves

	width, height int
	help          help.Model
}

// New lists gitPath's worktrees and, for each, its own uncommitted changes.
func New(gitPath string) Model {
	m := Model{gitPath: gitPath, help: components.NewHelp()}
	return m.refresh()
}

// refresh re-lists the worktrees and their change summaries from scratch.
func (m Model) refresh() Model {
	repo, err := git.Open(m.gitPath)
	if err != nil {
		m.loadErr = err
		return m
	}

	worktrees, err := repo.Worktrees()
	if err != nil {
		m.loadErr = err
		return m
	}

	m.loadErr = nil
	m.entries = make([]entry, len(worktrees))
	for i, wt := range worktrees {
		m.entries[i] = loadEntry(wt)
	}
	m.cursor = min(m.cursor, max(0, len(m.entries)-1))

	return m.clampListOffset()
}

// loadEntry summarizes wt's own uncommitted changes.
func loadEntry(wt git.Worktree) entry {
	e := entry{Worktree: wt}

	files, err := git.ChangedFilesAt(wt.Path)
	if err != nil {
		e.loadErr = err
		return e
	}

	e.counts = map[git.ChangeType]int{}
	for _, f := range files {
		e.counts[f.Type]++
	}

	return e
}

func (m Model) resize(width, height int) Model {
	m.width, m.height = width, height
	m.help.SetWidth(width)
	return m.clampListOffset()
}

// bodyHeight is the height left for the list once the footer has taken its
// rows.
func (m Model) bodyHeight() int {
	return max(0, m.height-m.footerHeight())
}

func (m Model) listRows() int {
	return max(0, m.bodyHeight()-headerHeight)
}

// toggleHelp expands or collapses the help bar. That changes the footer's
// height, so the list has to be re-clamped.
func (m Model) toggleHelp() Model {
	m.help.ShowAll = !m.help.ShowAll
	return m.clampListOffset()
}

func (m Model) moveCursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	return m.clampListOffset()
}

func (m Model) moveCursorDown() Model {
	if m.cursor < len(m.entries)-1 {
		m.cursor++
	}
	return m.clampListOffset()
}

// selected returns the entry the cursor is on, and whether there is one.
func (m Model) selected() (entry, bool) {
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return entry{}, false
	}
	return m.entries[m.cursor], true
}

// clampListOffset scrolls the list just far enough to keep the cursor on
// screen, and never past the end of the list. Same logic as diffview's.
func (m Model) clampListOffset() Model {
	rows := m.listRows()
	if rows <= 0 {
		return m
	}

	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+rows {
		m.listOffset = m.cursor - rows + 1
	}

	m.listOffset = min(m.listOffset, max(0, len(m.entries)-rows))
	m.listOffset = max(0, m.listOffset)

	return m
}
