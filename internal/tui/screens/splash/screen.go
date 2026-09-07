// Package splash is the startup screen, shown until its timer or a key ends it.
package splash

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JNSAPH/gdiff/internal/core"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

const (
	duration     = 750 * time.Millisecond // how long the screen stays up
	tickInterval = 50 * time.Millisecond  // how often the progress bar redraws
)

// DoneMsg tells the router the splash is finished, by timer or by key.
type DoneMsg struct{}

// tickMsg advances the progress bar by tickInterval.
type tickMsg struct{}

// Init starts the auto-advance timer and the progress bar's tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(duration, func(time.Time) tea.Msg { return DoneMsg{} }),
		tick(),
	)
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// done finishes the screen immediately, skipping the remaining wait.
func done() tea.Cmd {
	return func() tea.Msg { return DoneMsg{} }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height), nil

	case tickMsg:
		return m.advance(tickInterval), tick()

	// Any key skips the splash screen.
	case tea.KeyMsg:
		return m, done()
	}

	return m, nil
}

func (m Model) HeaderContent() (title string, segments []string) {
	return "", nil
}

func (m Model) View() string {
	block := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.BrandTitle.Render("gdiff"),
			" ",
			styles.Subtle.Render("v"+core.Version),
		),
		styles.Muted.Render("by aph"),
		"",
		styles.Subtle.Render("press any key to continue"),
	)

	const barHeight = 1
	body := lipgloss.Place(m.width, m.height-barHeight, lipgloss.Center, lipgloss.Center, block)

	return lipgloss.JoinVertical(lipgloss.Left, body, m.progressBar())
}

// progressBar returns the progress bar string for the splash screen.
func (m Model) progressBar() string {
	if m.width <= 0 {
		return ""
	}

	// Calculate the progress fraction and filled width
	fraction := min(float64(m.elapsed)/float64(duration), 1)
	filled := int(fraction * float64(m.width))

	shades := lipgloss.Blend1D(m.width, styles.BrandDeepColor, styles.BrandColor)

	// Build the progress bar string
	var bar strings.Builder
	for i := 0; i < m.width; i++ {
		if i < filled {
			bar.WriteString(lipgloss.NewStyle().Foreground(shades[i]).Render("━"))
		} else {
			bar.WriteString(styles.BorderLine.Render("─"))
		}
	}

	return bar.String()
}
