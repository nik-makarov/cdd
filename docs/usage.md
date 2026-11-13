# Usage

## Run without installing

From the source checkout:

```sh
make build
eval "$(./bin/0go --init zsh)"
0go
```

Use `--init bash` for bash. This defines a function for the current shell only.
It does not edit shell configuration or require adding a directory to `PATH`.
The function stores the binary's absolute path, so moving the binary requires
running its `--init` command again.

## Options

```sh
0go                         # Start in the current directory
0go ~/projects              # Start in another directory
0go --hidden /tmp           # Include hidden directories initially
0go -- --directory-name     # A starting path beginning with a dash
0go --help
0go --version
```

Place options before the starting directory. Only directories and symbolic links
to directories are listed. Broken symlinks and links to regular files are omitted.

## Controls

| Key | Action |
| --- | --- |
| Up / Down, `k` / `j` | Move selection |
| Right, `l` | Open the highlighted directory |
| Left, `h`, Backspace | Go to the parent directory |
| Enter | Choose the highlighted directory and exit |
| Home / End | Jump to the first / last row |
| Page Up / Page Down | Move by one page |
| `/` | Filter names, ignoring case |
| Enter / Esc while filtering | Finish editing and keep the filter |
| `.` | Toggle hidden directories |
| `~` | Go to the home directory |
| Ctrl+R | Refresh the listing |
| Esc, `q`, Ctrl+C / Ctrl+D | Cancel |

The first row, `.`, selects the directory currently being viewed. Enter selects
a directory; Right opens it for further browsing. When returning to a parent,
the cursor stays on the directory you came from.

Press `/` again to clear the current filter. The `.` row remains available even
when no child names match. While editing a filter, the first Enter or Esc finishes
editing; a subsequent Enter selects, and a subsequent Esc cancels.

Directory read errors appear inside the interface without changing the current
browsing location. Continue navigating or cancel to return to your shell.

## Installation

Prebuilt archives target Linux amd64/arm64 and macOS arm64. They can be used
without Go or Make; see [archive installation](releases.md#use-an-archive).
The commands below build and install from a source checkout.

```sh
make install
```

Defaults:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PREFIX` | `$HOME/.local` | Installation prefix |
| `BINDIR` | `$(PREFIX)/bin` | Executable directory |
| `LICENSEDIR` | `$(PREFIX)/share/licenses/0go` | Installed license directory |

For example, `make install PREFIX=/your/prefix` installs the executable and
licenses beneath that prefix. `BINDIR` remains independently configurable.
The invoking user needs write access to the selected directories.

For persistent zsh integration, add this line to `~/.zshrc`:

```sh
eval "$("$HOME/.local/bin/0go" --init zsh)"
```

For bash, use `--init bash` in `~/.bashrc`. Adjust the binary path for a custom
prefix. Open a new shell or run the line directly. The function invokes the
absolute binary path, so adding the installation directory to `PATH` is optional.

The published source repository can also be installed with Go:

```sh
go install github.com/nik-makarov/0go@latest
eval "$("$(go env GOPATH)/bin/0go" --init zsh)"
```

If `GOBIN` is set, use that directory instead of `$(go env GOPATH)/bin`.
`go install` does not generate redistribution notices; use `make build` or
`make notices` from a source checkout before redistributing compiled binaries.

## Shell integration contract

A child process cannot change the parent shell's working directory. `--init`
prints a function named `0go` that runs the binary, captures its selected absolute
path, and executes `builtin cd -- "$path"` in the current shell.

The initial `eval` defines that function. Selected paths never pass through
`eval`: they are literal arguments, preserving whitespace, quotes, shell
metacharacters, and trailing newlines. Cancellation or failure before selection
leaves the current working directory unchanged. The shell reports a failed `cd`,
for example if a directory is removed after selection.

The binary outputs only the chosen absolute path, without a trailing newline.
The UI uses terminal stdin/stderr, falling back to `/dev/tty` when needed.
Help and errors go to stderr. `--help`, `--version`, and `--init` work without
an interactive terminal.

| Exit code | Meaning |
| --- | --- |
| `0` | Directory selected, or an informational command succeeded |
| `1` | Runtime error |
| `2` | Invalid arguments |
| `130` | Selection cancelled |

## Build output

`make build` creates `bin/0go`, `bin/LICENSE`, and
`bin/THIRD_PARTY_LICENSES.txt`. `make cross-build` creates binaries for Linux
amd64/arm64 and macOS arm64, sharing the same license files. Keep both license files
with any redistributed executable.

`make dist` packages local previews under `dist/dev/`, including release notes,
licenses, archive instructions, and SHA-256 checksums. Versioned packaging requires
a matching dated section in CHANGELOG.md; see [releases](releases.md).

Running plain `make` prints a command summary without building or installing.

`make clean` removes those generated files. Build output is excluded from Git.
