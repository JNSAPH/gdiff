// Package worktrees lists the repository's git worktrees, each with its own
// uncommitted-change summary, so a dirty one left behind by an agent stands
// out. Enter switches the diff view to the selected one.
package worktrees

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SelectMsg is sent when the user picks a worktree to switch the diff view
// to. The router reopens diffView on its path.
type SelectMsg struct {
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
		case key.Matches(msg, keys.Open):
			if e, ok := m.selected(); ok {
				return m, func() tea.Msg { return SelectMsg{Path: e.Path} }
			}
		case key.Matches(msg, keys.Refresh):
			return m.refresh(), nil
		case key.Matches(msg, keys.Help):
			return m.toggleHelp(), nil
		}
	}

	return m, nil
}

func (m Model) HeaderContent() (title string, segments []string) {
	count := strconv.Itoa(len(m.entries)) + " worktrees"
	if len(m.entries) == 1 {
		count = "1 worktree"
	}
	return "Worktrees", []string{styles.Muted.Render(count)}
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.body(), m.footer())
}

// body renders everything above the footer: the list, or the load error in
// its place.
func (m Model) body() string {
	height := m.bodyHeight()

	if m.loadErr != nil {
		return lipgloss.NewStyle().Width(m.width).Height(height).Render(
			lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center,
				styles.Error.Render("error: "+m.loadErr.Error())),
		)
	}

	rows := m.listRows()
	end := min(m.listOffset+rows, len(m.entries))
	window := m.entries[min(m.listOffset, len(m.entries)):end]

	return lipgloss.NewStyle().Width(m.width).Height(height).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			components.Rule(m.width, styles.Muted.Render(" worktrees "), styles.BorderLine),
			"",
			list(window, m.listOffset, m.cursor, m.width),
		),
	)
}

// footer renders the bottom help bar.
func (m Model) footer() string {
	return components.Footer(m.width, m.help, keys)
}

// footerHeight measures the footer by rendering it, so the layout can't
// drift from the real help bar the way a hardcoded constant would.
func (m Model) footerHeight() int {
	return lipgloss.Height(m.footer())
}
