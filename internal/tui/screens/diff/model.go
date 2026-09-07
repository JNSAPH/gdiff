package diffview

import (
	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// resize stores the new size and re-derives everything that depends on it.
func (m Model) resize(width, height int) Model {
	m.width = width
	m.height = height

	m.viewport.SetWidth(m.contentWidth())
	m.viewport.SetHeight(max(0, m.bodyHeight()-contentHeaderHeight-m.tabsHeight()))
	m.filter.SetWidth(m.filterWidth())

	// A resize re-renders the padded rows — but doesn't re-read them.
	return m.clampListOffset().renderDiff()
}

// bodyHeight is what's left for the panes once the footer has taken its rows.
func (m Model) bodyHeight() int {
	return max(0, m.height-m.footerHeight())
}

// filterWidth is the room the input has, less the column its prompt takes.
func (m Model) filterWidth() int {
	if m.narrow() {
		return max(0, m.width-1)
	}
	return max(0, m.sidebarW()-components.SidebarBorderWidth-1)
}

// listRows is how many file rows the sidebar can show at once.
func (m Model) listRows() int {
	return max(0, m.bodyHeight()-sidebarHeaderHeight)
}

// selectPrev and selectNext step through the list, up/down or left/right.
func (m Model) selectPrev() Model {
	if m.cursor > 0 {
		m.cursor--
		m.scrollOffset = 0
	}
	return m.clampListOffset().loadDiff()
}

func (m Model) selectNext() Model {
	if m.cursor < len(m.files)-1 {
		m.cursor++
		m.scrollOffset = 0
	}
	return m.clampListOffset().loadDiff()
}

// refreshList re-derives the display list; route every change through it.
func (m Model) refreshList() Model {
	m = m.applySort().applyFilter()
	m.cursor = max(0, min(m.cursor, len(m.files)-1))

	return m.clampListOffset().loadDiff()
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

	m.listOffset = min(m.listOffset, max(0, len(m.files)-rows))
	m.listOffset = max(0, m.listOffset)

	return m
}

// AdvanceScroll steps the marquee a column. The router calls it on its clock.
func (m Model) AdvanceScroll() Model {
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

// loadDiff reads the selected file's diff; renderDiff only re-draws it.
func (m Model) loadDiff() Model {
	if m.loadErr != nil {
		return m.showMessage(styles.Error.Render("error: " + m.loadErr.Error()))
	}

	file, ok := m.selected()
	if !ok {
		if m.filterActive() && len(m.allFiles) > 0 {
			return m.showMessage(styles.Muted.Render("No files match the filter"))
		}
		return m.showMessage(styles.Muted.Render("No changes"))
	}

	var lines []git.DiffLine
	var err error
	if m.base == baseCheckpoint {
		var ref string
		if ref, err = m.repo.CheckpointRef(); err == nil {
			lines, err = m.repo.FileDiffAgainst(ref, file.Name())
		}
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

// leadingContext keeps a few unchanged lines above the first change.
const leadingContext = 3

// firstChange opens the diff at the first real change, rarely the top.
func firstChange(lines []git.DiffLine) int {
	for i, l := range lines {
		if l.Type != git.LineEqual {
			return max(0, i-leadingContext)
		}
	}
	return 0
}

// showMessage replaces the diff with a centered line, kept so a resize
// can re-center it.
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

// toggleLayout pins the layout that isn't showing. Seeding from narrow() means
// the first press always gives the opposite of what's on screen.
func (m Model) toggleLayout() Model {
	if m.narrow() {
		m.layout = layoutSidebar
	} else {
		m.layout = layoutTabs
	}
	return m.resize(m.width, m.height)
}

// toggleHelp resizes the help bar, so the layout is re-derived.
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

	// Best-effort: catches a real commit made outside gdiff since the last
	// check. Only checkpoint mode needs it, and that surfaces its own error.
	_ = m.repo.EnsureCheckpoint()

	var files []git.FileChange
	var err error
	if m.base == baseCheckpoint {
		var ref string
		if ref, err = m.repo.CheckpointRef(); err == nil {
			files, err = git.ChangedFilesAgainst(m.gitPath, ref)
		}
	} else {
		files, err = m.repo.ChangedFiles()
	}
	if err != nil {
		m.loadErr = err
		return m.loadDiff()
	}

	m.loadErr = nil
	m.allFiles = files
	m.repoName = m.repo.Name()
	m.branch = m.repo.Branch()

	return m.refreshList()
}
