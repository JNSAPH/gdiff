package components

import (
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// PadLeft right-aligns s in width columns, measured so styling doesn't skew it.
func PadLeft(s string, width int, fill lipgloss.Style) string {
	if pad := width - lipgloss.Width(s); pad > 0 {
		return fill.Render(strings.Repeat(" ", pad)) + s
	}
	return s
}

// Relative renders a time as "5m", "3h", "2d" or "1w" ago.
func Relative(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + "m"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d.Hours())) + "h"
	case d < 7*24*time.Hour:
		return strconv.Itoa(int(d.Hours()/24)) + "d"
	default:
		return strconv.Itoa(int(d.Hours()/24/7)) + "w"
	}
}
