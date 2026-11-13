# cdd

Choose a directory interactively and change into it in your current shell.

From the extracted archive, run without installation:

```sh
eval "$(./cdd --init zsh)"
cdd
```

For bash, replace `zsh` with `bash`. Use Up/Down to move, Right to open a directory,
Left to go to its parent, Enter to choose, and Esc to cancel. `/` filters names.
The first row, `.`, chooses the directory currently being viewed.

The shell function is required for `cd`; running the binary alone prints a path.
The function remembers the binary's absolute location. If you move the binary,
run its `--init` command again from the new location. For future shells, add
`eval "$(/absolute/path/to/cdd --init zsh)"` to your shell configuration.

Keep LICENSE and THIRD_PARTY_LICENSES.txt when redistributing the executable.
RELEASE_NOTES.md describes this version. Sources and full documentation:
https://github.com/nik-makarov/cdd
