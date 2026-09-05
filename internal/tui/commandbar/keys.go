package commandbar

import "charm.land/bubbles/v2/key"

var (
	submit = key.NewBinding(key.WithKeys("enter"))
	cancel = key.NewBinding(key.WithKeys("esc"))
)
