# gdiff

A terminal UI for browsing a git repository's changes, built with Bubbletea v2.

```
make run     # build + run
make dev     # build + run with --dev (debug logging)
make build   # -> bin/gdiff
go build ./... && go vet ./...
```

Flags: `-p/--path` (repository, defaults to `.`), `-d/--dev` (debug logging).

## For Software Tasks

- Choose the simplest design that fully satisfies the current requirements.
- Do not add features for hypothetical future needs.
- Prefer established, well-maintained libraries over custom implementations.
- Do not preserve backward compatibility unless the requirements explicitly ask for it. Clearly state any breaking changes.
- Optimize for readability and maintainability before cleverness.
- Follow the existing project's style and architecture unless there is a good reason to change it.
- If a trade-off exists, explain it briefly before choosing.

## Comments

- Never more than 2 lines; 1 is the target.
- Plain language, no flourish — understandable to someone seeing the file for the first time.
- Give each code block a headline comment, so the block can be skipped without reading it.
- Comment the *why*, not the *what*. Don't restate the code.

## Layout

```
main.go               -> cmd.Execute
cmd/root.go            cobra entry point, flags, starts the tea program
internal/core/         logger setup (log/slog -> gdiff.log, dev only)
internal/git/          go-git wrapper: Repo, FileChange, DiffLine
internal/tui/
  app.go               root Init/Update/View — the router
  model.go             root Model + its state transitions
  global/keys.go       bindings the router owns (quit, open command bar)
  styles/              palette and shared styles — the only place hex colors live
    style.go           colors, text styles, AppBorder
    diff.go            diff line/gutter styles and the sidebar RowStyles
  components/          stateless render helpers
    header.go          the app title bar
    rule.go            label-in-a-border/divider arithmetic
    sidebar.go         the left panel's chrome
    scrollbar.go       one-column vertical scrollbar
  commandbar/          the ":" command popup
  screens/splash/      startup screen (logo.go holds the block letters)
  screens/diffview/    main screen: file sidebar + diff pane
    diffview.go        Model, Init/Update/View, header, footer
    model.go           state transitions; loadDiff reads git
    content.go         diff pane: gutter, bands, content header
    filelist.go        sidebar rows
    sidebar.go         sidebar assembly + windowing
    sort.go            sort modes
    keys.go            key bindings
```

## Architecture

**One root model routes to screen models.** `internal/tui.Model` holds only global
state (terminal size, active screen, whether the command bar is open) and embeds one
Model per screen. It is the only `tea.Model` — screens are plain structs, not
`tea.Model` implementations, so their `Update` returns their own concrete type.

**Screens follow a convention, not an interface.** Each screen package exposes:

```go
func New(...) Model
func (m Model) Init() tea.Cmd                    // commands to run while active
func (m Model) Update(tea.Msg) (Model, tea.Cmd)  // concrete Model, not tea.Model
func (m Model) HeaderContent() (string, []string) // title + styled title-bar segments
func (m Model) View() string
```

There's deliberately no `Screen` interface — each `Update` returns a different
concrete type, so an interface would need `tea.Model` and cost a type assertion per
message for no gain at two screens. Add one only if the screen count makes the
router's switch statements unwieldy.

**`Update` is a thin dispatcher.** Every case calls a method on `Model`; no state
logic is written inline in the switch. Those methods return `(Model, tea.Cmd)` when
they need to emit a command and plain `Model` when they don't — never a bare `Model`
where a command was required, which once silently broke the quit key. State
transitions live in `model.go`; `Init`/`Update`/`View` live in `app.go` (root) or the
screen's own file.

**Screens don't know about the app's chrome.** The router owns the header, border and
command-bar popup, subtracts their size, and passes each screen a `tea.WindowSizeMsg`
with the space it actually gets. A screen reporting its title via `HeaderContent()`
instead of rendering its own header is what keeps `styles.AppBorderWidthOverhead` out
of screen code.

**Message ownership.** A message crossing a package boundary is exported
(`splash.DoneMsg`, `commandbar.SubmitMsg`/`CloseMsg`); one that never leaves its
package is not (`tickMsg`). Same rule for everything else — `diffview` exports only
what the router calls.

## Rendering

**Only `charm.land/lipgloss/v2`.** The repo briefly had both v2 and
`github.com/charmbracelet/lipgloss` v1; their `Color` types aren't interchangeable,
which forced colors to be passed around as hex strings. Don't reintroduce v1.

**All colors come from `internal/tui/styles`.** The hex constants there are
unexported. Add a named style or color to that package rather than calling
`lipgloss.Color("#...")` anywhere else. The palette assumes a dark terminal: it paints
its own backgrounds (diff bands, the selected row) and pairs them with light text.

**`Style.Width(n)` is the TOTAL width; `Style.Height(n)` is the INTERIOR height.** This
asymmetry is the single easiest thing to get wrong here, and it has caused real
layout bugs twice. With a bordered style, `Width(n)` renders exactly `n` columns —
the border eats into it, leaving `n-2` for content (`n-1` with one side). `Height(n)`
does the opposite: it sets the content rows and the border adds on top. So:

```go
box := styles.AppBorder.Width(m.width).                       // total
    Height(m.height - AppBorderHeightOverhead - HeaderHeight) // interior
```

Verify with a throwaway `lipgloss.Width(style.Render(...))` rather than reasoning
about it. And `Width()` **wraps** overflowing text, it does not truncate — use
`MaxWidth()` to truncate, which is why every header clamps itself.

**Header and command bar draw their own top border row.** Both boxes set
`BorderTop(false)` and join a hand-built row above themselves, so a label can sit in
the border. `components.TopBorderWithLabel` and `components.Rule` do that arithmetic —
use them instead of writing another `strings.Repeat` fill.

**Measure, don't hardcode.** `diffview.footerHeight()` renders the footer and measures
it with `lipgloss.Height`, because the help bar's height changes when `?` expands it.
`gutterWidth()` does the same by rendering an empty gutter. Where a constant is
unavoidable it sits next to the thing it describes (`components.HeaderHeight`,
`styles.AppBorder*Overhead`, `contentHeaderHeight`).

**Width in columns, not bytes.** Use `lipgloss.Width` and `[]rune`, never `len()`, on
anything user-visible. `len("…")` is 3.

**The diff pane is built from viewport hooks, not pre-styled text.** `LeftGutterFunc`
draws the line numbers and `+`/`-` sign and stays pinned during horizontal scrolling;
`StyleLineFunc` colors each row. Content lines are padded to the pane width in
`renderDiff` so the background bands span the full row — `StyleLineFunc` supplies
color only, because a style with `Width` set would wrap. `SoftWrap` is off so long
lines scroll sideways instead of wrapping and shredding the frame.

**Rendering it to look at it.** A TUI can't be checked by reading the code. Run it
under a pty and feed the output to a real VT emulator (`pyte`), then print the screen
grid plus the distinct (fg, bg) pairs — that catches both layout overflow and styles
that silently didn't apply. Stripping ANSI with a regex does *not* work: Bubble Tea
redraws incrementally with cursor positioning, so you get garbled fragments.

## Visual language

Keep new UI consistent with what's there:

- **One accent.** Brand orange marks focus and identity only — the focused pane's
  border/bar, the selected row's `▌`, the logo, the command prompt. Never decoration.
- **Change type has one color everywhere.** `changeTypeColor` maps it, and
  `ChangeType.Symbol()` gives the glyph (`+` `~` `-` `»`). The sidebar row, the
  content header and the header's counts all read from those two.
- **Three levels of grey**, and picking the wrong one is what makes a TUI look noisy:
  `Text` for content, `Muted` for secondary text (a path's directory), `Subtle` for
  chrome that should barely register (scrollbar track, gutter numbers, separators).
- **Selection is a background band, not a color change**, so the row's own colors
  survive. `styles.RowStyles` groups the variants so every part of the row — accent,
  glyph, dir, name, padding — carries the same background.
- **Right-aligned status is padded to a fixed width** (`statusWidth`) so numbers
  don't shift the layout as they change.

## Gotchas

- **`resize` re-derives everything size-dependent.** `diffview.resize` sets the
  viewport's width and height too. Anything that changes the layout must go through
  it — `toggleHelp` calls it because expanding the help bar shrinks the viewport.
- **`applySort` sorts `m.files` in place.** Model is passed by value but a slice header
  isn't a deep copy, so this mutates the backing array the previous Model also points
  at. Safe only because Bubbletea discards the old model; don't rely on a previous
  Model's slice contents.
- **The cursor is an index into display order.** Re-sorting moves the selection rather
  than following the selected file. Deliberate; change it by matching on identity.
- **`loadDiff` reads git; `renderDiff` only redraws.** The diff stays on the model as
  `m.diffLines`, so a resize re-renders without touching git. Keep that split — don't
  read git from a render path.
- **Logging is dev-only.** The TUI owns the terminal, so `slog` can't go to stderr;
  `core.SetupLogger` writes `gdiff.log` only under `--dev` and discards otherwise. It
  must stay that way: gdiff is pointed at *other people's* repositories, and a log file
  dropped in the working directory shows up as an untracked file in its own file list.
  `tail -f gdiff.log` in another terminal.

## Known gaps

- **Git I/O is synchronous.** `diffview.New` opens the repo and lists changes, and
  `loadDiff` reads and diffs a file inside `Update` — both block the event loop.
  The Bubbletea-idiomatic fix is to return a `tea.Cmd` that does the work and delivers
  a result message, with a loading state in the model. Worth doing before diffs get
  bigger.
- **No hunks.** A diff is the whole file, with every unchanged line as context.
  `loadDiff` opens the pane at the first change so this isn't felt immediately, but
  collapsing long context runs into `@@` headers is the real fix.
- **The file list never refreshes.** Changes are read once at startup; editing a file
  in another window doesn't update the list. Needs a reload key, or a watcher.
- **Load errors surface as text in the diff pane.** `diffview.Model.loadErr` is
  rendered by `loadDiff`; there's no dedicated error state or retry.
- **No tests.** The pure functions are the cheap place to start: `git.lineDiff`,
  `diffview.scrollWindow`/`truncateFront`, `components.fillAround`.
