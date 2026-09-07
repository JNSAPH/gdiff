package diffview

// diffBase is what the pane compares against: the last commit, or the last
// checkpoint. baseHead is the zero value, so the screen starts there.
type diffBase int

const (
	baseHead diffBase = iota
	baseCheckpoint
)

// label names the base for the header segment.
func (b diffBase) label() string {
	if b == baseCheckpoint {
		return "checkpoint"
	}
	return "HEAD"
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
