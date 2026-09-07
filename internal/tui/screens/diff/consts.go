package diffview

// The constants more than one of this screen's files reads.

// The widths the layout is built from.
const (
	// sidebarWidth is the preferred column; minSidebarWidth is where it yields.
	sidebarWidth    = 36
	minSidebarWidth = 18

	// narrowWidth is where the sidebar gives up its column for a tab strip.
	narrowWidth = 100

	// minStatusWidth is the narrowest pane that still fits the stats.
	minStatusWidth = 50

	// statusWidth is fixed, so the stats beside it don't shift as it changes.
	statusWidth = 4
)

// The rows each pane's header takes off the body's height.
const (
	// sidebarHeaderHeight covers the sort state and the file-count divider.
	sidebarHeaderHeight = 2

	// contentHeaderHeight is how many rows contentHeader always renders.
	contentHeaderHeight = 1
)
