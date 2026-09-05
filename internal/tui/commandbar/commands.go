package commandbar

import "charm.land/lipgloss/v2"

// Commands the bar accepts. Add one here and it becomes an input
// suggestion; handle it in the router's commandbar.SubmitMsg case.
const (
	CommandQuit      = "quit"
	CommandWorktrees = "worktrees"
	CommandDiff      = "diff"
	CommandAccept    = "checkpoint/accept"      // the selected file only
	CommandAcceptAll = "checkpoint/accept-all"  // every pending file
	CommandReject    = "checkpoint/reject-file" // the selected file only
	CommandRejectAll = "checkpoint/reject-all"  // every pending file
)

type command struct {
	name string
	desc string
}

var commandList = []command{
	{CommandQuit, "quit gdiff"},
	{CommandWorktrees, "browse worktrees"},
	{CommandDiff, "back to the diff view"},
	{CommandAccept, "accept the selected file"},
	{CommandAcceptAll, "accept every pending file"},
	{CommandReject, "reject the selected file"},
	{CommandRejectAll, "reject every pending file"},
}

var Commands = func() []string {
	names := make([]string, len(commandList))
	for i, c := range commandList {
		names[i] = c.name
	}
	return names
}()

var descriptions = func() map[string]string {
	m := make(map[string]string, len(commandList))
	for _, c := range commandList {
		m[c.name] = c.desc
	}
	return m
}()

// descColWidth is the suggestion list's description column width, derived
// from the longest description rather than hand-picked, so it can't go
// stale as commandList changes.
var descColWidth = func() int {
	w := 0
	for _, c := range commandList {
		w = max(w, lipgloss.Width(c.desc))
	}
	return w
}()
