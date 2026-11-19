# cdd

Browse directories in your terminal, choose one, and `cd` into it in your current
shell. A small Go CLI for macOS and Linux with zsh and bash integration.

- Keyboard and mouse navigation, filtering, hidden directories, and directory symlinks.
- Create a directory with `n`, then choose it.
- Create an empty named file in the highlighted folder with Shift+N.
- Rename a file or folder with `r` or the Rename button.
- Open the highlighted folder in Finder with `f` or the Finder button on macOS.
- Copy the highlighted file or folder's absolute path with `y` or Copy Path.
- On macOS, copy a file or folder with Cmd+C and paste with Cmd+V. Review the
  destination path and press Enter to confirm. The clipboard works with Finder.
- Move an entry to Trash with Cmd+Backspace (macOS), Ctrl+Backspace (Linux), or
  the Delete button, after typing its exact name. Linux requires `gio` from GLib.
- Preview a file with Space: Quick Look on macOS, a text viewer on Linux.
- Open a file in its default application with Enter while keeping cdd open.
- Open a folder or file in an editor with `z`: Zed, then VS Code. Reuse open
  projects and open files in new windows, with [separate commands](docs/configuration.md).
- Expandable directory tree and three-pane browsing with colored files and
  directory links; `v` switches between the two views.
- Browsing loads one directory at a time.
- Two application dependencies, both maintained within the Go project.

The scope is choosing or creating a destination directory and shell integration.
Features must directly support that task while keeping the program and its
dependency graph small.

## Install with Homebrew

Homebrew is the recommended installation method. For zsh on Apple Silicon macOS:

```sh
brew install nik-makarov/tap/cdd
echo 'eval "$(/opt/homebrew/bin/cdd --init zsh)"' >> ~/.zshrc
source ~/.zshrc
cdd
```

The configuration line enables cdd in future zsh sessions; `source` activates it
in the current terminal. Run the `echo` command once. Homebrew installs the build
dependencies automatically.

On Intel macOS or Linux, replace `/opt/homebrew` with the output of
`brew --prefix`. For bash, use `--init bash` in `~/.bashrc`. See
[installation and shell integration](docs/usage.md#installation) for details.

## Usage

Use an interactive ANSI-compatible terminal.

Press `v` to cycle through **tree → panes**. The last used view is restored
on startup; tree is the default when no valid preference exists.
In tree view, Right expands a branch or focuses its first child; Left collapses
it or focuses its parent. Click its arrow or double-click its name to toggle it.
Enter chooses the highlighted directory, including while filtering. The tree
starts at home when it is an ancestor, otherwise `/`, and reveals the browsing
location. Your launch directory stays dim and marked `(starting dir)`.
The tree loads folders as you expand them; `/` filters expanded branches.
Press Tab to switch to your last 50 distinct destinations, newest first, and
Enter to return to one. The list sits beside the tree when every expanded row
fits and 30 columns remain for history. Otherwise Tab swaps full-width panels.

In panes views, use Up/Down to move and Enter or Right to open a directory. Left
first focuses `← Back to parent`; pressing Left again opens the parent.
Enter on `.` confirms the directory currently being viewed and exits. Esc cancels;
`/` filters names; `n` creates a directory; `?` opens controls help.
In mouse-compatible terminals, click to highlight and double-click to open a
directory, including `..` to go up; double-click `.` to confirm the current directory.
In a compatible terminal, Cmd+Enter on macOS or Ctrl+Enter on Linux chooses
the highlighted directory from either view. The mouse wheel moves selection.

```sh
cdd ~/projects
cdd --hidden /tmp
cdd --help
```

The binary cannot change its parent shell's directory on its own. The generated
shell function receives the chosen path and executes `cd` in the current shell.
See [usage and shell integration](docs/usage.md) for all controls and options.

## Other installation methods

### Prebuilt archives

Prebuilt archives are prepared for Linux x64/arm64 and macOS Intel/Apple Silicon.
Download the matching archive from
[GitHub Releases](https://github.com/nik-makarov/cdd/releases). Each archive includes
the executable, installation instructions, licenses, and release notes. See
[release installation](docs/releases.md#use-an-archive).

### Build from source

Use the latest stable Go release matching the minimum in [go.mod](go.mod) and Make.

```sh
git clone https://github.com/nik-makarov/cdd.git
cd cdd
make install
eval "$("$HOME/.local/bin/cdd" --init zsh)"
cdd
```

The default binary location is `~/.local/bin/cdd`; license files are installed
under `~/.local/share/licenses/cdd`. Set `PREFIX` to choose another installation
prefix. Add the `eval` line to `~/.zshrc` for future shells, or use `--init bash`
in `~/.bashrc`. Full installation and `go install` instructions are in
[usage](docs/usage.md#installation).

## Verification

```sh
make check
make audit
make cross-build
```

`make check` verifies formatting, runs vet and race-enabled tests for the
application and license generator. `make audit` uses the pinned vulnerability
scanner from [tools/](tools/README.md); no manual installation is needed.
The audit needs access to the Go vulnerability database.

`make build` and `make cross-build` generate `bin/THIRD_PARTY_LICENSES.txt` and
`bin/LICENSE` alongside the binaries. Retain both files when redistributing a
standalone executable. Generated output is not committed.

GitHub Actions checks Linux and macOS and runs a weekly vulnerability audit.
Successful pushes to `main` publish a release and update the Homebrew tap,
starting at `v0.1.0` and increasing the patch version automatically. Pushes that
only change Markdown files or `docs/` skip publication.
Hosted CI execution and GitHub security settings require a published repository;
local validation does not establish that they have run. See
[testing](docs/testing.md) for verification commands, evidence, and limits.

## Documentation

- [Update Go and dependencies with a button](docs/maintenance.md)
- [Release notes](CHANGELOG.md) and [release packaging](docs/releases.md)
- [Usage, installation, and shell integration](docs/usage.md)
- [Architecture and dependency boundaries](docs/architecture.md)
- [Verification and manual acceptance](docs/testing.md)
- [Pinned development tools](tools/README.md)
- [Contributing](CONTRIBUTING.md), [security reporting](SECURITY.md), and [development journal](docs/development-journal.md)
- [Planned work](TODO.md)
- [Redistribution notices](THIRD_PARTY_NOTICES.md)

## License

MIT. See [LICENSE](LICENSE). Dependency and Go standard library licenses remain
separate and are included in the generated redistribution notices.
