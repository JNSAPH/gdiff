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

func New() Model {
	input := textinput.New()
	input.Prompt = ": "
	input.Placeholder = "type a command"
	input.ShowSuggestions = true
	input.SetSuggestions(Commands)
	input.SetStyles(inputStyles())
	input.SetWidth(styles.CommandBarInnerWidth)

	return Model{input: input}
}

// inputStyles paints the text input in the app's palette, dimming it while
// it doesn't have focus.
func inputStyles() textinput.Styles {
	s := textinput.DefaultDarkStyles()

	s.Focused.Prompt = styles.BrandTitle
	s.Focused.Text = styles.Text
	s.Focused.Placeholder = styles.Subtle
	s.Focused.Suggestion = styles.Subtle

	s.Blurred.Prompt = styles.Subtle
	s.Blurred.Text = styles.Muted
	s.Blurred.Placeholder = styles.Subtle
	s.Blurred.Suggestion = styles.Subtle

	s.Cursor.Color = styles.BrandColor

	return s
}

func (m Model) close() (Model, tea.Cmd) {
	return m, func() tea.Msg { return CloseMsg{} }
}

func (m Model) submit() (Model, tea.Cmd) {
	value := m.input.Value()
	return m, func() tea.Msg { return SubmitMsg{Value: value} }
}
