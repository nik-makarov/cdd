# cdd

Browse directories in your terminal, choose one, and `cd` into it in your current
shell. A small Go CLI for macOS and Linux with zsh and bash integration.

- Keyboard navigation, filtering, hidden directories, and directory symlinks.
- Three-pane browsing with colored files and directory links; `v` switches to
  a compact directory-only list with automatic two-column layout.
- Reads one directory at a time without recursively scanning the filesystem.
- Two application dependencies, both maintained within the Go project.

The scope is directory selection and shell integration. Features must directly
support that task while keeping the program and its dependency graph small.

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

Use Up/Down to move, Right to open a directory, Left to go to its parent, and
Enter to choose the highlighted directory. The first row, `.`, chooses the
directory currently being viewed. Esc cancels; `/` filters names; `?` opens
keyboard help.

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
