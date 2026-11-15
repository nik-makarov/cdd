# Configuration

cdd ships its defaults in [config/default.json](../config/default.json), embedded
into the binary. Override them in `cdd/config.json` under the operating system's
user configuration directory:

- macOS: `~/Library/Application Support/cdd/config.json`.
- Linux: `$XDG_CONFIG_HOME/cdd/config.json`, or `~/.config/cdd/config.json`.

## Editor commands

Separate strings control editor commands and fallback order for dirs and files:

```json
{
  "editor_dir_command": "zed --classic || zeditor --classic || code --new-window",
  "editor_file_command": "zed --new || zeditor --new || code --new-window"
}
```

These are the shipped defaults. Press `z` to open the highlighted entry:

- Dirs use `editor_dir_command`. Zed's `--classic` reuses an already-open project,
  or opens a new window for a new project. This hidden CLI flag was verified with
  Zed 1.28.1. VS Code also focuses an already-open folder with `--new-window`.
- Files use `editor_file_command` and open by themselves in a new window.
  Directory symlinks use the dir command; file symlinks use the file command.

cdd chooses the first installed CLI in the applicable chain: Zed (`zed`, or
`zeditor` on some Linux distributions), then VS Code (`code`). If none is
installed, the key does nothing. If an installed editor fails to launch, cdd
reports the error without opening another editor. The browsing location and
shell directory remain unchanged. Project matching and focus are handled by
the editor; see [Zed's implementation](https://github.com/zed-industries/zed/blob/v1.28.1/crates/zed/src/zed/open_listener.rs#L743)
and [VS Code's implementation](https://github.com/microsoft/vscode/blob/main/src/vs/platform/windows/electron-main/windowsMainService.ts#L644).

For example, use only VS Code for files while keeping the default dir command:

```json
{"editor_file_command": "code --new-window"}
```

An empty string disables only the corresponding action. The legacy
`editor_command` setting still sets both commands; either new setting overrides
its own value, regardless of JSON field order. To adopt the new dir default,
remove the old shared override or set `editor_dir_command` explicitly.

Commands support arguments, single/double quotes, backslash escapes, and `||`
alternatives. For example, `code --new-window --profile 'Work Profile'` selects a
VS Code profile. cdd appends `--` followed by the highlighted absolute path as one
literal argument; do not include a path placeholder. This is a command list,
not a shell script: pipes, redirects, variables, command substitution, and shell
aliases are unsupported. Use an absolute executable path if it is outside PATH,
quoting it if it contains spaces.

Both editors must expose their CLI in PATH. See the official
[Zed CLI reference](https://zed.dev/docs/reference/cli) and
[VS Code CLI reference](https://code.visualstudio.com/docs/configure/command-line),
including the macOS command to install `code` in PATH. cdd does not install or
search for application bundles.

## Editor shortcut

For example, change the editor shortcut from `z` to `e`:

```json
{"keys": {"open_in_editor": "e"}}
```

Use an empty string to disable the binding. It accepts one ASCII letter, including
uppercase letters; built-in keys `h`, `j`, `k`, `l`, `q`, `n`, `N`, and `v`
cannot be reassigned. Shortcut letters stay literal while editing a filter or
file/dir name. Help displays the configured key as `Editor`. The old
`keys.open_in_zed` name remains an alias; `keys.open_in_editor` wins if both appear.
The commands and shortcut can be combined in the same JSON object.

Restart cdd after editing. Missing files and omitted settings use the shipped
defaults. Malformed JSON, unknown keys, invalid commands, and conflicting bindings
show an error and use defaults for that run. cdd does not rewrite this file.
The remembered view remains in the separate `last-view` file; it stores `tree` or
`panes`. Old `list` values use tree.
