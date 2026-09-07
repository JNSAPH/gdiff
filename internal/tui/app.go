// Package tui is the root model: it owns the global chrome and routes messages.
package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/commandbar"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/global"
	"github.com/JNSAPH/gdiff/internal/tui/screens/branches"
	"github.com/JNSAPH/gdiff/internal/tui/screens/splash"
	"github.com/JNSAPH/gdiff/internal/tui/screens/worktrees"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.splash.Init(), marqueeTick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys, unless something else owns the keyboard — "q" quits.
		if !m.showCommandBar && !m.capturesInput() {
			switch {
			case key.Matches(msg, global.Quit):
				return m.quitApp()
			case key.Matches(msg, global.OpenCommandBar):
				return m.openCommandBar()
			}
		}

	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height)

	case marqueeTickMsg:
		return m.advanceMarquee()

	case splash.DoneMsg:
		return m.showDiffView()

	case worktrees.SelectMsg:
		return m.openDiffAt(msg.Path)

	case branches.SwitchedMsg:
		return m.openDiffAt(msg.Path)

	case commandbar.CloseMsg:
		return m.closeCommandBar()

	case commandbar.SubmitMsg:
		return m.runCommand(msg.Value)
	}

	// The bar owns the keyboard while it's open, but only the keyboard: routing
	// everything to it would drop whatever else is in flight.
	if m.showCommandBar {
		var barCmd tea.Cmd
		m.commandBar, barCmd = m.commandBar.Update(msg)

		if _, isKey := msg.(tea.KeyMsg); isKey {
			return m, barCmd
		}

		var screenCmd tea.Cmd
		m, screenCmd = m.updateActiveScreen(msg)
		return m, tea.Batch(barCmd, screenCmd)
	}

	return m.updateActiveScreen(msg)
}

func (m Model) View() tea.View {
	title, segments, body := m.activeScreen()

	// The header sits above this box, so its rows come off the interior height.
	box := styles.AppBorder.
		Width(m.width).
		Height(max(0, m.height-styles.AppBorderHeightOverhead-components.HeaderHeight)).
		Render(body)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		components.Header(m.width, title, segments...),
		box,
	)

	if m.showCommandBar {
		content = m.overlayCommandBar(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

// activeScreen renders the active screen and reports its title bar.
func (m Model) activeScreen() (title string, segments []string, body string) {
	switch m.active {
	case screenSplash:
		title, segments = m.splash.HeaderContent()
		return title, segments, m.splash.View()
	case screenWorktrees:
		title, segments = m.worktrees.HeaderContent()
		return title, segments, m.worktrees.View()
	case screenBranches:
		title, segments = m.branches.HeaderContent()
		return title, segments, m.branches.View()
	default:
		title, segments = m.diffView.HeaderContent()
		return title, segments, m.diffView.View()
	}
}

// overlayCommandBar centers the popup below the header, inside the border.
func (m Model) overlayCommandBar(content string) string {
	box := m.commandBar.View()
	boxW, boxH := lipgloss.Size(box)

	bodyW := m.width - styles.AppBorderWidthOverhead
	bodyH := m.height - components.HeaderHeight - styles.AppBorderHeightOverhead

	x := 1 + max(0, (bodyW-boxW)/2)
	y := components.HeaderHeight + max(0, (bodyH-boxH)/2)

	base := lipgloss.NewLayer(content).Z(0)
	popup := lipgloss.NewLayer(box).X(x).Y(y).Z(1)

	return lipgloss.NewCompositor(base, popup).Render()
}
