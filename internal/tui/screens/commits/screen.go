// Package commits lists the repository's history; enter opens one commit's
// changes in the diff view.
package commits

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SelectMsg tells the router to show this commit's changes in the diff view.
type SelectMsg struct {
	Commit git.Commit
}

type refreshMsg struct{}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return refreshMsg{} }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height), nil
	case refreshMsg:
		return m.refresh(), nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			return m.moveCursorUp(), nil
		case key.Matches(msg, keys.Down):
			return m.moveCursorDown(), nil
		case key.Matches(msg, keys.Open):
			return m.open()
		case key.Matches(msg, keys.Refresh):
			return m.refresh(), nil
		case key.Matches(msg, keys.Help):
			return m.toggleHelp(), nil
		}
	}

	return m, nil
}

func (m Model) HeaderContent() (title string, segments []string) {
	count := strconv.Itoa(len(m.commits)) + " commits"
	if len(m.commits) == 1 {
		count = "1 commit"
	}
	return "Commits", []string{styles.Muted.Render(count)}
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.body(), m.footer())
}

// body renders the list, or the load error in its place.
func (m Model) body() string {
	height := m.listRows()

	if m.loadErr != nil {
		return lipgloss.NewStyle().Width(m.width).Height(height).Render(
			lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center,
				styles.Error.Render("error: "+m.loadErr.Error())),
		)
	}

	end := min(m.listOffset+height, len(m.commits))
	window := m.commits[min(m.listOffset, len(m.commits)):end]

	return lipgloss.NewStyle().Width(m.width).Height(height).Render(
		list(window, m.listOffset, m.cursor, m.width),
	)
}

// footer renders the bottom help bar.
func (m Model) footer() string {
	return components.Footer(m.width, m.help, keys)
}

// footerHeight measures the footer, since the help bar's height changes.
func (m Model) footerHeight() int {
	return lipgloss.Height(m.footer())
}
