package styles

import "charm.land/lipgloss/v2"

// Splash screen progress bar
var (
	ProgressFilled = lipgloss.NewStyle().Foreground(FocusedBorderColor)
	ProgressEmpty  = lipgloss.NewStyle().Foreground(BorderColor)
)
