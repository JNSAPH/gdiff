package splash

import "time"

// Model holds the splash screen's state: the size, and how much wait elapsed.
type Model struct {
	width, height int
	elapsed       time.Duration
}

// New builds the splash screen; it sizes itself from the first WindowSizeMsg.
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
