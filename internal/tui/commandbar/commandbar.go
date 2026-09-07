// Package commandbar is the ":"-triggered command popup.
package commandbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/tui/components"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// SubmitMsg is sent when the user presses enter with a command typed in.
type SubmitMsg struct {
	Value string
}

// CloseMsg is sent when the user cancels out of the command bar.
type CloseMsg struct{}

// Open resets and focuses the input. The router calls this when the command
// bar is opened.
func (m Model) Open() (Model, tea.Cmd) {
	m.input.Reset()
	return m, m.input.Focus()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, cancel):
			return m.close()
		case key.Matches(msg, submit):
			return m.submit()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	suggestions := m.input.MatchedSuggestions()
	current := m.input.CurrentSuggestionIndex()

	// Builds the list of matched commands, name left / description right,
	// each in its own fixed-width column so every row lines up.
	nameWidth := styles.CommandBarInnerWidth - descColWidth
	rows := make([]string, len(suggestions))
	for i, s := range suggestions {
		style := styles.Text
		if i == current {
			style = styles.BrandTitle
		}

		name := lipgloss.NewStyle().Width(nameWidth).Render(style.Render(s))
		desc := lipgloss.NewStyle().Width(descColWidth).Align(lipgloss.Right).Render(styles.Muted.Render(descriptions[s]))
		rows[i] = name + desc
	}

	body := styles.CommandBar.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.input.View(),
			strings.Join(rows, "\n"),
		),
	)

	// The box draws no top edge, so build that row by hand with the title
	// in it. Its width has to match the body's, hence the measure.
	firstLine, _, _ := strings.Cut(body, "\n")
	label := styles.Title.Render(" Commands ")
	borderStyle := lipgloss.NewStyle().Foreground(styles.FocusedBorderColor)

	top := components.TopBorderWithLabel(
		lipgloss.Width(firstLine), label, lipgloss.Left, lipgloss.RoundedBorder(), borderStyle,
	)

	return lipgloss.JoinVertical(lipgloss.Left, top, body)
}
