package diffview

import (
	"charm.land/bubbles/v2/key"

	"github.com/JNSAPH/gdiff/internal/tui/global"
)

// keyMap is the screen's bindings. It implements help.KeyMap via ShortHelp
// and FullHelp, so the footer's help bar renders straight from it.
type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	Open            key.Binding
	Back            key.Binding
	Tab             key.Binding
	ToggleLayout    key.Binding
	SortMode        key.Binding
	SortModeReverse key.Binding
	Filter          key.Binding
	FilterApply     key.Binding
	FilterCancel    key.Binding
	ToggleBase      key.Binding
	Accept          key.Binding
	AcceptAll       key.Binding
	Reject          key.Binding
	Help            key.Binding
	OpenCommandBar  key.Binding
	Refresh         key.Binding
	Quit            key.Binding

	// filterFocus narrows ShortHelp to the two keys that end the filter.
	filterFocus bool

	// checkpointFocus trims ShortHelp to the keys relevant to reviewing a
	// checkpoint, dropping ones common enough elsewhere not to need a
	// standing reminder. FullHelp ignores it — '?' still shows everything.
	checkpointFocus bool
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
	// Left/Right replace Up/Down once the file list is a tab strip.
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev file"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next file"),
	),
	Open: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	// Esc alone: in the diff pane h/l and the arrows scroll it sideways.
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch pane"),
	),
	ToggleLayout: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "tabs/sidebar"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	SortMode: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "cycle sort"),
	),
	SortModeReverse: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "cycle sort (reverse)"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	// Matched only while filtering, where enter and esc aren't Open and Back.
	FilterApply: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply filter"),
	),
	FilterCancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
	ToggleBase: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "toggle diff base"),
	),
	Accept: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "accept file"),
	),
	AcceptAll: key.NewBinding(
		key.WithKeys("Y"),
		key.WithHelp("Y", "accept all"),
	),
	Reject: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "reject file"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),

	// Handled by the router, listed here so they show up in the help bar.
	OpenCommandBar: global.OpenCommandBar,
	Quit:           global.Quit,
}

// ShortHelp returns the bindings shown in the collapsed, one-line help. In
// checkpointFocus mode, Tab/SortMode/Refresh give way to the checkpoint
// actions — they're still one "?" away in FullHelp.
func (k keyMap) ShortHelp() []key.Binding {
	if k.filterFocus {
		return []key.Binding{k.FilterApply, k.FilterCancel}
	}
	if k.checkpointFocus {
		return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Accept, k.AcceptAll, k.Reject, k.Help, k.Quit}
	}
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Tab, k.SortMode, k.Help, k.Refresh, k.Quit}
}

// FullHelp returns the bindings shown in the expanded, multi-column help.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Open, k.Back},
		{k.SortMode, k.SortModeReverse, k.Filter, k.Refresh},
		{k.ToggleBase, k.Accept, k.AcceptAll, k.Reject},
		{k.ToggleLayout, k.Help, k.OpenCommandBar, k.Quit},
	}
}

// activeKeys is the keymap as it applies right now. It feeds both the footer
// and Update's dispatch, so a disabled binding neither shows nor fires.
func (m Model) activeKeys() keyMap {
	k := keys
	k.filterFocus = m.filtering

	// The file list is either a sidebar or a tab strip, so only one pair of
	// arrows moves through it at a time.
	narrow := m.narrow()
	k.Up.SetEnabled(!narrow)
	k.Down.SetEnabled(!narrow)
	k.Left.SetEnabled(narrow)
	k.Right.SetEnabled(narrow)

	// Accept/reject only mean something against a checkpoint, and the short
	// help narrows to those keys rather than advertising ones that no-op.
	inCheckpoint := m.base == baseCheckpoint
	k.Accept.SetEnabled(inCheckpoint)
	k.AcceptAll.SetEnabled(inCheckpoint)
	k.Reject.SetEnabled(inCheckpoint)
	k.checkpointFocus = inCheckpoint

	return k
}
