// Package styles holds the palette and shared styles, so retheming is one edit.
package styles

import "charm.land/lipgloss/v2"

// Palette, unexported. It assumes a dark terminal: it paints its own
// backgrounds and pairs them with light text.
const (
	brandColorHex     = "#FC5000" // primary accent — focus, emphasis
	brandDeepColorHex = "#B33800" // darker end of the brand gradient
	headlineColorHex  = "#FFFFFF" // titles, headings
	textColorHex      = "#D4D4D4" // regular body text
	mutedColorHex     = "#8A8A8A" // secondary text, dimmed paths
	subtleColorHex    = "#5A5A5A" // gutters, separators — quieter than muted
	borderColorHex    = "#3A3A3A" // structural borders
	surfaceColorHex   = "#262626" // selected row background

	addedColorHex    = "#3FB950" // new files
	deletedColorHex  = "#F85149" // deleted files
	modifiedColorHex = "#E3B341" // modified files
	renamedColorHex  = "#58A6FF" // renamed files
)

// Colors
var (
	BrandColor     = lipgloss.Color(brandColorHex)
	BrandDeepColor = lipgloss.Color(brandDeepColorHex)
	HeadlineColor  = lipgloss.Color(headlineColorHex)
	TextColor      = lipgloss.Color(textColorHex)
	MutedColor     = lipgloss.Color(mutedColorHex)
	SubtleColor    = lipgloss.Color(subtleColorHex)
	BorderColor    = lipgloss.Color(borderColorHex)
	SurfaceColor   = lipgloss.Color(surfaceColorHex)

	AddedColor    = lipgloss.Color(addedColorHex)
	DeletedColor  = lipgloss.Color(deletedColorHex)
	ModifiedColor = lipgloss.Color(modifiedColorHex)
	RenamedColor  = lipgloss.Color(renamedColorHex)

	FocusedBorderColor = BrandColor // border of whatever currently has focus
)

// Text styles
var (
	BrandTitle = lipgloss.NewStyle().Bold(true).Foreground(BrandColor)
	Title      = lipgloss.NewStyle().Bold(true).Foreground(HeadlineColor)
	Text       = lipgloss.NewStyle().Foreground(TextColor)
	Error      = lipgloss.NewStyle().Foreground(DeletedColor)

	// Muted is for text that should recede, e.g. a path's directory.
	Muted = lipgloss.NewStyle().Foreground(MutedColor)

	// Subtle is chrome that should barely register, e.g. gutter line numbers.
	Subtle = lipgloss.NewStyle().Foreground(SubtleColor)

	// BorderLine styles border runes drawn by hand, not a lipgloss border.
	BorderLine = lipgloss.NewStyle().Foreground(BorderColor)

	// FocusedBorderLine is BorderLine for a pane that has focus.
	FocusedBorderLine = lipgloss.NewStyle().Foreground(FocusedBorderColor)
)

// AppBorder leaves its top edge undrawn — the header above it forms that edge.
var AppBorder = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderTop(false).
	BorderForeground(BorderColor)

// What AppBorder's frame costs. Set Width to the total and Height to the
// interior — lipgloss counts the two differently.
const (
	AppBorderWidthOverhead  = 2 // the left and right edges
	AppBorderHeightOverhead = 1 // the bottom edge
)
