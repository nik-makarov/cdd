package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var version = "dev"

const help = `cdd — interactively choose a dir

Usage: cdd [--hidden] [START_DIR]
       cdd --init zsh|bash

Options:
  --hidden       Show hidden dirs initially
  --init SHELL   Print shell integration (zsh or bash)
  --version      Print version
  -h, --help     Show help

Keys:
  ↑/↓, k/j       Move selection
  v              Cycle views: tree, panes
  Tab / Shift+Tab  Switch between the tree and recent places
  Enter          Open file; change dir in tree; browse dir in panes (. confirms)
  →, l           Open highlighted dir
  ←, h, Backspace  Focus parent first; press again to go up
  Cmd+Enter      Change dir and exit (macOS)
  Ctrl+Enter     Change dir and exit (Linux)
  n              Create a dir in the current location
  Shift+N        Create an empty file in the highlighted dir (or file's parent)
  r              Rename the highlighted file/dir without overwriting
  f              Open the highlighted dir in Finder (macOS)
  y              Copy the highlighted file/dir path to the clipboard
  Cmd+C / Cmd+V  Copy / paste a file or dir on macOS (also Copy / Paste buttons)
                 Review the full destination, then Enter to paste or Esc to cancel
  Cmd+Backspace  Move to Trash after typing the exact name (macOS)
  Ctrl+Backspace Move to Trash after typing the exact name (Linux; requires gio)
  Space          Preview file (Quick Look on macOS; text preview on Linux)
  Configured key Open dir/file in editor (Zed, then VS Code)
  /              Filter names (Esc finishes; Enter confirms in tree)
  .              Toggle hidden dirs
  ~              Go to home dir
  Ctrl+R         Refresh
  \\              Show tree from filesystem root
  ?              Toggle controls help (Esc closes help)
  Esc, q, Ctrl+C Cancel without changing dir

Mouse (in compatible terminals):
  Left click     Highlight an entry
  Double click   Open dir; on . change dir and exit
  Wheel          Move selection up/down

Tree view expands branches in place: Right expands or focuses the first child;
Left collapses or focuses the parent. Enter always chooses the highlighted dir,
including while filtering. The root is home when it contains the location, or /.
Ancestors open automatically; the launch location is marked (starting dir).
Click a disclosure arrow or double-click a dir to toggle its branch.
Recent places lists the last 50 distinct confirmed dirs, newest first.
At 80 columns or more it sits beside the tree. Below 80, Tab swaps panels.
The tree filter matches expanded branches and keeps their ancestors visible.
In tree view, n creates inside the highlighted dir (or a file's parent).

In panes, the first row (← Back to parent) goes up; it is absent at /.
The next row, . (current dir), selects the current dir and is highlighted initially.
On macOS, the Finder button opens the highlighted dir while keeping cdd open.
Copy Path copies the absolute path; Linux requires wl-copy (Wayland) or xclip (X11).
The last used view is restored on startup; tree is the default.
Customize editor_dir_command, editor_file_command, and keys.open_in_editor
in cdd/config.json. Dirs focus an open project; files open in a new window.
The binding is ignored when none of the configured editor CLIs is installed.
Modified Enter/Backspace use the kitty keyboard protocol.
macOS also accepts the legacy Cmd+Backspace byte used by Ghostty/agterm.
Delete always uses the system Trash, with no permanent-deletion fallback.
Files open in the default application with Enter; they cannot be a cd destination.
Shell integration is required to choose a dir and change the current shell.
The binary returns only the selected absolute path, without a trailing newline.
Cancellation produces no output and exits 130.
Add eval "$(/path/to/cdd --init zsh)" to ~/.zshrc, then reopen your terminal.
For bash, use --init bash in ~/.bashrc.
`

type options struct {
	hidden, showVersion, showHelp bool
	initShell                     string
	args                          []string
}

// Like flag, accept one or two dashes and stop at -- or the first positional arg.
func parseArgs(args []string) (options, error) {
	var opts options
	for len(args) > 0 {
		arg := args[0]
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			break
		}
		args = args[1:]
		if arg == "--" {
			break
		}
		name := strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")
		if name == "" || name[0] == '-' || name[0] == '=' {
			return opts, errors.New("bad flag syntax: " + arg)
		}
		name, value, hasValue := strings.Cut(name, "=")
		switch name {
		case "h", "help":
			opts.showHelp = true
			return opts, nil
		case "hidden", "version":
			enabled := true
			if hasValue {
				var err error
				enabled, err = strconv.ParseBool(value)
				if err != nil {
					return opts, errors.New("invalid boolean value " + strconv.Quote(value) + " for -" + name + ": parse error")
				}
			}
			if name == "hidden" {
				opts.hidden = enabled
			} else {
				opts.showVersion = enabled
			}
		case "init":
			if !hasValue {
				if len(args) == 0 {
					return opts, errors.New("flag needs an argument: -init")
				}
				value, args = args[0], args[1:]
			}
			opts.initShell = value
		default:
			return opts, errors.New("flag provided but not defined: -" + name)
		}
	}
	opts.args = args
	return opts, nil
}

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseArgs(args)
	if err != nil {
		io.WriteString(stderr, err.Error()+"\n"+help)
		return 2
	}
	if opts.showHelp {
		io.WriteString(stderr, help)
		return 0
	}
	if len(opts.args) > 1 {
		io.WriteString(stderr, "cdd: expected at most one starting dir\n")
		return 2
	}
	if opts.initShell != "" {
		if opts.initShell != "zsh" && opts.initShell != "bash" {
			io.WriteString(stderr, "cdd: --init supports zsh and bash\n")
			return 2
		}
		executable, err := os.Executable()
		if err != nil {
			io.WriteString(stderr, "cdd: locate executable: "+err.Error()+"\n")
			return 1
		}
		return writeResult(stdout, stderr, shellInit(executable))
	}
	if opts.showVersion {
		return writeResult(stdout, stderr, "cdd "+version+"\n")
	}
	if !shellIntegrationActive(stdout) {
		var styles viewStyles
		if output, ok := stderr.(*os.File); ok {
			styles = terminalStyles(term.IsTerminal(int(output.Fd())), os.Getenv("TERM"), os.Getenv("NO_COLOR"))
		}
		io.WriteString(stderr, shellSetupMessage(styles))
		return 1
	}
	start := "."
	if len(opts.args) == 1 {
		start = opts.args[0]
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		io.WriteString(stderr, "cdd: resolve starting dir: "+err.Error()+"\n")
		return 1
	}
	m, err := newBrowser(abs, opts.hidden)
	if err != nil {
		io.WriteString(stderr, "cdd: "+err.Error()+"\n")
		return 1
	}
	if origin, err := os.Getwd(); err == nil {
		m.origin = origin
	}
	// Command substitution redirects stdout, but normally leaves stdin and
	// stderr connected to the terminal. Fall back to /dev/tty when necessary.
	input, output := os.Stdin, os.Stderr
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			io.WriteString(stderr, "cdd: an interactive terminal is required: "+err.Error()+"\n")
			return 1
		}
		defer tty.Close()
		input, output = tty, tty
	}
	m.restoreViewPreference(viewPreferencePath())
	recentFile := configurationPath("recent-dirs")
	m.recent.paths, err = readRecent(recentFile)
	if err != nil {
		io.WriteString(stderr, "cdd: could not load recent places: "+display(err.Error())+"\n")
	}
	config, configErr := readConfiguration(configurationPath("config.json"))
	m.config = config
	if configErr != nil {
		if m.message != "" {
			m.message += "; "
		}
		m.message += "Could not load config: " + configErr.Error()
	}
	if err := runTerminal(&m, input, output); err != nil {
		if errors.Is(err, errCancelled) {
			return 130
		}
		io.WriteString(stderr, "cdd: terminal session: "+err.Error()+"\n")
		return 1
	}
	selected := m.selected
	if selected == "" {
		return 130
	}
	if code := writeResult(stdout, stderr, selected); code != 0 {
		return code
	}
	if err := recordRecent(recentFile, selected); err != nil {
		io.WriteString(stderr, "cdd: could not save recent place: "+display(err.Error())+"\n")
	}
	return 0
}

func writeResult(stdout, stderr io.Writer, value string) int {
	if _, err := io.WriteString(stdout, value); err != nil {
		io.WriteString(stderr, "cdd: write output: "+err.Error()+"\n")
		return 1
	}
	return 0
}
