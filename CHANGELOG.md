# Changelog

User-facing release notes. Changes below are not released until a matching
version is published on GitHub. Development history is recorded separately in
[the development journal](docs/development-journal.md).

## Unreleased

### Added

- Revisit the last 50 distinct directory selections from Recent places in
  tree view. Tab switches panels. Enter or a double-click highlights the
  directory in the tree so you can browse, run an action, or confirm with Enter.
  The list sits to the right when the expanded tree fits in full. Otherwise
  Tab switches between full-width panels. Addresses keep their home or filesystem root prefix.
- Copy and paste a file or folder between cdd and Finder on macOS with Cmd+C
  and Cmd+V or the Copy and Paste buttons. Confirm the full destination path
  before writing. A name conflict offers a numbered copy when needed.
- Copy the highlighted file, folder, or symlink's absolute path using `y` or
  Copy Path in either view. Show confirmation after copying; Linux requires
  `wl-copy` for Wayland or `xclip` for X11.
- Open the highlighted dir in Finder on macOS using `f` or the Finder button,
  including current/parent dirs and dir links, while retaining the cdd selection.
- Move a file, folder, or symlink to the system Trash using Cmd+Backspace on
  macOS, Ctrl+Backspace on Linux, or the Delete button. Require its exact name
  in a blank confirmation field; refuse mismatches and never permanently delete
  when Trash is unavailable. Linux uses `gio` from GLib.
- Rename the highlighted file, folder, or symlink with `r` or the Rename button,
  preserving contents and refusing to overwrite existing entries.
- Create an empty named file with Shift+N inside the highlighted folder or beside
  the highlighted file. Focus the new file without opening it or overwriting
  existing entries.
- Open a folder as a project or a single file through an installed editor CLI:
  prefer Zed, then VS Code. Configure separate `editor_dir_command` and
  `editor_file_command` strings and the binding in `keys.open_in_editor` (default:
  `z`). Existing `editor_command` and `keys.open_in_zed` settings still work.
  Ignore the key if no editor is available.
- Open highlighted files in the default application with Enter, preserving
  browsing state and keeping files out of shell directory selection.
- Preview files with Space: native Quick Look on macOS and a bounded, scrollable
  text preview on Linux. Close the built-in preview with Esc or Space.
- Click visible folder components in header paths to navigate without changing
  the shell directory. Tree breadcrumbs follow the highlighted directory.
- Show the tree from `/` using the header's `[/ Root]` button or `\`, retaining
  the highlighted path and allowing navigation above home.
- Remember the last successfully chosen view, including after cancellation.
  Restore it on startup; missing or unknown preferences fall back to tree.
- Add an expandable tree as the default view: cycle tree and panes with
  `v`. Browse nested branches with arrows or the mouse, load folders on demand,
  filter expanded branches with ancestor context, and create a directory inside
  the highlighted folder. Anchor the tree at home or `/`, reveal the browsing
  location, and distinguish the launch directory with dim text and a `(starting dir)`
  label. Enter confirms any highlighted directory, including filtered matches.
  Preserve the selected path when changing views.
- Browse directories with the mouse: click to highlight, double-click to open,
  and use the wheel to move through the views. The `← Back to parent` row above the current
  directory opens the parent. Double-clicks tolerate slight horizontal pointer drift.
- Create a directory in the current location with `n`, edit its name, and focus
  it after creation; cancel name input with Esc or correct creation errors.
- Show each column's directory path in panes view, with clipped
  paths that follow navigation and the highlighted directory preview.
- Install through the source-built `nik-makarov/tap` Homebrew tap, with automatic
  versioned releases and formula updates after successful pushes to main.
- Browse directories with Enter, arrow keys, or `hjkl`, then confirm the destination
  with Cmd+Enter on macOS or Ctrl+Enter on Linux in a compatible terminal.
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

- Keep the highlighted row centered when scrolling through a long tree. Near the
  beginning or end, keep the first or last rows in view. Expanding a folder
  places its name on the first tree row to show more of its contents.
- Remove the four-row Change Dir button to show more entries in tree and panes views.
- Make the Delete button compact and use the same color as the other entry
  actions. Keep its keyboard shortcut available in controls help.
- Reuse an already-open project by default when opening a folder in the editor;
  open a new project window otherwise. Keep opening individual files in new
  windows. Configure or disable folder and file editor actions independently.
- Use tree as the default view, followed by panes. Remove the compact list mode;
  saved `list` preferences now fall back to tree.
  Continue restoring a valid saved mode.
- Display folder names without a trailing slash in all views and side panes.
- Shorten interface wording to dir/dirs and label the current row `(current dir)`.
- In panes views, confirm the current directory and exit with Enter or a
  double-click on the `.` row. Enter and double-clicks on other directories continue browsing.
- Focus the emphasized `Back to parent` row first with Left, `h`, or Backspace;
  open the parent when that row is already highlighted.
- Make the current directory's address more prominent with a bold final name,
  subdued pane labels, and dim side entries that retain their type colors.
- Skip automatic releases and tap updates for pushes that only change Markdown
  files or docs/; include those changes in the next release.
- Rename the command, shell function, executables, archives, and Go modules to
  `cdd`, using `github.com/nik-makarov/cdd` as the repository path.
- Use dim theme-aware text for files so directory entries remain more prominent.
- Start in a three-pane view with an active middle list, parent/child
  context, colored files, and distinctly marked directory symlinks. Press `v`
  to switch between panes and the expandable tree.
- Reduce executable size with a small compatible argument parser and simpler
  message formatting, preserving error causes and shell integration.

### Fixed

- Recognize Finder's Cmd+Delete (Backspace) shortcut in Ghostty/agterm on macOS,
  including their legacy Ctrl+U encoding. Keep exact-name confirmation and leave
  Linux Ctrl+U unchanged.
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
