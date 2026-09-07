package global

import "charm.land/bubbles/v2/key"

// Quit closes the app from any screen.
var Quit = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("q", "quit"),
)

// OpenCommandBar opens the ":" popup from any screen.
var OpenCommandBar = key.NewBinding(
	key.WithKeys(":"),
	key.WithHelp(":", "command bar"),
)
