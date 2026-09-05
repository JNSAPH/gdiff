package diffview

// diffBase identifies which ref the diff pane compares the worktree
// against: the last accepted checkpoint, or the last real commit.
// baseCheckpoint is the zero value — the screen starts there, since that's
// the view accept/reject actually affect. A fresh checkpoint starts out
// equal to HEAD, so this changes nothing about what's shown until the first
// accept.
type diffBase int

const (
	baseCheckpoint diffBase = iota
	baseHead
)

// label names the base for the header segment.
func (b diffBase) label() string {
	if b == baseCheckpoint {
		return "checkpoint"
	}
	return "HEAD"
}

// helpKeys is the keymap for the footer's help bar: Accept/AcceptAll/Reject
// only shown while they'd actually do something, and the short line
// narrowed to checkpoint-review keys rather than advertising ones that are
// no-ops here.
func (m Model) helpKeys() keyMap {
	k := keys
	inCheckpoint := m.base == baseCheckpoint
	k.Accept.SetEnabled(inCheckpoint)
	k.AcceptAll.SetEnabled(inCheckpoint)
	k.Reject.SetEnabled(inCheckpoint)
	k.checkpointFocus = inCheckpoint
	return k
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

// AcceptSelectedFile moves just the selected file's current content into
// the checkpoint, leaving every other pending file exactly as the
// checkpoint already had it. A no-op outside checkpoint mode or with
// nothing selected.
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

// RejectSelectedFile restores just the selected file to the checkpoint,
// discarding its own pending change. Destructive, and — like
// RejectCheckpoint — a no-op outside checkpoint mode.
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

// AcceptCheckpoint moves the checkpoint ref to match the current working
// tree — everything since the last checkpoint becomes the new baseline —
// and reloads. A no-op outside checkpoint mode: accepting only makes sense
// for the diff you're actually looking at.
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

// RejectCheckpoint resets the working tree back to the last checkpoint,
// discarding everything since, and reloads. Destructive, and — like
// AcceptCheckpoint — a no-op outside checkpoint mode.
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
