package components

import (
	"strings"

	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// StyledPath dims everything up to the last "/", so the eye lands on the leaf.
func StyledPath(path string, s styles.RowStyles) string {
	cut := strings.LastIndex(path, "/")
	if cut < 0 {
		return s.Name.Render(path)
	}

	return s.Dir.Render(path[:cut+1]) + s.Name.Render(path[cut+1:])
}
