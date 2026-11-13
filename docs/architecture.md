# Architecture

0go has one application package and no background service or persistent state.
Browsing reads directory entries without recursive indexing or file-content
reads. Tree view caches the current listing, its parent, and one level inside
the highlighted directory. It does not execute external commands to list entries.

## Application

| File | Responsibility |
| --- | --- |
| `main.go` | Process entry point and exit status |
| `cli.go` | Options, terminal selection, stdout/stderr contract |
| `errors.go` | Operation context with an unwrap-compatible error chain |
| `browser.go` | Directory listing, filtering, navigation, and keyboard help state |
| `entry.go` | File/directory/link classification and directories-first sorting |
| `view.go` | Layout, safe text clipping, theme colors, and contextual help |
| `tree.go` | Parent/current/child panes with one active middle list |
| `terminal.go` | Raw terminal mode, key decoding, rendering, resize and signal handling |
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

The renderer controls the alternate screen and cursor visibility. Cleanup restores
terminal settings on normal exit, cancellation, and handled termination signals.
An uncatchable process kill cannot run cleanup. Dynamic text is stripped of
non-printable characters before display; selected paths retain their original
bytes when passed back to the shell.

Text is clipped by terminal-cell budget before trusted ANSI styles are applied,
so clipping cannot leave an incomplete style sequence. Each styled span resets
its attributes; terminal cleanup also resets them before returning to the shell.
The final terminal column is reserved to avoid automatic wrapping. The footer
stays at the bottom in ordinary windows, and small windows prioritize keeping
the selected row visible.

Tree view is the default. Only the middle list owns the cursor; parent and child
panes are read-only context. Navigation and refresh update their cached entries,
and cursor movement refreshes child context only when its path changes. Files
can be highlighted but never opened or returned as a selected destination.
Symlink targets are checked for directory status without resolving their logical
paths. Colors and suffixes distinguish file and directory-link entries.

The `v` key switches to a directory-only compact list, preserving folder focus.
In compact view, when the list exceeds the available row count and each column can keep
at least 24 cells, the renderer uses two columns separated by a three-cell
divider. Both columns share one ordered list and one cursor. Page navigation
uses their combined capacity; Tab moves between corresponding rows without
changing directory navigation keys. Short lists and narrow terminals keep one
column. Horizontal terminal positioning anchors the divider independently of
Unicode glyph widths. Dynamic text still passes through sanitization and clipping.

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
