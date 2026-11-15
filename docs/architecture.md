# Architecture

cdd has one application package and no background service. It persists only the
last chosen view; browsing paths, filters, and tree expansion stay in memory.
Browsing reads directory entries without recursive indexing or file-content
reads except explicit preview. Directory creation uses one literal child name with `os.Mkdir`; it does
not overwrite existing entries or create intermediate directories.
Empty-file creation uses the same literal-name input with Shift+N, targeting the
highlighted directory in both modes or the highlighted file's parent. It uses
`os.OpenFile` with `O_CREATE|O_EXCL`, never truncation, closes immediately, and
refreshes focus without returning a file as a directory destination.
Rename uses an exclusive native rename (`RENAME_EXCL` on macOS,
`RENAME_NOREPLACE` on Linux). It captures the entry with `Lstat`, resolves the
parent without following the final symlink, and rechecks identity before moving.
No overwrite fallback is used. Current/starting locations and ancestors are
protected. Modal input freezes the target; successful renames invalidate affected
tree branches and refresh selection. Footer action hit testing shares rendered
button rows and respects clipped layouts.
Trash uses the same frozen target and identity checks, with an initially blank
exact-basename confirmation. The worker receives an immutable snapshot and
revalidates it before invocation. macOS calls
[Foundation's Trash API](https://developer.apple.com/documentation/foundation/filemanager/trashitem(at:resultingitemurl:))
through a fixed JXA script; Linux calls `gio trash -- PATH`, using
[GLib's recoverable Trash operation](https://gnome.pages.gitlab.gnome.org/gtk/gio/method.File.trash.html).
Paths are literal
arguments, and neither backend falls back to permanent deletion. Missing Trash
support fails visibly. The modal locks during the operation; success refreshes
the listing and failure clears confirmation for retry. Native calls run through
the cancellable external-action runner, keeping diagnostics out of stdout.
Panes view caches the current listing, its parent, and one level inside
the highlighted directory. It does not execute external commands to list entries.

## Application

| File | Responsibility |
| --- | --- |
| `main.go` | Process entry point and exit status |
| `cli.go` | Options, terminal selection, stdout/stderr contract |
| `errors.go` | Operation context with an unwrap-compatible error chain |
| `browser.go` | Directory listing, creation, filtering, navigation, and help state |
| `entry.go` | File/directory/link classification and directories-first sorting |
| `view.go` | Layout, safe text clipping, theme colors, and contextual help |
| `tree.go` | Parent/current/child panes with one active middle list |
| `outline.go` | View cycle, lazy expandable tree, branch state, and tree hit testing |
| `preferences.go` | Last-view storage, validation, and startup restoration |
| `preview.go` | Bounded text preview, scrolling, and platform preview selection |
| `actions.go` | Direct external CLI invocation with cancellation and bounded diagnostics |
| `mutation.go`, `rename_*.go` | Frozen targets, exclusive renaming, and listing refresh |
| `trash.go` | Exact-name confirmation and system Trash commands |
| `entry_buttons.go`, `mutation_view.go` | Entry action buttons and modal name input |
| `config.go`, `config/default.json` | Embedded key defaults and validated user overrides |
| `terminal.go` | Raw terminal mode, key decoding, rendering, resize and signal handling |
| `keyboard.go` | Extended keyboard decoding and platform confirmation shortcuts |
| `mouse.go` | Mouse report decoding, visible-row hit testing, and click handling |
| `shell.go` | zsh/bash integration with literal path handling |

The application's small argument parser preserves the previous `flag` syntax:
single or double dashes, `--name=value`, standard boolean values, `--`, and
stopping at the first positional argument. Tests compare it against `flag`;
the application itself imports neither `flag` nor `fmt`. It writes messages
with `io.WriteString` and formats counts and quoted paths with `strconv`.
Operation errors implement `Unwrap`, preserving `errors.Is` and `errors.As`
through contextual messages and terminal cleanup errors.

The Go terminal modules still import `fmt` transitively. Avoiding formatting
calls in the application nevertheless lets the linker remove unused code;
this does not claim to remove `fmt` from the full package graph. Executable
size measurements are recorded in [verification](testing.md#executable-size-validation-2026-11-13).

The terminal loop uses `poll` with a bounded wait to process keyboard input,
terminal resize events, and cancellation. Input decoding accepts fragmented UTF-8
and common CSI/SS3 key sequences. A short timeout distinguishes a standalone Esc
from the start of an escape sequence. Unknown sequences are discarded.

The alternate screen pushes the kitty keyboard protocol's disambiguation flag
and pops it during cleanup. The decoder accepts CSI-u control keys, including
Cmd+Enter/Backspace on macOS and Ctrl+Enter/Backspace on Linux, while retaining
legacy UTF-8/CSI/SS3 input. It ignores release events and confirmation repeats. Lock modifiers do not
affect shortcuts. In panes views, plain Enter confirms the current directory
on the `.` row and opens other directories for browsing. In tree view, Enter
confirms any highlighted directory, including while filtering. The confirmation shortcut and button
select the highlighted directory. Legacy ambiguous Enter input follows the same
rules as plain Enter. While naming a directory, Enter creates it; in panes
filters, Enter only finishes input.
Modified Backspace opens Trash confirmation only outside other text input;
repeats and releases never trigger it. Plain Backspace retains navigation and
editing behavior. Confirmation modals capture input before navigation or clicks.

Mouse input uses [xterm button tracking and SGR coordinates](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking),
with legacy mouse packets consumed as a whole for terminals without SGR support.
Hit testing shares the renderer's pane widths, row counts, and scroll position.
The viewport stays stable when highlighting another visible row. Double-clicks
require the same path and row within 400 ms, allowing two cells of horizontal
pointer drift; navigation, keyboard input, and
terminal resizing reset the click state. Double-clicks on `.` finish filter editing
and invoke the confirmation handler; other directories use the keyboard navigation
handler. In expandable tree view, directory double-clicks toggle branches in
place; disclosure arrows provide the same action with a single click.
Clicking any of the footer button's four rows
invokes the same confirmation handler as modified Enter, including its directory
checks. Button hit testing follows the centered rendered block after clipping,
help expansion, and input-mode changes. The outlined block pairs the action and
platform shortcut on one line, with the sanitized destination beneath it and its
final name emphasized. It appears above the keyboard shortcuts in the footer.
Its shortcut label is Enter for tree directories and the panes current row,
the platform confirmation shortcut for other panes directories, and empty for files. Short controls
omit opening and confirmation hints for files. Button size stays stable as the label changes.
Box-drawing glyphs and UI arrows use one cell; other non-ASCII text retains
conservative two-cell budgeting. The button's right edge is anchored explicitly
so narrower filename glyphs cannot displace the border.

The design applies clear action labels and visual hierarchy from
[NN/g's button guidance](https://www.nngroup.com/articles/button-states-communicate-interaction/),
visible button boundaries used by [Textual](https://textual.textualize.io/widgets/button/),
and generous pointer targets recommended by
[W3C](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum).
These principles inform the terminal layout; CSS pixel thresholds are not treated
as terminal-cell measurements. The entire outlined block is clickable. A double-click
confirms a shell change only on the panes current-directory row; tree
double-clicks always toggle branches.

In panes views, the emphasized `← Back to parent` action precedes the current-directory row `.` and filesystem
entries. The current row is highlighted initially; Home focuses the first row.
Both navigation rows bypass name and hidden-entry filtering and do not contribute
to entry counts or child previews. The root omits `..`; elsewhere it uses the logical parent path
and restores focus to the directory being left when opened.
Left, h, and Backspace first focus the parent row, then open it on a second press.
They leave the filesystem root unchanged when no parent row exists.

The renderer controls the alternate screen and cursor visibility. Cleanup restores
terminal settings and disables mouse reporting on normal exit, cancellation,
and handled termination signals.
An uncatchable process kill cannot run cleanup. Dynamic text is stripped of
non-printable characters before display; selected paths retain their original
bytes when passed back to the shell.

Text is clipped by terminal-cell budget before trusted ANSI styles are applied,
so clipping cannot leave an incomplete style sequence. Each styled span resets
its attributes; terminal cleanup also resets them before returning to the shell.
The final terminal column is reserved to avoid automatic wrapping. The footer
stays at the bottom in ordinary windows, and small windows prioritize keeping
the selected row visible.

Panes view is second and also an emergency fallback if tree initialization fails.
Only the middle list owns the cursor; parent and child
panes are read-only context. Navigation and refresh update their cached entries,
and cursor movement refreshes child context only when its path changes. Files
can be opened explicitly in their default application but never returned as a
selected directory destination.
Symlink targets are checked for directory status without resolving their logical
paths. Colors distinguish entry types; suffixes mark symlinks. Folder labels
have no trailing slash.

The `v` key switches between expandable tree and panes,
preserving folder focus. Leaving the tree with a nested selection opens its
containing directory in the flat view, keeping the same selected logical path.

After terminal availability and startup validation, the CLI loads `cdd/last-view`
from `os.UserConfigDir`. The file contains a stable name (`panes` or `tree`),
not an enum ordinal. Missing, unreadable, removed `list`, or unknown values use tree;
an unavailable saved tree keeps panes and reports its initialization
error. Each successful view change atomically replaces the preference through a
private temporary file in the same directory. Failed changes do not save, while
save failures leave navigation usable and report through the terminal. Cancellation
does not revert the last view. Pure browser construction performs no preference
I/O, keeping tests and noninteractive commands independent of personal settings.

Space previews a regular file explicitly. On Linux, a nonblocking open and a
post-open type check reject special files; at most 256 KiB is decoded as text.
Binary files show metadata. Rendering sanitizes file contents before styling.
On macOS, `qlmanage -p` receives the absolute path as one literal argument.
The terminal owns one cancellable child operation and joins it during cleanup.
Child stdout is discarded, stderr is bounded, and failures appear in the UI.
No process receives the terminal input stream or the selected-path output stream.
Enter on a regular file requests `open` on macOS or `xdg-open` on Linux, with
one absolute argument and no shell interpolation. The terminal keeps browsing
state; a new action cancels and joins the previous launcher first. File opening
does not pass through directory confirmation. Linux opener behavior follows
the [xdg-open specification](https://portland.freedesktop.org/doc/xdg-open.html).

The CLI also reads optional `config.json` beside the view preference, after
startup validation. It merges typed overrides onto embedded JSON defaults and
rejects unknown keys, invalid letters, and collisions with built-in shortcuts.
Browser key handling and help use the resulting editor binding, retaining the
old `keys.open_in_zed` alias. Separate `editor_dir_command` and
`editor_file_command` strings provide quoted arguments and `||` alternatives:
defaults prefer `zed`, then `zeditor`, then `code`. The selected entry's type
chooses the command, including symlinks and preview dispatch. Each can be
disabled independently. Legacy `editor_command` initializes both, with explicit
dir/file settings taking precedence. Zed dir defaults use `--classic` to match
open projects; file defaults retain `--new`. The launcher resolves the first
installed CLI through PATH and validates an existing
folder or regular file. It appends `--` and the selected absolute path as literal
arguments; no shell, variable expansion, or path interpolation is involved.
Absence is a no-op; an installed editor's launch failure is reported without
trying another editor. No parent folder is added for a file. The action carries
its configured command into the terminal-owned cancellable runner.

The expandable tree follows the arrow-key model from the
[WAI-ARIA tree pattern](https://www.w3.org/WAI/ARIA/apg/patterns/treeview/): Right
expands or moves to a first child, and Left collapses or focuses a parent.
Indentation, visible disclosure state, and guide lines follow established
filesystem tree conventions also exposed by
[Textual's DirectoryTree](https://textual.textualize.io/widgets/directory_tree/).
The terminal adaptation uses Enter to confirm any directory and keeps the
Change Dir button; it does not implement a browser ARIA widget.

The tree root is home when it contains the browsing location, otherwise `/`.
Entering the mode reads and expands the ancestor chain, preserving logical paths
and focus, including hidden ancestors. There are no synthetic navigation rows.
The immutable launch directory is captured separately from the browsing path,
marked with a dim foreground and a text suffix, and shown in the From header.
An explicit CLI path changes browsing without changing the shell's launch marker.

Each loaded branch caches immediate entries and an expanded flag, keyed by its
logical path. A flattened visible list shares the existing cursor, paging,
confirmation, and mouse row mapping. Parallel row metadata records parents and
guide lines. Rendering and filtering never read directories; initial ancestor
revelation, expansion, and explicit refresh load directories. Filtering retains ancestors within expanded
branches. Symlink cycles cannot cause automatic recursion because unseen logical
paths have no loaded children. Refresh drops stale caches and falls back to a
visible ancestor when focus disappears. Tree creation uses the highlighted
directory or a file's parent and expands the destination to reveal the new child.
Long indentation is clipped from the left so deep labels retain usable space.

Both modes keep one active selection column. Page navigation moves by its visible
row count; Tab does not change focus. Pane dividers use absolute terminal columns
independently of Unicode glyph widths. Dynamic text is sanitized and clipped.

Color detection uses the actual UI terminal, not stdout: stdout is commonly
captured by shell integration. Standard ANSI cyan and red inherit the terminal
theme. The design follows [CLI Guidelines](https://clig.dev/#output) for stream
separation and color control, and the [NO_COLOR convention](https://no-color.org/)
for suppressing added color while retaining non-color emphasis. Selection also
has a text marker, and errors have an explicit label.

Non-ASCII runes conservatively reserve two terminal cells when truncating lines.
This avoids a separate Unicode-width dependency, at the cost of unused horizontal
space for some names. The interface requires an ANSI-compatible terminal.

## Dependencies

The application module has only two external dependencies:

- `golang.org/x/term`: raw mode, terminal size, and terminal restoration.
- `golang.org/x/sys`: operating-system interfaces, including `poll`; also used by
  `x/term`.

Both are maintained within the Go project. Exact versions and checksums are in
the root `go.mod` and `go.sum`. New runtime dependencies require a clear need and
review of maintainers, licensing, and the entire transitive graph.

## Development tools

[tools/](../tools/README.md) is a separate Go module. It pins `govulncheck` and
its transitive dependencies with a `tool` directive and its own lock files.
The application does not import this module or the scanner.

The standard-library-only `tools/notices` command asks `go list -deps` for the
application's packages on each supported build target, then collects their
modules' license, NOTICE, and patent files. It includes the effective Go
toolchain's standard library license. Test-only and development-tool dependencies
are excluded. A module replacement requires explicit licensing review before
packaging, and missing or empty primary licenses fail the build.

The generated notices contain module paths and versions, not local module-cache
paths. They are build output, not a second maintained license tree.

The standard-library-only `tools/release` command creates portable archives for
Linux amd64/arm64 and macOS amd64/arm64. It bundles licenses, executable permissions,
standalone instructions, and version-specific notes from CHANGELOG.md, and writes
a SHA-256 manifest. Runtime directory-selection code does not import build tools.

## Publication boundary

GitHub Actions validates source and builds. A stable version-tag push can publish
four release archives after reusable CI passes, using matching curated notes.
The workflow uploads into a draft and publishes only after upload succeeds.
Repository visibility, private security
reporting, and branch-protection settings are configured separately on GitHub.
Local builds and validation do not imply that hosted CI or publication ran.
