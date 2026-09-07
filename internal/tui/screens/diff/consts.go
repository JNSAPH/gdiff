package diffview

// The constants more than one of this screen's files reads. Ones only their
// own file uses stay next to the code that reads them.

// The widths the layout is built from.
const (
	// sidebarWidth is the file list's preferred column; minSidebarWidth is
	// how far it gives way before the tabs take over.
	sidebarWidth    = 36
	minSidebarWidth = 18

	// narrowWidth is the window width below which the sidebar gives up its
	// column and the file list moves into a tab strip above the diff.
	narrowWidth = 100

	// minStatusWidth is the narrowest content pane that still shows the diff
	// stats and scroll position beside the file name.
	minStatusWidth = 50

	// statusWidth is the fixed width the scroll percentage is right-aligned
	// in, so the diff stats beside it don't shift as it changes.
	statusWidth = 4
)

// The rows each pane's header takes off the body's height.
const (
	// sidebarHeaderHeight covers the sort state and the file-count divider.
	sidebarHeaderHeight = 2

	// contentHeaderHeight is how many rows contentHeader always renders.
	contentHeaderHeight = 1
)
