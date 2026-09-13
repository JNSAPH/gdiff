package components

// TruncateFront drops characters from the front, keeping a path's tail.
func TruncateFront(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= maxWidth {
		return name
	}

	const ellipsis = "…"
	keep := maxWidth - 1 // the ellipsis takes one column
	if keep <= 0 {
		return ellipsis
	}

	return ellipsis + string(runes[len(runes)-keep:])
}

// TruncateTail drops them from the end, for a name whose front identifies it.
func TruncateTail(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= maxWidth {
		return name
	}

	const ellipsis = "…"
	keep := maxWidth - 1 // the ellipsis takes one column
	if keep <= 0 {
		return ellipsis
	}

	return string(runes[:keep]) + ellipsis
}
