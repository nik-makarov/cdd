# Changelog

User-facing release notes. Changes below are not released until a matching
version is published on GitHub. Development history is recorded separately in
[the development journal](docs/development-journal.md).

## Unreleased

### Added

- Browse directories with the mouse: click to highlight, double-click to open,
  and use the wheel to move through either view. The `← Back to parent` row above the current
  directory opens the parent. Double-clicks tolerate slight horizontal pointer drift.
- Confirm the highlighted directory with a centered, outlined `Change Dir`
  button above the keyboard shortcuts, showing its destination and Cmd+Enter
  (macOS) or Ctrl+Enter (Linux) shortcut. Modified Enter uses the kitty keyboard
  protocol in compatible terminals.
- Create a directory in the current location with `n`, edit its name, and focus
  it after creation; cancel name input with Esc or correct creation errors.
- Show each column's directory path in tree and compact views, with clipped
  paths that follow navigation and the highlighted directory preview.
- Install through the source-built `nik-makarov/tap` Homebrew tap, with automatic
  versioned releases and formula updates after successful pushes to main.
- Browse directories with Enter, arrow keys, or `hjkl`, then confirm the destination
  with the `Change Dir` button or its shortcut.
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

- Shorten interface wording to dir/dirs and label the current row `(current dir)`.
- Confirm the current directory and exit with Enter on the `.` row; Enter on
  other directories continues browsing. Show Enter on the confirmation button
  when the current-directory row is highlighted; hide the shortcut and opening
  hints for files.
- Focus the emphasized `Back to parent` row first with Left, `h`, or Backspace;
  open the parent when that row is already highlighted.
- Make the current directory's address more prominent with a bold final name,
  subdued pane labels, and dim side entries that retain their type colors.
- Skip automatic releases and tap updates for pushes that only change Markdown
  files or docs/; include those changes in the next release.
- Rename the command, shell function, executables, archives, and Go modules to
  `cdd`, using `github.com/nik-makarov/cdd` as the repository path.
- Use dim theme-aware text for files so directory entries remain more prominent.
- Start in a three-pane tree view with an active middle list, parent/child
  context, colored files, and distinctly marked directory symlinks. Press `v`
  to switch to the previous compact directory-only list.
- Reduce executable size with a small compatible argument parser and simpler
  message formatting, preserving error causes and shell integration.

### Fixed

- Keep previously loaded zsh/bash functions working after a binary update, even
  when they do not set the newer shell-integration marker.
- Detect missing shell integration before opening the picker and explain how to
  load the zsh or bash initialization script and reopen the terminal, with
  highlighted setup instructions in interactive terminals.
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
