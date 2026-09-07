package styles

import "charm.land/lipgloss/v2"

// Backgrounds paired with their text colors, so green sits on the green band.
const (
	addedLineFgHex   = "#7EE787"
	addedLineBgHex   = "#10261A"
	deletedLineFgHex = "#FFA198"
	deletedLineBgHex = "#2C1416"
)

// Diff line styles color a whole row, so the caller pads the text to the pane
// width first — see diffview's renderDiff.
var (
	AddedLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(addedLineFgHex)).
			Background(lipgloss.Color(addedLineBgHex))

	DeletedLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(deletedLineFgHex)).
			Background(lipgloss.Color(deletedLineBgHex))

	ContextLine = lipgloss.NewStyle().Foreground(TextColor)
)

// Diff gutter styles: the line-number columns and the +/- sign beside them.
var (
	GutterNum     = lipgloss.NewStyle().Foreground(SubtleColor)
	GutterSep     = lipgloss.NewStyle().Foreground(BorderColor)
	GutterAdded   = lipgloss.NewStyle().Foreground(AddedColor).Bold(true)
	GutterDeleted = lipgloss.NewStyle().Foreground(DeletedColor).Bold(true)
)

// RowStyles keeps the selected row's background on every part of the row.
type RowStyles struct {
	Row    lipgloss.Style // the background band, and the padding after the text
	Accent lipgloss.Style // the bar down the row's left edge
	Glyph  lipgloss.Style // the change-type marker; caller sets its foreground
	Dir    lipgloss.Style // the path's directory part, dimmed
	Name   lipgloss.Style // the path's final segment, the part worth reading
}

// RowHeader draws a path outside the sidebar, with no selection band behind it.
var RowHeader = RowStyles{
	Row:   lipgloss.NewStyle(),
	Dir:   lipgloss.NewStyle().Foreground(MutedColor),
	Name:  lipgloss.NewStyle().Foreground(HeadlineColor).Bold(true),
	Glyph: lipgloss.NewStyle().Bold(true),
}

// Row and RowSelected are the two states a sidebar row is drawn in.
var (
	Row = RowStyles{
		Row:    lipgloss.NewStyle(),
		Accent: lipgloss.NewStyle(),
		Glyph:  lipgloss.NewStyle().Bold(true),
		Dir:    lipgloss.NewStyle().Foreground(MutedColor),
		Name:   lipgloss.NewStyle().Foreground(TextColor),
	}

	RowSelected = RowStyles{
		Row:    lipgloss.NewStyle().Background(SurfaceColor),
		Accent: lipgloss.NewStyle().Foreground(BrandColor).Background(SurfaceColor),
		Glyph:  lipgloss.NewStyle().Background(SurfaceColor).Bold(true),
		Dir:    lipgloss.NewStyle().Foreground(MutedColor).Background(SurfaceColor),
		Name:   lipgloss.NewStyle().Foreground(HeadlineColor).Background(SurfaceColor).Bold(true),
	}
)
