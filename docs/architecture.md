# Architecture

0go has one application package and no background service or persistent state.
Browsing reads one filesystem level at a time. It does not execute external
commands to list directories or perform recursive indexing.

## Application

| File | Responsibility |
| --- | --- |
| `main.go` | Process entry point and exit status |
| `cli.go` | Options, terminal selection, stdout/stderr contract |
| `browser.go` | Directory listing, filtering, navigation, and plain-text view |
| `terminal.go` | Raw terminal mode, key decoding, rendering, resize and signal handling |
| `shell.go` | zsh/bash integration with literal path handling |

The terminal loop uses `poll` with a bounded wait to process keyboard input,
terminal resize events, and cancellation. Input decoding accepts fragmented UTF-8
and common CSI/SS3 key sequences. A short timeout distinguishes a standalone Esc
from the start of an escape sequence. Unknown sequences are discarded.

The renderer controls the alternate screen and cursor visibility. Cleanup restores
terminal settings on normal exit, cancellation, and handled termination signals.
An uncatchable process kill cannot run cleanup. Dynamic text is stripped of
non-printable characters before display; selected paths retain their original
bytes when passed back to the shell.

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

## Publication boundary

GitHub Actions validates source and builds. It does not create a GitHub Release,
publish containers, or deploy anything. Repository visibility, private security
reporting, and branch-protection settings are configured separately on GitHub.
Local builds and validation do not imply that hosted CI or publication ran.
