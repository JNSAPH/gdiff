package diffview

import (
	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/git"
)

// diffBase is what the pane compares against: the last commit, the last
// checkpoint, or one commit picked from the commits screen. baseHead is the
// zero value, so the screen starts there.
type diffBase int

const (
	baseHead diffBase = iota
	baseCheckpoint
	baseCommit
)

// baseLabel names the base for the header segment; a commit names itself.
func (m Model) baseLabel() string {
	switch m.base {
	case baseCheckpoint:
		return "checkpoint"
	case baseCommit:
		return m.commit.Short
	default:
		return "HEAD"
	}
}

// toggleBase flips between the two bases and reloads against the new one.
func (m Model) toggleBase() Model {
	if m.base == baseCheckpoint {
		m.base = baseHead
	} else {
		m.base = baseCheckpoint
	}
	return m.refreshGit()
}

// viewingCommit reports whether the pane is showing a commit rather than the
// working tree. The base is the whole state, so there's nothing else to keep
// in step with it.
func (m Model) viewingCommit() bool {
	return m.base == baseCommit
}

// OpenCommit points the pane at one commit's changes, its first parent as the
// old side. The cursor resets because the file list is a different set.
func (m Model) OpenCommit(c git.Commit) Model {
	m.base = baseCommit
	m.commit = c
	m.cursor = 0
	m.listOffset = 0
	return m.refreshGit()
}

// closeCommit returns to the working tree and asks the router for the commits
// screen. It resets the base because esc is the only way out of commit mode —
// b is disabled while a commit shows — so leaving has to actually leave.
func (m Model) closeCommit() (Model, tea.Cmd) {
	if !m.viewingCommit() {
		return m, nil
	}

	m.base = baseHead
	m.commit = git.Commit{}

	return m.refreshGit(), func() tea.Msg { return CloseCommitMsg{} }
}

// AcceptSelectedFile moves one file into the checkpoint, leaving every other
// pending file as it was. A no-op outside checkpoint mode.
func (m Model) AcceptSelectedFile() Model {
	if m.base != baseCheckpoint || m.repo == nil {
		return m
	}
	file, ok := m.selected()
	if !ok {
		return m
	}
	if err := m.repo.AcceptCheckpointFile(file); err != nil {
		m.loadErr = err
	}
	return m.refreshGit()
}

// RejectSelectedFile restores one file to the checkpoint. Destructive, and a
// no-op outside checkpoint mode.
func (m Model) RejectSelectedFile() Model {
	if m.base != baseCheckpoint || m.repo == nil {
		return m
	}
	file, ok := m.selected()
	if !ok {
		return m
	}
	if err := m.repo.RejectCheckpointFile(file); err != nil {
		m.loadErr = err
	}
	return m.refreshGit()
}

// AcceptCheckpoint makes the working tree the new baseline. A no-op outside
// checkpoint mode: accepting only means something for the diff you're seeing.
func (m Model) AcceptCheckpoint() Model {
	if m.base != baseCheckpoint {
		return m
	}
	if m.repo != nil {
		if err := m.repo.AcceptCheckpoint(); err != nil {
			m.loadErr = err
		}
	}
	return m.refreshGit()
}

// RejectCheckpoint resets the tree to the last checkpoint. Destructive, and a
// no-op outside checkpoint mode.
func (m Model) RejectCheckpoint() Model {
	if m.base != baseCheckpoint {
		return m
	}
	if m.repo != nil {
		if err := m.repo.RejectCheckpoint(); err != nil {
			m.loadErr = err
		}
	}
	return m.refreshGit()
}
