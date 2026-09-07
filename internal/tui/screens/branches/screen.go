// Package branches lists local and remote branches; enter checks one out.
package branches

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SwitchedMsg tells the router to reopen the diff view on the new branch.
type SwitchedMsg struct {
	Path string
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
		case key.Matches(msg, keys.Switch):
			return m.switchTo()
		case key.Matches(msg, keys.Refresh):
			return m.refresh(), nil
		case key.Matches(msg, keys.Help):
			return m.toggleHelp(), nil
		}
	}

	return m, nil
}

func (m Model) HeaderContent() (title string, segments []string) {
	count := strconv.Itoa(len(m.branches)) + " branches"
	if len(m.branches) == 1 {
		count = "1 branch"
	}
	return "Branches", []string{styles.Muted.Render(count)}
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.body(), m.footer())
}

// body renders the list, or the load error in its place.
func (m Model) body() string {
	height := m.bodyHeight()

	if m.loadErr != nil {
		return lipgloss.NewStyle().Width(m.width).Height(height).Render(
			lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center,
				styles.Error.Render("error: "+m.loadErr.Error())),
		)
	}

	rows := m.listRows()
	end := min(m.listOffset+rows, len(m.branches))
	window := m.branches[min(m.listOffset, len(m.branches)):end]

	return lipgloss.NewStyle().Width(m.width).Height(height).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.notice(),
			list(window, m.listOffset, m.cursor, m.width),
		),
	)
}

// notice is the row above the list: the last refusal, or blank to hold it.
// Rendered at the body's width, so a long one wraps here and not in the join.
func (m Model) notice() string {
	if m.switchErr == nil {
		return ""
	}
	return lipgloss.NewStyle().Width(m.width).Render(" " + styles.Error.Render(m.switchErr.Error()))
}

// headerHeight measures the notice, since git's errors wrap onto a second row.
func (m Model) headerHeight() int {
	return lipgloss.Height(m.notice())
}

// footer renders the bottom help bar.
func (m Model) footer() string {
	return components.Footer(m.width, m.help, keys)
}

// footerHeight measures the footer, since the help bar's height changes.
func (m Model) footerHeight() int {
	return lipgloss.Height(m.footer())
}
