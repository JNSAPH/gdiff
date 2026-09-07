package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/tui/commandbar"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/screens/branches"
	diffview "github.com/JNSAPH/gdiff/internal/tui/screens/diff"
	"github.com/JNSAPH/gdiff/internal/tui/screens/splash"
	"github.com/JNSAPH/gdiff/internal/tui/screens/worktrees"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// screen identifies which top-level screen is active.
type screen int

const (
	screenSplash screen = iota
	screenDiffView
	screenWorktrees
	screenBranches
)

// Model is the app's root state; everything else routes to a screen.
type Model struct {
	showCommandBar bool
	width, height  int

	active screen

	// Screens
	diffView   diffview.Model
	splash     splash.Model
	worktrees  worktrees.Model
	branches   branches.Model
	commandBar commandbar.Model
}

// NewModel builds every screen up front; they stay unsized until the first
// tea.WindowSizeMsg reaches resize.
func NewModel(gitPath string) Model {
	return Model{
		diffView:   diffview.New(gitPath),
		splash:     splash.New(),
		worktrees:  worktrees.New(gitPath),
		branches:   branches.New(gitPath),
		commandBar: commandbar.New(),
	}
}

func (m Model) quitApp() (Model, tea.Cmd) {
	return m, tea.Quit
}

// openCommandBar shows the popup and hands it focus.
func (m Model) openCommandBar() (Model, tea.Cmd) {
	m.showCommandBar = true

	var cmd tea.Cmd
	m.commandBar, cmd = m.commandBar.Open()
	return m, cmd
}

func (m Model) closeCommandBar() (Model, tea.Cmd) {
	m.showCommandBar = false
	return m, nil
}

// marqueeTickMsg steps the diff view's filename scroll.
type marqueeTickMsg struct{}

// marqueeTick schedules the next marquee step. The router owns the timer so
// exactly one chain runs: a screen re-arming its own loses it when a message
// is routed elsewhere, and starts a second on every re-entry.
func marqueeTick() tea.Cmd {
	return tea.Tick(diffview.ScrollTickInterval, func(time.Time) tea.Msg { return marqueeTickMsg{} })
}

// advanceMarquee steps the scroll and re-arms unconditionally, whatever screen
// is showing, so the chain can't die.
func (m Model) advanceMarquee() (Model, tea.Cmd) {
	m.diffView = m.diffView.AdvanceScroll()
	return m, marqueeTick()
}

// capturesInput reports whether the active screen is taking typed text.
func (m Model) capturesInput() bool {
	return m.active == screenDiffView && m.diffView.CapturesInput()
}

// showDiffView switches to the diff view and starts its commands.
func (m Model) showDiffView() (Model, tea.Cmd) {
	m.active = screenDiffView
	return m, m.diffView.Init()
}

// showWorktrees switches to the worktrees screen. Already there is a no-op.
func (m Model) showWorktrees() (Model, tea.Cmd) {
	if m.active == screenWorktrees {
		return m, nil
	}

	m.active = screenWorktrees
	return m, m.worktrees.Init()
}

// showBranches switches to the branches screen. Already there is a no-op.
func (m Model) showBranches() (Model, tea.Cmd) {
	if m.active == screenBranches {
		return m, nil
	}

	m.active = screenBranches
	return m, m.branches.Init()
}

// openDiffAt reopens the diff view on path — its files may have all changed.
func (m Model) openDiffAt(path string) (Model, tea.Cmd) {
	m.diffView = diffview.New(path)
	m.branches = branches.New(path)
	m.worktrees = m.worktrees.SetActive(path)

	// The new diffView is unsized until the router hands it space.
	m, resizeCmd := m.resize(m.width, m.height)
	m, showCmd := m.showDiffView()

	return m, tea.Batch(resizeCmd, showCmd)
}

// resize stores the terminal size and passes each screen what it actually
// gets: the terminal less the app border and header.
func (m Model) resize(width, height int) (Model, tea.Cmd) {
	m.width, m.height = width, height

	inner := tea.WindowSizeMsg{
		Width:  max(0, width-styles.AppBorderWidthOverhead),
		Height: max(0, height-styles.AppBorderHeightOverhead-components.HeaderHeight),
	}

	// The command bar is a fixed-size popup the router places, so it gets none.
	var cmd tea.Cmd
	m.splash, _ = m.splash.Update(inner)
	m.worktrees, _ = m.worktrees.Update(inner)
	m.branches, _ = m.branches.Update(inner)
	m.diffView, cmd = m.diffView.Update(inner)
	return m, cmd
}

// updateActiveScreen forwards msg to whichever screen is active.
func (m Model) updateActiveScreen(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.active {
	case screenSplash:
		m.splash, cmd = m.splash.Update(msg)
	case screenWorktrees:
		m.worktrees, cmd = m.worktrees.Update(msg)
	case screenBranches:
		m.branches, cmd = m.branches.Update(msg)
	default:
		m.diffView, cmd = m.diffView.Update(msg)
	}
	return m, cmd
}

// runCommand handles a command submitted in the command bar.
func (m Model) runCommand(value string) (Model, tea.Cmd) {
	m.showCommandBar = false

	switch value {
	case commandbar.CommandQuit:
		return m.quitApp()
	case commandbar.CommandWorktrees:
		return m.showWorktrees()
	case commandbar.CommandBranches:
		return m.showBranches()
	case commandbar.CommandDiff:
		return m.showDiffView()
	case commandbar.CommandAccept:
		return m.acceptCheckpoint(false)
	case commandbar.CommandAcceptAll:
		return m.acceptCheckpoint(true)
	case commandbar.CommandReject:
		return m.rejectCheckpoint(false)
	case commandbar.CommandRejectAll:
		return m.rejectCheckpoint(true)
	}

	return m, nil
}

// acceptCheckpoint and rejectCheckpoint only apply to the diff you're looking
// at — no way to accept or discard changes you haven't seen.

func (m Model) acceptCheckpoint(all bool) (Model, tea.Cmd) {
	if m.active != screenDiffView {
		return m, nil
	}
	if all {
		m.diffView = m.diffView.AcceptCheckpoint()
	} else {
		m.diffView = m.diffView.AcceptSelectedFile()
	}
	return m, nil
}

func (m Model) rejectCheckpoint(all bool) (Model, tea.Cmd) {
	if m.active != screenDiffView {
		return m, nil
	}
	if all {
		m.diffView = m.diffView.RejectCheckpoint()
	} else {
		m.diffView = m.diffView.RejectSelectedFile()
	}
	return m, nil
}
