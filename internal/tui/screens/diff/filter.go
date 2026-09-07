package diffview

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/JNSAPH/gdiff/internal/git"
	"github.com/JNSAPH/gdiff/internal/tui/styles"
)

// newFilterInput builds the "/" input that sits in the file list's header.
func newFilterInput() textinput.Model {
	in := textinput.New()
	in.Prompt = "/"
	in.Placeholder = "filter files"
	in.SetStyles(styles.TextInput())
	return in
}

// filterActive reports whether a query is narrowing the list.
func (m Model) filterActive() bool {
	return m.filtering || m.filter.Value() != ""
}

// CapturesInput tells the router to hold its globals back while typing.
func (m Model) CapturesInput() bool {
	return m.filtering
}

// toggleFilter is what "/" does: open the input, or clear an applied query.
func (m Model) toggleFilter() (Model, tea.Cmd) {
	if m.filter.Value() != "" {
		return m.clearFilter(), nil
	}

	m.filtering = true
	m = m.setFocus(focusSidebar)
	return m, m.filter.Focus()
}

// clearFilter drops the query and gives the whole list back.
func (m Model) clearFilter() Model {
	m.filtering = false
	m.filter.Blur()
	m.filter.Reset()
	return m.refreshList()
}

// lockFilter keeps the query but hands the keyboard back to the file list.
func (m Model) lockFilter() Model {
	m.filtering = false
	m.filter.Blur()
	return m
}

// updateFilter feeds a key to the input and re-derives the narrowed list.
func (m Model) updateFilter(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m.refreshList(), cmd
}

// applyFilter derives the display list from a case-insensitive path match.
func (m Model) applyFilter() Model {
	query := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if query == "" {
		m.files = m.allFiles
		return m
	}

	files := make([]git.FileChange, 0, len(m.allFiles))
	for _, f := range m.allFiles {
		if strings.Contains(strings.ToLower(f.Name()), query) {
			files = append(files, f)
		}
	}

	m.files = files
	return m
}

// filterRow is the header's filter line: the live input, or the applied query.
func (m Model) filterRow(width int) string {
	if m.filtering {
		return clamp(m.filter.View(), width)
	}

	query := styles.BrandTitle.Render("/") + styles.Text.Render(m.filter.Value())
	return spread(query, m.listPosition(), width)
}
