# Agent Notes — gdiff

## Project facts

- **What it is:** A terminal UI for browsing a git repository's uncommitted changes — a file sidebar next to a diff pane, plus a worktree list and a checkpoint mechanism for accepting or rejecting agent-written changes file by file.
- **Stack:** Go 1.25. Bubbletea v2, Bubbles v2 and Lipgloss v2, all from `charm.land/...` — **not** `github.com/charmbracelet/...`. go-git v5 for repository access, sergi/go-diff for the line diff, cobra for the CLI.
- **Layout:**
  - `cmd/` — cobra entry point, flags, starts the tea program
  - `internal/core/` — logger setup (`log/slog` → `gdiff.log`, dev only)
  - `internal/git/` — go-git wrapper: `Repo`, `FileChange`, `DiffLine`, worktrees, checkpoints
  - `internal/tui/` — root model and router; owns header, border, command bar
  - `internal/tui/components/` — stateless render helpers shared across screens
  - `internal/tui/styles/` — palette and shared styles; the only place hex colors live
  - `internal/tui/commandbar/` — the `:` command popup
  - `internal/tui/screens/` — one package per screen: `splash`, `diff`, `worktrees`
- **Shared helpers:** `internal/tui/components/` for anything that renders and has no state. Look there before writing a render helper; put new ones there once a second screen needs them. Git access goes through `internal/git`, never `os/exec` at a call site.
- **Shared constants:** `internal/tui/styles/` for colors and chrome dimensions. Per-screen constants read by more than one file of that screen go in that screen's `consts.go` (see `screens/diff/consts.go`); ones only their own file uses stay next to the code that reads them.
- **Commands:** `make check` (`go build ./...`, `go vet ./...`, `gofmt -l .`) · `make run` · `make dev` (debug logging) · `make build` → `bin/gdiff`. There is no test suite yet.
- **Flags:** `-p/--path` (repository, defaults to `.`), `-d/--dev` (debug logging).
- **Generated / do not edit:** `bin/`, `dist/`, `gdiff.log`, `go.sum`.
- **Local conventions that override anything below:**
  - Screens follow a *convention*, not an interface — each `Update` returns its own concrete `Model`. Don't introduce a `Screen` interface.
  - `Update` is a thin dispatcher: every case calls a method on `Model`. No state logic inline in the switch.
  - Only `charm.land/lipgloss/v2`. Never reintroduce lipgloss v1.
  - Every color comes from `internal/tui/styles`. Never call `lipgloss.Color("#...")` outside that package.
  - The directory `internal/tui/screens/diff/` holds package `diffview`, so its import is aliased. Keep the alias; don't rename the package to match the directory.

### File map

```
main.go               -> cmd.Execute
cmd/root.go            cobra entry point, flags, starts the tea program
internal/core/         logger setup (log/slog -> gdiff.log, dev only)
internal/git/
  git.go               Repo, ChangeType, FileChange, DiffLine, lineDiff
  worktree.go          worktree listing + porcelain status parsing
  checkpoint.go        the checkpoint ref: accept/reject, whole-file and per-file
internal/tui/
  app.go               root Init/Update/View — the router
  model.go             root Model + its state transitions
  global/keys.go       bindings the router owns (quit, open command bar)
  styles/              palette and shared styles — the only place hex colors live
    style.go           colors, text styles, AppBorder
    diff.go            diff line/gutter styles and the sidebar RowStyles
    commandbar.go      the popup's width
    splash.go          progress bar styles
  components/          stateless render helpers
    header.go          the app title bar
    footer.go          help model + footer row
    rule.go            label-in-a-border/divider arithmetic
    sidebar.go         the left panel's chrome
    changetype.go      ChangeTypeColor + ChangeCounts, shared by every screen
  commandbar/          the ":" command popup
    commands.go        the commands the bar accepts
  screens/splash/      startup screen
  screens/worktrees/   worktree list, each with its own change summary
  screens/diff/        main screen (package diffview): file sidebar + diff pane
    screen.go          Model, Init/Update/View, header, footer
    model.go           state transitions; loadDiff reads git
    base.go            diffBase: compare against HEAD or the last checkpoint
    consts.go          the widths and header heights more than one file reads
    content.go         diff pane: gutter, bands, content header
    filelist.go        sidebar rows
    sidebar.go         sidebar assembly + windowing
    tabs.go            the tab strip shown instead of the sidebar when narrow
    sort.go            sort modes
    keys.go            key bindings
```

## Working rules

- Ask before starting when the requirement is ambiguous or has more than one reasonable reading. A wrong guess costs more than a question.
- Change what the task needs and nothing else. No reformatting untouched lines, no drive-by fixes. Notice an unrelated bug? Say so, don't fix it here.
- Run `make check` before reporting done. Report failures; don't route around them.
- Delete what you replace. No `_v2`, no old path kept beside the new one, no dead code "just in case".
- Never commit secrets, tokens, or `.env` files, and never print them in chat.
- Existing code that breaks these rules stays as it is unless the task is to fix it.

## Design

- Simplest design that fully satisfies the current requirement. Nothing for hypothetical future needs.
- Established library over a custom implementation. No new dependency, config flag, or abstraction layer without a reason tied to this task.
- Follow the existing architecture unless there's a good reason to depart. State the reason.
- Don't preserve backward compatibility unless asked. State breaking changes plainly.
- **Performance exception:** when code is genuinely hot — here that means a render path, since `View` runs on every message — speed wins over style. Bend any rule here and leave a comment saying what the ugly shape buys. Don't invoke this for code that merely might be hot one day.

## Architecture

**One root model routes to screen models.** `internal/tui.Model` holds only global
state (terminal size, active screen, whether the command bar is open) and embeds one
Model per screen. It is the only `tea.Model` — screens are plain structs, not
`tea.Model` implementations, so their `Update` returns their own concrete type.

**Screens follow a convention, not an interface.** Each screen package exposes:

```go
func New(...) Model
func (m Model) Init() tea.Cmd                     // commands to run while active
func (m Model) Update(tea.Msg) (Model, tea.Cmd)   // concrete Model, not tea.Model
func (m Model) HeaderContent() (string, []string) // title + styled title-bar segments
func (m Model) View() string
```

There's deliberately no `Screen` interface — each `Update` returns a different
concrete type, so an interface would need `tea.Model` and cost a type assertion per
message for no gain at three screens. Add one only if the screen count makes the
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
of screen code. A screen built after startup — `openWorktree` replaces `diffView`
wholesale — has to be handed that size before it renders.

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
- **Change type has one color everywhere.** `components.ChangeTypeColor` maps it, and
  `git.ChangeType.Symbol()` gives the glyph (`+` `~` `-` `»`). The sidebar row, the
  content header, the worktree summary and the header's counts all read from those two.
- **Three levels of grey**, and picking the wrong one is what makes a TUI look noisy:
  `Text` for content, `Muted` for secondary text (a path's directory), `Subtle` for
  chrome that should barely register (gutter numbers, separators).
- **Selection is a background band, not a color change**, so the row's own colors
  survive. `styles.RowStyles` groups the variants so every part of the row — accent,
  glyph, dir, name, padding — carries the same background.
- **Right-aligned status is padded to a fixed width** (`statusWidth`) so numbers
  don't shift the layout as they change.

## Structure inside a function

One readable function with labeled blocks beats three functions with one caller each.
The reader follows the data top to bottom without opening another file.

```go
// fileRows renders the visible slice of the file list, one row per line,
// padded out to the panel's full height.
func (m Model) fileRows(width, rows int) string {
	if rows <= 0 {
		return ""
	}

	end := min(m.listOffset+rows, len(m.files))
	window := m.files[min(m.listOffset, len(m.files)):end]

	list := fileList(window, m.listOffset, m.cursor, m.scrollOffset, width-1)

	// Every row is padded to the same width, and rows the file list didn't
	// fill are left blank, so a short list still fills the panel.
	lines := make([]string, rows)
	listLines := strings.Split(list, "\n")

	for i := range lines {
		row := ""
		if i < len(window) && i < len(listLines) {
			row = listLines[i]
		}
		lines[i] = lipgloss.NewStyle().Width(width - 1).Render(row)
	}

	return strings.Join(lines, "\n")
}
```

- Split a function out only when it's reused, when it enforces a real invariant, or when it's long and the parts stand on their own.
- Number blocks only when the function has three or more stages and each feeds the next. Two stages, or independent ones, get a plain label or nothing.
- Step numbers live inside one function body. Never number a sequence of calls to other functions, never continue across files.

## Sharing code

- Search for an existing helper before writing one. A duplicated helper is worse than no helper.
- Near-identical logic across screens or components belongs in **one** helper in `internal/tui/components/`. `ChangeTypeColor` is the model: three screens needed the same mapping, so it moved once instead of being copied.
- Screen-local logic stays in the screen package. Don't promote something to `components/` before the second caller exists.

## Comments and doc comments

Default: **one line, above the block it describes.**

Worth a comment: what a block does when the code takes a moment to read; why a decision was made when a reader can't infer it; a workaround for an upstream bug (link the issue); anything that behaves surprisingly.

Never write a comment that narrates your change (`// Added tenant check`, `// Refactored to use X`), explains the conversation that produced the code, restates the line below it, or keeps a changelog.

**Leave comments you didn't write alone.** The rules above govern comments *you* add. Update someone else's only when your change makes it wrong.

Longer comments are wanted in exactly two places: a decision worth recording, and unavoidable magic. `git/checkpoint.go` is the example — a ref deliberately kept unreachable from any branch needs its several lines, because someone will otherwise "fix" it into a normal commit. Say what happens, when it runs, and what breaks if someone moves it. If a plain version exists, write the plain version instead.

Doc comments: every exported identifier gets one, starting with its name, in the imperative. Each package gets a package comment on one file. Document a parameter only when its meaning isn't obvious from the name and type. Unexported helpers get none; the name and signature are the contract.

The one exception is the screen convention — `Init`, `Update`, `View` and `HeaderContent`. The shape is documented once under **Architecture**; a doc comment on each implementation would only restate the signature. Leave them bare.

## Typing and naming

- Give data a named type where it **changes shape or gains meaning** — `FileChange`, `DiffLine`, `Worktree`, `diffBase`. Leave pass-through shapes plain.
- Enumerations are a named integer type with `iota` and a method for anything derived from them (`ChangeType.Symbol()`), never bare strings compared at call sites.
- `any` and a bare `map[string]any` admit you don't know the shape. Find out and write it down. `tea.Msg` is the one honest exception — it's the framework's type — so narrow it in a type switch immediately.
- Don't mutate an argument unless the function's name says so. Return new data. The Bubbletea models are value receivers for exactly this reason; keep them that way.
- Names encode what's in the thing and how it's keyed: `filesByPath`, `linesPerPage`. Short names are fine in a two-or-three-line scope, and `m` for the receiver is the house style. Don't pad a name with its type (`fileSlice`). Never `data`, `result`, `temp`, `item` at function scope.

## Constants

A literal with meaning becomes a named constant at the top of its file. Obvious values (`0`, `1`, `""`) stay inline. The moment a second file in the same package needs it, move it to that package's `consts.go`; the moment a second package needs it, it belongs in `styles/` (if it's visual) or beside the type it describes. Don't import it from whichever file declared it first, and don't redeclare it.

## Guards up front, clean body

```go
// loadDiff reads the selected file's diff and points the viewport at the first change.
func (m Model) loadDiff() Model {
	// Guards — each one puts a message in the pane instead of a diff
	if m.loadErr != nil {
		return m.showMessage(styles.Error.Render("error: " + m.loadErr.Error()))
	}

	file, ok := m.selected()
	if !ok {
		return m.showMessage(styles.Muted.Render("No changes"))
	}

	var lines []git.DiffLine
	var err error
	if m.base == baseCheckpoint {
		lines, err = m.repo.FileDiffAgainst(git.CheckpointRef, file.Name())
	} else {
		lines, err = m.repo.FileDiff(file.Name())
	}
	if err != nil {
		return m.showMessage(styles.Error.Render("error loading diff: " + err.Error()))
	}

	// Happy path
	m.diffLines = lines
	m.lineNumWidth = numWidth(lines)
	m.viewport.SetXOffset(0)
	m.viewport.SetYOffset(firstChange(lines))

	return m.renderDiff()
}
```

- Validate at the boundary. After the guards, trust the data.
- Don't check the same thing twice. If the type system or an earlier guard guarantees it, move on.
- Handle the errors git, the filesystem and the terminal actually produce. Don't wrap a call that can't fail, and don't add a fallback for a state the caller can't produce.
- Let programmer errors panic. An index the code just bounds-checked doesn't need a second check.

## Use the platform, not a regex

```go
dir := regexp.MustCompile(`^(.*)/`).FindStringSubmatch(path)[1] // bad
dir := filepath.Dir(path)                                      // good
```

Same for `filepath`/`path` over string surgery on paths, `strings.Cut` and the other `strings` functions over patterns, a real parser over slicing. Use a regex when the task really is pattern matching — then keep it on its own line with a comment saying what it matches. `parsePorcelainStatus` and `parseWorktreeList` parse fixed-format git output by column and field, not by regex; keep new parsers that way.

## Logging

**Logging is dev-only, and must stay that way.** The TUI owns the terminal, so `slog` can't go to stderr; `core.SetupLogger` writes `gdiff.log` only under `--dev` and discards otherwise. gdiff is pointed at *other people's* repositories, and a log file dropped in the working directory shows up as an untracked file in its own file list. `tail -f gdiff.log` in another terminal.

Enough to reconstruct what happened before an error, no more.

- Info for major events: a repo opened, a diff loaded, a checkpoint accepted. Include the paths and counts that make the line useful.
- Warn or error for anything degraded or failed, with the context needed to act.
- No entry/exit tracing, no narration inside a function, and never inside `View` or `Update` — they run on every message.

```go
slog.Info("checkpoint accepted", "path", change.Path, "files", len(m.files))
```

## Tests

There is no test suite yet. When adding one:

- Test **pure logic**: `git.lineDiff`, `git.parsePorcelainStatus`, `git.parseWorktreeList`, `diffview.scrollWindow`/`truncateFront`/`truncateTail`/`pad`, `components.fillAround`. Code with real branching and no terminal.
- Skip tests of `Update` wiring and rendered output unless asked. Golden-file tests of a TUI break on every style tweak and prove little.
- Only write a test that can fail for a real reason.

## Gotchas

- **`resize` re-derives everything size-dependent.** `diffview.resize` sets the
  viewport's width and height too. Anything that changes the layout must go through
  it — `toggleHelp` calls it because expanding the help bar shrinks the viewport, and
  the sidebar/tabs switch at `narrowWidth` happens there.
- **`applySort` sorts `m.files` in place.** Model is passed by value but a slice header
  isn't a deep copy, so this mutates the backing array the previous Model also points
  at. Safe only because Bubbletea discards the old model; don't rely on a previous
  Model's slice contents.
- **The cursor is an index into display order.** Re-sorting moves the selection rather
  than following the selected file. Deliberate; change it by matching on identity.
- **`loadDiff` reads git; `renderDiff` only redraws.** The diff stays on the model as
  `m.diffLines`, so a resize re-renders without touching git. Keep that split — don't
  read git from a render path.
- **The router owns the marquee ticker, not the diff screen.** A screen that
  re-arms its own `tea.Tick` from its `Update` loses the chain the moment one
  fires while another screen or the command bar is handling messages, and
  starts a second chain every time the screen is re-entered. Both happened.
  `marqueeTick` lives in `tui` and re-arms unconditionally, so exactly one
  chain runs for the app's lifetime. Don't move it back into `diffview`.
- **Checkpoint operations write to the working tree.** `RejectCheckpoint*` overwrites
  and deletes real files. Treat anything in `git/checkpoint.go` as destructive and
  read the whole function before changing it.

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
- **Load errors surface as text in the diff pane.** `loadDiff` hands them to
  `showMessage`, so an error looks like content; there's no dedicated error state
  and no retry.
- **No tests.** The pure functions listed above are the cheap place to start.

## How to answer

Clear, simple language. Active voice, short words, no clichés. Explain a technical term only when the surrounding work suggests it's unfamiliar. Say what changed and why in a few sentences — don't restate the diff or add summary tables.

## Before you finish

Reread your own diff and remove:

- Comments that narrate the change or explain the conversation.
- Functions with one caller that could be a labeled block.
- Copy-pasted logic, or a helper that already exists in `internal/tui/components/`.
- A hex color written anywhere but `internal/tui/styles/`.
- A constant duplicated across files instead of moved to the package's `consts.go`.
- Defensive checks for cases that can't happen.
- Log lines that trace steps instead of recording events, or any log on a render path.
- Reformatting and drive-by edits the task didn't call for.
- Config flags, options, and abstraction nobody asked for.
