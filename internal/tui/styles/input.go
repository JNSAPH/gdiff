package styles

import "charm.land/bubbles/v2/textinput"

// TextInput paints an input in the app's palette, dimmed while unfocused.
func TextInput() textinput.Styles {
	s := textinput.DefaultDarkStyles()

	s.Focused.Prompt = BrandTitle
	s.Focused.Text = Text
	s.Focused.Placeholder = Subtle
	s.Focused.Suggestion = Subtle

	s.Blurred.Prompt = Subtle
	s.Blurred.Text = Muted
	s.Blurred.Placeholder = Subtle
	s.Blurred.Suggestion = Subtle

	s.Cursor.Color = BrandColor

	return s
}
