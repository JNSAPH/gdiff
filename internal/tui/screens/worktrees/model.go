package worktrees

import (
	"charm.land/bubbles/v2/help"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
)

// entry pairs a worktree with its own changes — each has an independent index.
type entry struct {
	git.Worktree
	counts  map[git.ChangeType]int
	loadErr error
}

// headerHeight is how many rows sit above the list: the title rule.
const headerHeight = 2

// Model holds the worktrees screen's state.
type Model struct {
	gitPath string

	// active is the worktree the diff view is on, resolved to the root git
	// lists it under — the path the router sets can be any directory inside it.
	active string

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
	return m.SetActive(gitPath).refresh()
}

// SetActive points the marker at the worktree the diff view just opened.
func (m Model) SetActive(path string) Model {
	m.active = ""

	repo, err := git.Open(path)
	if err != nil {
		return m
	}
	if root, err := repo.Root(); err == nil {
		m.active = root
	}

	return m
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
	return m.clampListOffset()
}

// bodyHeight is what's left for the list once the footer has taken its rows.
func (m Model) bodyHeight() int {
	return max(0, m.height-m.footerHeight())
}

func (m Model) listRows() int {
	return max(0, m.bodyHeight()-headerHeight)
}

// toggleHelp resizes the help bar, so the list has to be re-clamped.
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

// clampListOffset scrolls the list only as far as the cursor needs.
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
