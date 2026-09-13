package commits

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
)

// commitLimit caps the log: enough history to scroll, not so much that opening
// the screen stalls on a large repository.
const commitLimit = 200

// Model holds the commits screen's state.
type Model struct {
	gitPath string

	commits    []git.Commit
	cursor     int
	listOffset int

	loadErr error

	width, height int
	help          help.Model
}

// New lists gitPath's history, newest first.
func New(gitPath string) Model {
	m := Model{gitPath: gitPath, help: components.NewHelp()}
	return m.refresh()
}

// refresh re-reads the log from scratch.
func (m Model) refresh() Model {
	repo, err := git.Open(m.gitPath)
	if err != nil {
		m.loadErr = err
		return m
	}

	commits, err := repo.Commits(commitLimit)
	if err != nil {
		m.loadErr = err
		return m
	}

	m.loadErr = nil
	m.commits = commits
	m.cursor = min(m.cursor, max(0, len(m.commits)-1))

	return m.clampListOffset()
}

// open hands the selected commit to the router, which shows it in the diff view.
func (m Model) open() (Model, tea.Cmd) {
	c, ok := m.selected()
	if !ok {
		return m, nil
	}
	return m, func() tea.Msg { return SelectMsg{Commit: c} }
}

func (m Model) resize(width, height int) Model {
	m.width, m.height = width, height
	return m.clampListOffset()
}

// listRows is what's left for the list once the footer has taken its rows.
func (m Model) listRows() int {
	return max(0, m.height-m.footerHeight())
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
	if m.cursor < len(m.commits)-1 {
		m.cursor++
	}
	return m.clampListOffset()
}

// selected returns the commit the cursor is on, and whether there is one.
func (m Model) selected() (git.Commit, bool) {
	if m.cursor < 0 || m.cursor >= len(m.commits) {
		return git.Commit{}, false
	}
	return m.commits[m.cursor], true
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

	m.listOffset = min(m.listOffset, max(0, len(m.commits)-rows))
	m.listOffset = max(0, m.listOffset)

	return m
}
