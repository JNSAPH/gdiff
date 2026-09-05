// Package tui is the app's root model. It owns the global chrome — header,
// border, command bar — and routes messages to the active screen.
package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/commandbar"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/global"
	"github.com/JNSAPH/gdiff/internal/tui/screens/splash"
	"github.com/JNSAPH/gdiff/internal/tui/screens/worktrees"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

func (m Model) Init() tea.Cmd {
	return m.splash.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys, but only while the command bar has no focus.
		if !m.showCommandBar {
			switch {
			case key.Matches(msg, global.Quit):
				return m.quitApp()
			case key.Matches(msg, global.OpenCommandBar):
				return m.openCommandBar()
			}
		}

	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height)

	case splash.DoneMsg:
		return m.showDiffView()

	case worktrees.SelectMsg:
		return m.openWorktree(msg.Path)

	case commandbar.CloseMsg:
		return m.closeCommandBar()

	case commandbar.SubmitMsg:
		return m.runCommand(msg.Value)
	}

	// While it's open, the command bar takes everything else.
	if m.showCommandBar {
		var cmd tea.Cmd
		m.commandBar, cmd = m.commandBar.Update(msg)
		return m, cmd
	}

	return m.updateActiveScreen(msg)
}

func (m Model) View() tea.View {
	title, segments, body := m.activeScreen()

	// Render the body of the active screen (inkl. border)
	box := styles.AppBorder.
		Width(m.width).
		Height(max(0, m.height-styles.AppBorderHeightOverhead-components.HeaderHeight)).
		Render(body)

	// Combine Header and Body into a single view
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

// activeScreen renders the active screen and reports what its title bar
// should say.
func (m Model) activeScreen() (title string, segments []string, body string) {
	switch m.active {
	case screenSplash:
		title, segments = m.splash.HeaderContent()
		return title, segments, m.splash.View()
	case screenWorktrees:
		title, segments = m.worktrees.HeaderContent()
		return title, segments, m.worktrees.View()
	default:
		title, segments = m.diffView.HeaderContent()
		return title, segments, m.diffView.View()
	}
}

// overlayCommandBar centers the command bar popup in the bordered body,
// which starts below the header and one column in from the border.
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
