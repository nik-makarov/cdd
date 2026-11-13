# Changelog

User-facing release notes. Changes below are not released until a matching
version is published on GitHub. Development history is recorded separately in
[the development journal](docs/development-journal.md).

## Unreleased

### Added

- Browse directories with arrow keys or `hjkl`, then choose one with Enter.
- Filter directory names, show hidden directories, and follow directory symlinks.
- Change the current shell's directory through zsh or bash integration.
- Report the original and destination paths after a successful shell directory
  change, with safely escaped paths on stderr.
- Cancel selection without changing the current working directory.
- Use a theme-aware directory picker with a full-row selection bar, highlighted
  current path, compact controls, and expandable keyboard help via `?`.
- Disable added UI colors with `NO_COLOR` while retaining a visible selection.
- Install from portable archives for Linux x64/arm64 and macOS Intel/Apple Silicon.

### Changed

- Reduce executable size with a small compatible argument parser and simpler
  message formatting, preserving error causes and shell integration.

### Notes

- Shell integration is required to execute `cd` in the current shell.
- The application stays focused on choosing directories, with two runtime
  dependencies maintained within the Go project.
