# Changelog

User-facing release notes. Changes below are not released until a matching
version is published on GitHub. Development history is recorded separately in
[the development journal](docs/development-journal.md).

## Unreleased

### Added

- Install through the source-built `nik-makarov/tap` Homebrew tap, with automatic
  versioned releases and formula updates after successful pushes to main.
- Browse directories with arrow keys or `hjkl`, then choose one with Enter.
- Filter directory names, show hidden directories, and follow directory symlinks.
- Change the current shell's directory through zsh or bash integration.
- Report the original and destination paths after a successful shell directory
  change, with safely escaped paths on stderr.
- Cancel selection without changing the current working directory.
- Use a theme-aware directory picker with a full-row selection bar, highlighted
  current path, compact controls, and expandable keyboard help via `?`.
- Disable added UI colors with `NO_COLOR` while retaining a visible selection.
- View long lists in two automatic columns, switch columns with Tab, and page
  through both together; short lists and narrow windows keep one column.
- Install from portable archives for Linux x64/arm64 and macOS Intel/Apple Silicon.

### Changed

- Rename the command, shell function, executables, archives, and Go modules to
  `cdd`, using `github.com/nik-makarov/cdd` as the repository path.
- Use dim theme-aware text for files so directory entries remain more prominent.
- Start in a three-pane tree view with an active middle list, parent/child
  context, colored files, and distinctly marked directory symlinks. Press `v`
  to switch to the previous compact directory-only list.
- Reduce executable size with a small compatible argument parser and simpler
  message formatting, preserving error causes and shell integration.

### Fixed

- Pass strict Homebrew package validation by inferring the version from the
  source archive URL.
- Relay informational output through shell integration for every supported flag
  spelling, including `--version=true`, instead of attempting to `cd` into it.
- Clear Children when the current-directory (`.`) row is highlighted, including
  after entering a folder, instead of duplicating the Current listing.

### Notes

- Shell integration is required to execute `cd` in the current shell.
- The application stays focused on choosing directories, with two runtime
  dependencies maintained within the Go project.
