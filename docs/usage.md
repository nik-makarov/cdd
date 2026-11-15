# Usage

## Run without installing

From the source checkout:

```sh
make build
eval "$(./bin/cdd --init zsh)"
cdd
```

Use `--init bash` for bash. This defines a function for the current shell only.
It does not edit shell configuration or require adding a directory to `PATH`.
The function stores the binary's absolute path, so moving the binary requires
running its `--init` command again.

Starting the binary directly in the terminal without the shell function exits
with a setup message before opening the picker. The message repeats the README's
Homebrew setup commands:
append the initialization line to `~/.zshrc` with `echo`, then run `source ~/.zshrc`.
Adjust the binary path for your installation, use `~/.bashrc` and `--init bash`
for bash, and run the append command only once. Then close and reopen your terminal.
For zsh with a custom `ZDOTDIR`, use `.zshrc` in that directory instead. `--help`,
`--version`, and `--init` remain available before integration is loaded.
On terminal stderr, the setup message highlights its heading in red and commands
in cyan. `NO_COLOR` removes colors while retaining bold emphasis; `TERM=dumb`
and redirected stderr produce plain text.

## Options

```sh
cdd                         # Start in the current directory
cdd ~/projects              # Start in another directory
cdd --hidden /tmp           # Include hidden directories initially
cdd -- --directory-name     # A starting path beginning with a dash
cdd --help
cdd --version
```

Place options before the starting directory. The default tree view lists files,
directories, and symbolic links; only directories and links to directories can
be chosen as a destination. Enter on a file opens its default desktop application.

## Controls

Enter on a highlighted file invokes `open` on macOS or `xdg-open` on Linux and
keeps cdd open at the same location. Enter also opens a file from the built-in
preview. In panes view, Enter while editing a filter finishes input first;
in tree view it acts on the highlighted result. Modified Enter and Change Dir
remain directory-only. Missing files or application-launch failures appear in
the terminal; file paths are never returned to shell integration.

Space on a file opens Quick Look on macOS, or the built-in text preview on Linux.
The built-in viewer supports arrows, mouse wheel, Page Up/Down, Home/End, and
closes with Esc or Space, restoring browser focus. It reads at most 256 KiB and
shows a size placeholder for binary files; special files such as FIFOs are not
read. Close the native Quick Look window to return to browsing; another Space
in cdd stops the active preview. Neither action chooses a shell directory.

| Key | Action |
| --- | --- |
| Up / Down, `k` / `j` | Move selection |
| Enter | Open a file; in tree, choose a directory; in panes, confirm `.` or browse another directory |
| Space | Preview a file: Quick Look on macOS, built-in viewer on Linux |
| Configured editor key (default `z`) | Focus/open the highlighted folder as a project; open a file by itself in a new editor window |
| Right, `l` | Open the highlighted directory |
| Left, `h`, Backspace | Focus `..` first; from `..` go to the parent directory |
| Cmd+Enter (macOS) / Ctrl+Enter (Linux) | Choose the highlighted directory and exit |
| Home / End | Jump to the first / last row |
| Page Up / Page Down | Move by one page |
| `v` | Switch between expandable tree (default) and panes |
| `/` | Filter names, ignoring case |
| Enter / Esc while filtering | In panes, finish editing; in tree, Enter chooses the highlighted directory and Esc finishes editing |
| `n` | Enter a name for a new directory in the current location |
| Shift+N | Enter a name for an empty file inside the highlighted directory (or beside a highlighted file) |
| `r` / Rename button | Rename the highlighted file or dir; Enter saves, Esc cancels |
| Cmd+Backspace (macOS) / Ctrl+Backspace (Linux) / Delete button | Move to Trash after exact name confirmation |
| Enter / Esc while naming | Create the directory or file / cancel name input |
| `.` | Toggle hidden directories |
| `~` | Go to the home directory |
| Ctrl+R | Refresh the listing |
| `?` | Show or hide controls help; Esc also closes help |
| Esc, `q`, Ctrl+C / Ctrl+D | Cancel |

The editor key and separate dir/file commands come from
[configuration](configuration.md). Both defaults try Zed's `zed`/`zeditor` CLI,
then VS Code's `code`. Folders reuse an open project or open a new project window;
individual files open in new windows. The key is ignored when the applicable
command is disabled or none of its CLIs is installed, and stays literal during
text input. The file command also works from the built-in preview. File and
editor actions keep cdd open.

In terminals that support mouse reporting:

| Mouse action | Result |
| --- | --- |
| Left click | Highlight the entry under the pointer |
| Double left click | Open the directory for browsing; `..` opens the parent; `.` confirms the current directory and exits |
| Click `Change Dir` | Choose the highlighted directory and exit |
| Wheel up / down | Move selection one entry up / down |

Mouse actions work in the middle pane of panes view and the expandable tree.
Header paths navigate through their visible components. Side previews, dividers,
and empty rows do not select anything. Tree-specific controls are described below.
Double-click the same entry and row within 400 ms; up to two cells of horizontal
pointer drift are allowed. On `.`, this confirms the current directory and exits;
other directories open for browsing.
Right-click has no action. Files cannot be chosen as cd destinations. While
filtering, double-clicking `.` also confirms the current directory; double-clicking
another directory opens it and clears the filter. While entering
a new directory name, mouse actions are ignored. Mouse reporting is disabled
when the picker exits, including cancellation.

The centered button above the keyboard shortcuts uses a thin rounded outline
instead of a filled block. Its title and platform shortcut share one line; the
destination appears below with the final directory name emphasized. All four rows,
including the border, are clickable. Long paths are shortened from the left.
The shortcut is Enter for every directory in tree view. In panes views, it
changes to Enter on the current-directory row; other directories show Cmd+Enter
on macOS or Ctrl+Enter on Linux.
The outline and bold title remain visible with `NO_COLOR`. For files it appears
dim, shows no shortcut, and asks you to choose a directory. The short controls
also omit opening and confirmation hints for files. The button is hidden while editing a
filter or directory name. While filtering in tree view, Enter or modified Enter
chooses the highlighted directory and Esc finishes editing. In panes filters,
Enter finishes input and modified Enter has no action. While naming a new
directory, Enter creates it and modified Enter has no action.
Small windows omit the entire button if it cannot fit beside the selected row
and any input/error message.

Modified Enter and Backspace require a terminal that supports the
[kitty keyboard protocol](https://sw.kovidgoyal.net/kitty/keyboard-protocol/).
cdd enables disambiguated keys for the picker and restores the previous mode on
exit. A terminal or desktop may intercept the shortcut; in that case free the
binding in its settings or click the button. Legacy terminals may send the same
bytes for plain and modified Enter; cdd treats those bytes as plain Enter,
confirming `.` or opening another directory for browsing.

In panes views, the first row, `← Back to parent`, is a navigation action for `..`.
Its arrow sits in the left gutter, with the action text aligned to directory names.
The next row, `.  (current dir)`, selects the
directory currently being viewed and is highlighted initially. Enter or a
double-click on `.` confirms that directory and exits, applying the shell directory
change. On other directory rows, Enter opens them for further browsing, like Right
or double-click.
Opening `..` moves up and focuses the directory you came from, while the
confirmation shortcut or button selects the parent as the destination.
Both rows remain available regardless of the name filter or hidden-entry setting;
the filesystem root omits `..`. Left, `h`, or Backspace first
focuses `..` in panes views without leaving the directory or clearing a completed filter.
Pressing it again, Enter, or Right opens the parent, clears the filter, and
focuses the directory you came from. At the filesystem root, Left does nothing.
While editing a filter, Backspace still deletes text.

Press `/` again to clear the current filter. In panes, the `.` and `..` rows remain available
even when no child names match. While editing a filter, the first Enter or Esc finishes
editing; a subsequent Enter confirms `.` or opens another highlighted directory,
and a subsequent Esc cancels.
While editing a filter, `?` is a literal part of the query. Opening a filter
closes controls help. Navigation stays available while help is expanded; the
first Esc closes help without cancelling selection.

Directory read errors appear inside the interface without changing the current
browsing location. Continue navigating or cancel to return to your shell.

## Create a directory

Press `n`, type a name, and press Enter to create one directory inside the current
location in panes views. In tree view, creation uses the highlighted directory,
or a file's parent; the prompt shows this destination. Backspace edits the name; Esc dismisses
the input without creating anything. While naming, letters such as `q`, `v`, and
`n` are ordinary text. Ctrl+C or Ctrl+D cancels the picker. Finish an active filter
before pressing `n` to start creation.

Names may include spaces, Unicode, and shell metacharacters, which remain literal.
Use a single name: empty names, `.`, `..`, and paths containing `/` are rejected.
Existing files, directories, and symlinks are never overwritten. Creation errors
remain visible while the name stays editable, so you can correct it and retry.

After creation, the list refreshes, the filter clears, and the new directory is
highlighted. A name beginning with `.` enables hidden entries so the new directory
is visible. Right opens it for browsing; use the confirmation shortcut or
`Change Dir` to select it and exit. Enter confirms it in tree view and opens it
for browsing in panes views.
Cancelling the picker after creation leaves the new directory on disk.

## Rename a file or dir

Highlight an entry and press `r` or click Rename above the shortcut hints. The
existing name is prefilled; typing replaces it, and Backspace clears it before
editing. Enter applies one literal name in the same parent dir. Esc cancels.
Existing entries, including broken symlinks, are never overwritten. Errors leave
the name editable. Renaming a symlink changes the link itself, not its target.

Navigation rows and the current/starting dirs and their ancestors are protected.
The prompt keeps the original target fixed and rejects an entry replaced while
input was open. After success, the listing refreshes and selects the renamed
entry; hidden names become visible. Shell directory selection stays unchanged.

## Move a file or dir to Trash

Highlight an entry and press Cmd+Backspace on macOS, Ctrl+Backspace on Linux, or
click Delete above the shortcut hints. The prompt shows the target path and name
with an empty `Confirm name:` field. Type the complete name, including its case
and any leading or trailing spaces, then press Enter. A mismatch leaves the
prompt open and moves nothing. Backspace edits; Esc cancels before confirmation.
Other actions and mouse navigation do not change the target while the prompt is
open. Finish any active filter or naming input before using this shortcut.

The entry goes to the system Trash and can be restored there. A folder moves
with its contents; a symlink moves as a link without affecting its target.
macOS uses the system file manager API. Linux requires the `gio` CLI from GLib
and a location that supports Trash. Missing support, permissions errors, or
command failures show an error and require fresh name confirmation. cdd never
falls back to permanent deletion.

Navigation rows, current/starting dirs, and their ancestors are protected. An
entry replaced while the prompt was open is rejected. Names containing terminal
control characters or invalid UTF-8 cannot be confirmed. After success, the
listing refreshes and tree focus returns to the parent. Cancelling cdd after
confirmation does not undo a move that has already completed; restore it from
Trash instead. Plain Backspace keeps its navigation and text-editing behavior.

## Create an empty file

Highlight a directory and press Shift+N, type a file name, then press Enter. Both
views create inside the highlighted directory; when a file is highlighted, its
parent is used. `New in:` shows the destination and `New file:` shows the name.
Backspace edits it and Esc cancels without creating anything. Finish an active
filter before invoking the shortcut; letters remain literal during text input.

Creation uses one literal name and never overwrites an existing file, directory,
or symlink, including a broken symlink. Empty names, `.`, `..`, `/`, and NUL are
rejected. Errors keep the name editable for retry. The new file is empty, has
permissions `0644` restricted by the process umask, and is highlighted after
refresh. Hidden names become visible. In panes view, browsing enters the target
folder to show the new file; shell selection remains unchanged.

Space previews the new file, Enter opens it, and the configured editor shortcut
opens it in the editor. It cannot be selected as a cd destination. Cancelling
the picker after creation leaves the file on disk.

## Appearance

### Panes view (second)

The middle pane is always active and lists the current directory's folders and
files. The left pane shows its parent's entries as context. The right pane
previews entries inside the highlighted directory, including directory symlinks.
Highlighting the `.` or `..` row leaves Children blank, including after entering
a directory.
Highlighting a file shows no directory preview. File contents are read only after
an explicit Space preview action.

Each pane shows its directory's absolute path above its entries: the parent on
the left, the current directory in the middle, and the highlighted directory
on the right. The right path is blank for files and the `.` row; the parent
path is blank at the filesystem root. Long paths are shortened from the left
within their own column, preserving the directory name when space permits.

Folders are cyan, files use the theme's dim foreground, and directory symlinks
are magenta. Folder names have no trailing slash. Symlink markers remain visible
without color: `@` for directory and file links, and `@!` for broken or inaccessible
links. Enter opens a file in its default application; files cannot be cd destinations.
Preview read errors stay in their side pane without preventing navigation in
the middle.

Pane labels are dim, including `Current` without brackets. Both side panes use
dim text while retaining entry-type colors and markers; the current folder in
the parent pane is not bold. The active address uses the normal foreground for
the parent path and bold cyan for its final directory name, so the address and
selected row stand out from the surrounding context.

The center has more width than either side pane. Below 59 terminal cells, the
side panes are hidden while the active list and its file display remain.
Expanding the terminal restores the panes without changing the selected entry.

### Expandable tree view (default)

Tree is the default when no saved preference exists. A single column shows nested folders with guide
lines, `▸` for a collapsed branch, and `▾` for an expanded branch. The root is your
home directory when it is an ancestor of the browsing location; otherwise it is
`/`. Ancestors expand automatically to reveal and focus the browsing location.
The tree contains real directory nodes, without synthetic `.` or parent rows.
Files stay visible and Enter opens their default application. They cannot be
selected as cd destinations. Empty folders
lose the disclosure arrow after
being checked and display `(empty)`; folders with only hidden entries say so.

| Action | Result |
| --- | --- |
| Up / Down, `k` / `j` | Move through visible rows without changing expansion |
| Right / `l` | Expand a collapsed folder; on an expanded folder, focus its first child |
| Left / `h` / Backspace | Collapse an expanded folder; otherwise focus its parent |
| Enter | Choose the highlighted directory and exit, including while filtering |
| Click a name | Highlight that row |
| Click a disclosure arrow | Expand or collapse the branch |
| Double-click a name | Expand or collapse the branch |
| Home / End, Page Up / Down | Move to the first/last visible row or by one screen |
| Tab / Shift+Tab | Keep the single tree column focused |

Left on a collapsed root stays at that root. The Change Dir button shows Enter
for directories; the platform confirmation shortcut continues to work too.

The directory from which the shell launched cdd uses the theme's dim foreground, with a
`(starting dir)` label, independently of focus and view changes. The label and
dimming remain with `NO_COLOR`. This identifies the destination that would
leave the shell in the same place; it remains selectable. The `From:` header keeps
the launch path visible even if it is outside the tree or its branch is collapsed.
An explicit starting-path argument changes the browsing location, not this marker.

Apart from revealing ancestor chains, children are read only when a branch is
first expanded, then cached. Collapsing a branch retains its nested expansion
state. Ctrl+R refreshes expanded branches;
if a selected entry disappears, focus returns to its nearest visible ancestor.
Directory symlinks keep their logical paths. Even a link back to an ancestor loads
only one level per explicit expansion, without recursively walking the filesystem.

The `/` filter searches currently expanded branches and retains ancestors of
matching entries, the root, and the starting directory when its branch is visible.
It does not search unopened folders. Esc finishes input and retains the filter;
Enter chooses the highlighted directory. Expanding, collapsing, or moving to a
parent clears the filter. The footer counts displayed entries, excluding the root
and including retained ancestors,
rather than claiming a recursive match count. The hidden-entry toggle applies at
every loaded level; hidden ancestors needed to reveal the browsing and starting
locations remain visible without revealing unrelated hidden folders.

Switching views preserves folder focus. For a nested selection, the flat view
opens its containing directory and retains the highlighted file or folder.
Switching without moving the root preserves cached expansion state within a run.
The selected view is remembered between runs.

### Remembered view

In tree view, click `[/ Root]` in the header or press `\` to reveal the tree
from `/` while retaining the highlighted path. This root remains active while
browsing and switching views during the current session. `~` still navigates
home. Root selection does not change the shell directory.

Every successful `v` switch saves the selected mode immediately, including if you
later cancel with Esc or Ctrl+C. The next run restores that view at its new
starting location. Missing, unreadable, empty, or unrecognized saved values use
tree. If a saved tree cannot load its ancestors, cdd keeps panes and reports the
error. A save failure reports an error while allowing navigation and selection.

The preference stores only `panes` or `tree` in `cdd/last-view` beneath
the operating system's user configuration directory:

- macOS: `~/Library/Application Support/cdd/last-view`.
- Linux: `$XDG_CONFIG_HOME/cdd/last-view`, or `~/.config/cdd/last-view` by default.

The removed `list` preference is treated as unknown and opens tree view.
Remove that file to reset the startup view. Paths, filters, hidden-entry settings,
and expanded branches are not saved. If no user configuration directory can be
determined, cdd runs with the default view without persisting switches.

### Shared styling and controls

Click any visible folder name in a header path to browse there without confirming
a shell directory change. In tree view the path follows the highlighted directory
(or a file's parent); this also exposes ancestors above the initial home root.
The `/` component navigates to the filesystem root. Clipped ellipses are inert.

The path highlights the directory currently being viewed; long paths are
shortened from the left. A reverse-video bar and a `>` marker identify the
selected row. Errors use the terminal's red and secondary labels use dim text.
The interface uses the terminal's own
background, foreground, and palette instead of fixed RGB colors or custom fonts.

The footer keeps essential controls visible, with more shortcuts available via
`?`. Counts exclude the `.` and `..` rows. Filtered lists show a match count;
long lists indicate when more rows are above or below the visible page.

Set a non-empty `NO_COLOR` value to disable added colors:

```sh
NO_COLOR=1 cdd
```

Bold, dim text, and reverse selection remain available in this mode. `TERM=dumb`
disables all text styling, retaining the `>` marker. Cursor movement and screen
redrawing still require an ANSI-compatible interactive terminal. The selected
path on stdout never includes styling, including when shell integration captures
it through command substitution.

## Installation

Prebuilt archives target Linux amd64/arm64 and macOS amd64/arm64. They can be used
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
| `LICENSEDIR` | `$(PREFIX)/share/licenses/cdd` | Installed license directory |

For example, `make install PREFIX=/your/prefix` installs the executable and
licenses beneath that prefix. `BINDIR` remains independently configurable.
The invoking user needs write access to the selected directories.

For persistent zsh integration, add this line to `~/.zshrc`:

```sh
eval "$("$HOME/.local/bin/cdd" --init zsh)"
```

For bash, use `--init bash` in `~/.bashrc`. Adjust the binary path for a custom
prefix. Open a new shell or run the line directly. The function invokes the
absolute binary path, so adding the installation directory to `PATH` is optional.

The published source repository can also be installed with Go:

```sh
go install github.com/nik-makarov/cdd@latest
eval "$("$(go env GOPATH)/bin/cdd" --init zsh)"
```

If `GOBIN` is set, use that directory instead of `$(go env GOPATH)/bin`.
`go install` does not generate redistribution notices; use `make build` or
`make notices` from a source checkout before redistributing compiled binaries.

## Shell integration contract

A child process cannot change the parent shell's working directory. `--init`
prints a function named `cdd` that runs the binary, captures its selected absolute
path, and executes `builtin cd -- "$path"` in the current shell.

The function sets `CDD_SHELL_INTEGRATION=1` only for the picker process. Older
functions remain supported because they capture stdout through a pipe, even
without the marker. The binary also allows explicit path capture through pipes;
this does not prove that a function is loaded or change the parent directory
automatically. Merely having zsh installed or setting `SHELL` does not enable
integration. Do not export the marker in your shell configuration; load the
generated function instead.

The initial `eval` defines that function. Selected paths never pass through
`eval`: they are literal arguments, preserving whitespace, quotes, shell
metacharacters, and trailing newlines. Cancellation or failure before selection
leaves the current working directory unchanged. The shell reports a failed `cd`,
for example if a directory is removed after selection.

After a successful directory change, the function prints the original and
resulting shell directories to stderr:

```text
cdd: /old/path -> /new/path
```

Paths are shell-escaped so spaces, newlines, and terminal control characters
cannot make the notification ambiguous or inject terminal commands. Cancellation,
picker failure, failed `cd`, and selecting the current directory do not print a
move notification. A failed `cd` still reports the shell's own error.
After updating the binary, rerun its `--init` command through `eval` to refresh
an existing function; replacing the binary alone does not update shell code.

The binary outputs only the chosen absolute path, without a trailing newline.
The shell function sends only absolute path output to `cd` and relays other
successful stdout unchanged. It does not parse flags independently, so aliases,
boolean values such as `--version=true`, and `--init` output follow the binary's
argument parser without accidentally becoming directory names.
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

`make build` creates `bin/cdd`, `bin/LICENSE`, and
`bin/THIRD_PARTY_LICENSES.txt`. `make cross-build` creates binaries for Linux
amd64/arm64 and macOS amd64/arm64, sharing the same license files. Keep both license
files with any redistributed executable.

`make dist` packages four local preview archives under `dist/dev/`, including
release notes, licenses, archive instructions, and SHA-256 checksums. Versioned
packaging requires a matching dated section in CHANGELOG.md; see
[releases](releases.md).

Running plain `make` prints a command summary without building or installing.

`make clean` removes those generated files. Build output is excluded from Git.
