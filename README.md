# 0go

Browse directories in your terminal, choose one, and `cd` into it in your current
shell.

0go is a small Go CLI for **macOS and Linux**, with **zsh and bash** integration.
It reads one directory at a time, supports keyboard navigation and filtering,
and depends only on two modules maintained by the Go project.

## Try it without installing

Requires Go **1.27.4 or newer**, Make, and an interactive ANSI-compatible terminal.

```sh
git clone https://github.com/nik-makarov/0go.git
cd 0go
make build
eval "$(./bin/0go --init zsh)"
0go
```

For bash, replace `zsh` with `bash`. This defines a function in the current shell
only; it does not change your shell configuration or require a `PATH` change.
The function records the absolute binary path, so it keeps working after `cd`.

Select a directory with **Enter**, open it with **Right**, or cancel with **Esc**.
The first row, `.`, selects the directory currently being viewed.

```text
0go  /  choose a directory
Path: /home/user/projects

  .  (current directory)
> app/
  notes/
  tools/

3 directories  ·  hidden: off  ·  2/4
↑↓/jk move  →/l open  ←/h parent  Enter select
/ filter  . hidden  ~ home  Ctrl+R refresh  Esc/q cancel
```

## Usage

```sh
0go                         # Start in the current directory
0go ~/projects              # Start in another directory
0go --hidden /tmp           # Include hidden directories
0go -- --directory-name     # A starting path beginning with a dash
0go --help
```

Put options before the starting directory. Only directories and symbolic links
to directories are listed; the filesystem is not scanned recursively.

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
| Ctrl+R | Refresh the directory listing |
| Esc, `q`, Ctrl+C / Ctrl+D | Cancel |

Press `/` again to clear the filter and start a new one. Returning to a parent
directory preserves the selection of the directory you came from. Errors are
shown in the interface, allowing you to continue browsing or cancel.

## Install

From the source checkout:

```sh
make install
```

The default destination is `~/.local/bin/0go`. Use `make install BINDIR=/your/bin`
to choose another directory. Add this line to `~/.zshrc`:

```sh
eval "$("$HOME/.local/bin/0go" --init zsh)"
```

For bash, add the equivalent line with `--init bash` to `~/.bashrc`. Open a new
shell or run the line directly to activate it. The shell function calls the
binary by its absolute path; adding the install directory to `PATH` is optional.
If you move the binary, regenerate the function from its new location.

You can also install the published repository with Go:

```sh
go install github.com/nik-makarov/0go@latest
eval "$("$(go env GOPATH)/bin/0go" --init zsh)"
```

If you set `GOBIN`, use that directory instead of `$(go env GOPATH)/bin`.
Shell integration is still required for changing the current shell's directory.

## How shell integration works

A child process cannot change its parent shell's working directory. The generated
`0go` function runs the binary, captures the chosen absolute path, and calls
`builtin cd -- "$path"` in the current shell.

The path is passed as a literal argument, never evaluated as shell code.
Whitespace, quotes, shell metacharacters, and trailing newlines in directory
names are preserved. Cancellation or an error leaves the current directory
unchanged.

The binary alone outputs the chosen absolute path **without a trailing newline**.
The interface uses terminal stdin/stderr, falling back to `/dev/tty` when needed.
Help and errors go to stderr. `--help`, `--version`, and `--init` also work without
an interactive terminal.

| Exit code | Meaning |
| --- | --- |
| `0` | Directory selected, or an informational command succeeded |
| `1` | Runtime error |
| `2` | Invalid arguments |
| `130` | Selection cancelled |

## Dependencies

The application has exactly two dependencies outside the standard library:

| Module | Purpose | Upstream |
| --- | --- | --- |
| `golang.org/x/term` | Raw mode, terminal size, and terminal restoration | [Go term repository](https://go.googlesource.com/term/) |
| `golang.org/x/sys` | Waiting for input with `poll`; also used by `x/term` | [Go sys repository](https://go.googlesource.com/sys/) |

Both are maintained within the Go project. Versions are pinned in [go.mod](go.mod)
and checksums in [go.sum](go.sum); there are no additional transitive modules.
New dependencies require a clear need and review of their maintainers, license,
and complete dependency graph.

To avoid an extra Unicode-width dependency, non-ASCII characters are
conservatively counted as two terminal cells when truncating lines. Some names
may leave unused space at the right edge.

## Development and checks

Run these commands from the project directory:

```sh
make fmt          # Format Go source
make check        # Verify formatting, run go vet and race-enabled tests
make build        # Build bin/0go
make audit        # Verify checksums and check known vulnerabilities
```

From a parent directory, use `make -C 0go audit`. `audit` is a Make target, not
a subcommand of the application.

`make audit` runs the pinned Go vulnerability scanner with `go run`. It needs
network access to download the tool on first use and to query `vuln.go.dev`.
It does not require a separate installation or add the scanner's dependencies
to the application's module files.

Tests cover navigation, filtering, symlinks, cancellation, malformed keyboard
input, and real zsh/bash integration with unusual directory names. Install both
shells to run all shell tests; an unavailable shell is otherwise skipped.
CI checks Linux and macOS and runs a vulnerability audit weekly as well as on
pushes and pull requests.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines and
[SECURITY.md](SECURITY.md) for vulnerability reporting.

## License

[MIT](LICENSE), copyright 2026 Nikita Makarov. Go dependencies retain their
respective BSD-style licenses; see [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
