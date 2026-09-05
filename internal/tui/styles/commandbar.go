package styles

import "charm.land/lipgloss/v2"

// CommandBarWidth is the popup's total width, border included.
const CommandBarWidth = 55

// CommandBarInnerWidth is what's left for the input: the total less the two
// border columns and the padding on each side.
const CommandBarInnerWidth = CommandBarWidth - 4

// CommandBar is the command popup's box. Like AppBorder it leaves its top
// edge to be drawn by hand, so the "Commands" label can sit in it.
var CommandBar = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderTop(false).
	BorderForeground(FocusedBorderColor).
	Padding(0, 1).
	Width(CommandBarWidth)
