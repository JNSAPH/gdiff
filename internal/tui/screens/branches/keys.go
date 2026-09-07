package branches

import (
	"charm.land/bubbles/v2/key"

	"github.com/JNSAPH/gdiff/internal/tui/global"
)

// keyMap is the screen's bindings, rendered straight into the help bar.
type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	Switch         key.Binding
	Refresh        key.Binding
	Help           key.Binding
	OpenCommandBar key.Binding
	Quit           key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Switch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "switch"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),

	// Handled by the router, listed here so they show up in the help bar.
	OpenCommandBar: global.OpenCommandBar,
	Quit:           global.Quit,
}

// ShortHelp returns the bindings shown in the collapsed, one-line help.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Switch, k.Refresh, k.Help, k.Quit}
}

// FullHelp returns the bindings shown in the expanded, multi-column help.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Switch, k.Refresh},
		{k.Help, k.OpenCommandBar, k.Quit},
	}
}
