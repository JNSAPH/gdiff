// Package styles holds the app's palette and shared styles. Build on the
// colors here instead of naming a hex value, so retheming means one edit.
package styles

import "charm.land/lipgloss/v2"

// Palette, unexported — the rest of the app uses the colors and styles below.
// It assumes a dark terminal: it paints backgrounds and pairs them with light text.
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

	// Subtle is for chrome that should barely register, e.g. the gutter's
	// line numbers or its separator.
	Subtle = lipgloss.NewStyle().Foreground(SubtleColor)

	// BorderLine styles border and divider runes drawn by hand, i.e. not
	// through a lipgloss border.
	BorderLine = lipgloss.NewStyle().Foreground(BorderColor)

	// FocusedBorderLine is BorderLine for a pane that has focus.
	FocusedBorderLine = lipgloss.NewStyle().Foreground(FocusedBorderColor)
)

// AppBorder wraps the body below a screen's header, leaving its top edge
// undrawn — the header joined above it forms that edge instead.
var AppBorder = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderTop(false).
	BorderForeground(BorderColor)

// What AppBorder's frame costs the content inside it. Set Width to the box's
// total and Height to its interior — lipgloss counts the two differently.
const (
	AppBorderWidthOverhead  = 2 // the left and right edges
	AppBorderHeightOverhead = 1 // the bottom edge
)
