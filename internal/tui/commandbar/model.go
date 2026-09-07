package commandbar

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// Model holds the command bar's state. Its size is fixed (see
// styles.CommandBarWidth) and the router positions it, so it tracks neither.
type Model struct {
	input textinput.Model
}

// New builds the command bar, closed, with its input pre-loaded with the
// command names as suggestions.
func New() Model {
	input := textinput.New()
	input.Prompt = ": "
	input.Placeholder = "type a command"
	input.ShowSuggestions = true
	input.SetSuggestions(Commands)
	input.SetStyles(styles.TextInput())
	input.SetWidth(styles.CommandBarInnerWidth)

	return Model{input: input}
}

func (m Model) close() (Model, tea.Cmd) {
	return m, func() tea.Msg { return CloseMsg{} }
}

func (m Model) submit() (Model, tea.Cmd) {
	value := m.input.Value()
	return m, func() tea.Msg { return SubmitMsg{Value: value} }
}
