package branches

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
)

// Model holds the branches screen's state.
type Model struct {
	gitPath string

	branches   []git.Branch
	cursor     int
	listOffset int

	loadErr   error // failure listing the branches themselves
	switchErr error // why the last switch didn't happen

	width, height int
	help          help.Model
}

// New lists gitPath's local and remote branches.
func New(gitPath string) Model {
	m := Model{gitPath: gitPath, help: components.NewHelp()}
	return m.refresh()
}

// refresh re-lists the branches from scratch.
func (m Model) refresh() Model {
	repo, err := git.Open(m.gitPath)
	if err != nil {
		m.loadErr = err
		return m
	}

	branches, err := repo.Branches()
	if err != nil {
		m.loadErr = err
		return m
	}

	m.loadErr = nil
	m.branches = branches
	m.cursor = min(m.cursor, max(0, len(m.branches)-1))

	return m.clampListOffset()
}

// switchTo checks out the selected branch and reopens the diff view. It
// refuses on a dirty tree: git carries uncommitted changes across.
func (m Model) switchTo() (Model, tea.Cmd) {
	b, ok := m.selected()
	if !ok || b.Current {
		return m, nil
	}

	repo, err := git.Open(m.gitPath)
	if err != nil {
		m.switchErr = err
		return m, nil
	}

	pending, err := git.ChangedFilesAt(m.gitPath)
	if err != nil {
		m.switchErr = err
		return m, nil
	}
	if len(pending) > 0 {
		m.switchErr = fmt.Errorf("%s — commit or stash first", pendingCount(len(pending)))
		return m, nil
	}

	if err := repo.Checkout(b.Name); err != nil {
		m.switchErr = err
		return m, nil
	}

	m.switchErr = nil
	return m, func() tea.Msg { return SwitchedMsg{Path: m.gitPath} }
}

// pendingCount phrases the count of uncommitted changes blocking a switch.
func pendingCount(n int) string {
	if n == 1 {
		return "1 uncommitted change"
	}
	return strconv.Itoa(n) + " uncommitted changes"
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
	return max(0, m.bodyHeight()-m.headerHeight())
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
	if m.cursor < len(m.branches)-1 {
		m.cursor++
	}
	return m.clampListOffset()
}

// selected returns the branch the cursor is on, and whether there is one.
func (m Model) selected() (git.Branch, bool) {
	if m.cursor < 0 || m.cursor >= len(m.branches) {
		return git.Branch{}, false
	}
	return m.branches[m.cursor], true
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

	m.listOffset = min(m.listOffset, max(0, len(m.branches)-rows))
	m.listOffset = max(0, m.listOffset)

	return m
}
