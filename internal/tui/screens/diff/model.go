package diffview

import (
	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// resize stores the new size and re-derives everything that depends on it.
func (m Model) resize(width, height int) Model {
	m.width = width
	m.height = height
	m.help.SetWidth(width)

	m.viewport.SetWidth(m.contentWidth())
	m.viewport.SetHeight(max(0, m.bodyHeight()-contentHeaderHeight))

	// The diff's rows are padded to the pane's width, so a resize has to
	// re-render them — but not re-read them.
	return m.clampListOffset().renderDiff()
}

// bodyHeight is the height left for the sidebar and content pane once the
// footer has taken its rows.
func (m Model) bodyHeight() int {
	return max(0, m.height-m.footerHeight())
}

// listRows is how many file rows the sidebar can show at once.
func (m Model) listRows() int {
	return max(0, m.bodyHeight()-sidebarHeaderHeight)
}

func (m Model) moveCursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m.scrollOffset = 0
	}
	return m.clampListOffset().loadDiff()
}

func (m Model) moveCursorDown() Model {
	if m.cursor < len(m.files)-1 {
		m.cursor++
		m.scrollOffset = 0
	}
	return m.clampListOffset().loadDiff()
}

// clampListOffset scrolls the file list just far enough to keep the cursor
// on screen, and never past the end of the list.
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

	m.listOffset = min(m.listOffset, max(0, len(m.files)-rows))
	m.listOffset = max(0, m.listOffset)

	return m
}

func (m Model) advanceScroll() Model {
	m.scrollOffset++
	return m
}

// selected returns the file the cursor is on, and whether there is one.
func (m Model) selected() (git.FileChange, bool) {
	if m.cursor < 0 || m.cursor >= len(m.files) {
		return git.FileChange{}, false
	}
	return m.files[m.cursor], true
}

// loadDiff reads the selected file's diff and hands it to the viewport.
// This is the step that touches git; renderDiff only re-draws what it read.
func (m Model) loadDiff() Model {
	if m.loadErr != nil {
		return m.showMessage(styles.Error.Render("error: " + m.loadErr.Error()))
	}

	file, ok := m.selected()
	if !ok {
		return m.showMessage(styles.Muted.Render("No changes"))
	}

	var lines []git.DiffLine
	var err error
	if m.base == baseCheckpoint {
		lines, err = m.repo.FileDiffAgainst(git.CheckpointRef, file.Name())
	} else {
		lines, err = m.repo.FileDiff(file.Name())
	}
	if err != nil {
		return m.showMessage(styles.Error.Render("error loading diff: " + err.Error()))
	}

	m.diffLines = lines
	m.lineNumWidth = numWidth(lines)
	m.viewport.SetXOffset(0)
	m.viewport.SetYOffset(firstChange(lines))

	return m.renderDiff()
}

// leadingContext is how many unchanged lines to keep above the first change,
// so it doesn't open flush against the top edge.
const leadingContext = 3

// firstChange is the line to open a diff at: far enough down to show the
// first actual change, since a file's changes are rarely at the top.
func firstChange(lines []git.DiffLine) int {
	for i, l := range lines {
		if l.Type != git.LineEqual {
			return max(0, i-leadingContext)
		}
	}
	return 0
}

// showMessage puts a centered line in the diff pane instead of a diff. It's
// kept on the model so a resize can re-center it.
func (m Model) showMessage(text string) Model {
	m.diffLines = nil
	m.message = text
	return m.renderDiff()
}

func (m Model) setFocus(f focus) Model {
	m.focus = f
	return m
}

func (m Model) toggleFocus() Model {
	if m.focus == focusSidebar {
		return m.setFocus(focusContent)
	}
	return m.setFocus(focusSidebar)
}

// toggleHelp expands or collapses the help bar. That changes the footer's
// height, so the layout has to be re-derived.
func (m Model) toggleHelp() Model {
	m.help.ShowAll = !m.help.ShowAll
	return m.resize(m.width, m.height)
}

func (m Model) refreshGit() Model {
	if m.repo == nil {
		repo, err := git.Open(m.gitPath)
		if err != nil {
			m.loadErr = err
			return m.loadDiff()
		}
		m.repo = repo
	}

	// Best-effort, every refresh: catches a real `git commit` made outside
	// gdiff since we last checked, fast-forwarding the checkpoint to match.
	// HEAD mode works fine even if this fails; only checkpoint mode needs
	// it, and that surfaces its own error at that point.
	_ = m.repo.EnsureCheckpoint()

	var files []git.FileChange
	var err error
	if m.base == baseCheckpoint {
		files, err = git.ChangedFilesAgainst(m.gitPath, git.CheckpointRef)
	} else {
		files, err = m.repo.ChangedFiles()
	}
	if err != nil {
		m.loadErr = err
		return m.loadDiff()
	}

	m.loadErr = nil
	m.files = files
	m.repoName = m.repo.Name()
	m.branch = m.repo.Branch()
	m.cursor = min(m.cursor, max(0, len(m.files)-1))

	return m.applySort().clampListOffset().loadDiff()
}
