// Package diffview is the main screen: a sidebar listing changed files
// next to a pane showing the selected file's diff.
package diffview

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// ScrollTickInterval is how often the selected row's marquee advances. The
// router drives it, so the interval lives here but the timer does not.
const ScrollTickInterval = 300 * time.Millisecond

// focus identifies which pane has keyboard focus.
type focus int

const (
	focusSidebar focus = iota
	focusContent
)

// Model is the diff screen's state: the repository, the files it found, and
// the diff of whichever one the cursor is on.
type Model struct {
	gitPath string
	repo    *git.Repo

	repoName string
	branch   string

	allFiles []git.FileChange // every changed file git reported
	files    []git.FileChange // display order: sorted, then filtered
	loadErr  error

	cursor       int // index into files
	listOffset   int // first file row the sidebar shows
	scrollOffset int // how far the selected row's marquee has advanced
	sortIndex    int // index into sortOptions
	base         diffBase
	focus        focus
	layout       layoutMode

	// The "/" filter: filtering is focus, the query stays until "/" clears it.
	filter    textinput.Model
	filtering bool

	// The selected file's diff, kept so the pane can be re-rendered on a
	// resize without reading git again. When there's none, message says why.
	diffLines    []git.DiffLine
	lineNumWidth int
	message      string

	width, height int
	viewport      viewport.Model
	help          help.Model
}

// New opens the repository and loads its changed files. A failure goes to
// loadErr and shows in the content pane, so the TUI can still start.
func New(gitPath string) Model {
	m := Model{
		gitPath:  gitPath,
		help:     components.NewHelp(),
		viewport: newViewport(),
		filter:   newFilterInput(),
	}

	repo, err := git.Open(gitPath)
	if err != nil {
		m.loadErr = err
		return m
	}
	m.repo = repo

	return m.refreshGit() // calls EnsureCheckpoint too
}

// newViewport builds the diff pane. Soft wrap is off so a long line scrolls
// sideways instead of wrapping and breaking the row's background band.
func newViewport() viewport.Model {
	vp := viewport.New()
	vp.SoftWrap = false
	vp.FillHeight = true
	vp.MouseWheelEnabled = true
	vp.SetHorizontalStep(4)
	return vp
}

// Init returns the commands the screen needs while it's active. The marquee
// runs off the router's clock, not this screen's, so there are none.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height), nil

	case tea.MouseWheelMsg:
		if m.focus == focusSidebar {
			switch msg.Button {
			case tea.MouseWheelUp:
				return m.selectPrev(), nil
			case tea.MouseWheelDown:
				return m.selectNext(), nil
			case tea.MouseMiddle:
				return m.setFocus(focusContent), nil
			}
		}

	case tea.KeyMsg:
		k := m.activeKeys()

		// The filter takes every key while focused, or typing "n" rejects a file.
		if m.filtering {
			switch {
			case key.Matches(msg, k.FilterApply):
				return m.lockFilter(), nil
			case key.Matches(msg, k.FilterCancel):
				return m.clearFilter(), nil
			}
			return m.updateFilter(msg)
		}

		// Keys that work in either pane
		switch {
		case key.Matches(msg, k.Help):
			return m.toggleHelp(), nil
		case key.Matches(msg, k.Tab):
			return m.toggleFocus(), nil
		case key.Matches(msg, k.ToggleLayout):
			return m.toggleLayout(), nil
		case key.Matches(msg, k.SortMode):
			return m.cycleSort(false), nil
		case key.Matches(msg, k.SortModeReverse):
			return m.cycleSort(true), nil
		case key.Matches(msg, k.Refresh):
			return m.refreshGit(), nil
		case key.Matches(msg, k.Filter):
			return m.toggleFilter()
		case key.Matches(msg, k.ToggleBase):
			return m.toggleBase(), nil
		case key.Matches(msg, k.Accept):
			return m.AcceptSelectedFile(), nil
		case key.Matches(msg, k.AcceptAll):
			return m.AcceptCheckpoint(), nil
		case key.Matches(msg, k.Reject):
			return m.RejectSelectedFile(), nil
		}

		// Sidebar keys
		if m.focus == focusSidebar {
			switch {
			case key.Matches(msg, k.Up), key.Matches(msg, k.Left):
				return m.selectPrev(), nil
			case key.Matches(msg, k.Down), key.Matches(msg, k.Right):
				return m.selectNext(), nil
			case key.Matches(msg, k.Open):
				return m.setFocus(focusContent), nil
			}
			return m, nil
		}

		// Content pane keys
		if key.Matches(msg, k.Back) {
			return m.setFocus(focusSidebar), nil
		}
	}

	// Anything left over, while the content pane has focus, goes to the
	// viewport for its own scrolling.
	if m.focus == focusContent {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// HeaderContent returns the repository's name plus the branch and change
// counts, for the app title bar.
func (m Model) HeaderContent() (title string, segments []string) {
	name := m.repoName
	if name == "" {
		name = "gdiff"
	}

	return name, []string{styles.Muted.Render(m.branch), styles.Subtle.Render(m.base.label()), m.changeCounts()}
}

// changeCounts summarizes the working tree as "+3 ~12 -1", leaving out any
// kind that isn't present.
func (m Model) changeCounts() string {
	counts := map[git.ChangeType]int{}
	for _, f := range m.files {
		counts[f.Type]++
	}
	return components.ChangeCounts(counts)
}

func (m Model) View() string {
	footer := m.footer()
	body := m.bodyHeight()

	// To narrow: Tabs on Top
	if m.narrow() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.tabs(m.width),
			lipgloss.NewStyle().
				Width(m.width).
				Height(max(0, body-m.tabsHeight())).
				Render(m.content()),
			footer,
		)
	}

	// Wide: Sidebar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Width(m.width).
			Height(body).
			Render(lipgloss.JoinHorizontal(lipgloss.Top, m.sidebar(body), m.content())),
		footer,
	)
}

// footer renders the bottom help bar.
func (m Model) footer() string {
	return components.Footer(m.width, m.help, m.activeKeys())
}

// footerHeight measures the footer by rendering it, so the layout can't
// drift from the real help bar the way a hardcoded constant would.
func (m Model) footerHeight() int {
	return lipgloss.Height(m.footer())
}
