# Changelog

User-facing release notes. Changes below are not released until a matching
version is published on GitHub. Development history is recorded separately in
[the development journal](docs/development-journal.md).

## Unreleased

### Added

- Browse directories with arrow keys or `hjkl`, then choose one with Enter.
- Filter directory names, show hidden directories, and follow directory symlinks.
- Change the current shell's directory through zsh or bash integration.
- Cancel selection without changing the current working directory.
- Install from portable archives for Linux x64/arm64 and macOS arm64.

### Notes

- Shell integration is required to execute `cd` in the current shell.
- The application stays focused on choosing directories, with two runtime
  dependencies maintained within the Go project.
