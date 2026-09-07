package styles

import "charm.land/lipgloss/v2"

// CommandBarWidth is the popup's total width, border included.
const CommandBarWidth = 55

// CommandBarInnerWidth is the total less the border columns and padding.
const CommandBarInnerWidth = CommandBarWidth - 4

// CommandBar leaves its top edge undrawn, so a label can sit in it.
var CommandBar = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderTop(false).
	BorderForeground(FocusedBorderColor).
	Padding(0, 1).
	Width(CommandBarWidth)
