package global

import "charm.land/bubbles/v2/key"

var Quit = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("q", "quit"),
)

var OpenCommandBar = key.NewBinding(
	key.WithKeys(":"),
	key.WithHelp(":", "command bar"),
)
