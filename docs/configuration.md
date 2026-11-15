# Configuration

cdd ships its key defaults in [config/default.json](../config/default.json),
embedded into the binary. Override them in `cdd/config.json` under the operating
system's user configuration directory:

- macOS: `~/Library/Application Support/cdd/config.json`.
- Linux: `$XDG_CONFIG_HOME/cdd/config.json`, or `~/.config/cdd/config.json`.

For example, change the Zed shortcut from `z` to `e`:

```json
{
  "keys": {
    "open_in_zed": "e"
  }
}
```

Use an empty string to disable it. The binding accepts one ASCII letter, including
uppercase letters; built-in keys `h`, `j`, `k`, `l`, `q`, `n`, `N`, and `v`
cannot be reassigned. Shortcut letters stay literal while editing a filter or
file/dir name. Help displays the configured key. Restart cdd after editing.

Missing files and omitted keys use the shipped defaults. Malformed JSON, unknown
keys, and conflicting bindings show an error and use defaults for that run.
cdd does not rewrite this file. The remembered view remains in the separate
`last-view` file; it stores `tree` or `panes`. Old `list` values use tree.

The Zed action uses the installed `zed` CLI, or `zeditor` on Linux if `zed` is
absent. No CLI means the key is ignored. It passes only the highlighted path to
`--new`: a folder opens as a project and a file opens by itself in a new window.
The browsing location and shell directory remain unchanged. Install the CLI using
Zed's own instructions; cdd neither installs nor searches for the application.
See the official [Zed CLI reference](https://zed.dev/docs/reference/cli).
