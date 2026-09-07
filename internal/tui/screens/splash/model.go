package splash

import "time"

// Model holds the splash screen's state: the terminal size, so the text can
// be centered, and how much of the wait has elapsed, for the progress bar.
type Model struct {
	width, height int
	elapsed       time.Duration
}

// New builds the splash screen. It sizes itself from the first
// tea.WindowSizeMsg, so the zero value is a valid starting state.
func New() Model {
	return Model{}
}

func (m Model) resize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

func (m Model) advance(by time.Duration) Model {
	m.elapsed += by
	return m
}
