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

const help = `0go — interactively choose a directory

Usage: 0go [--hidden] [START_DIRECTORY]
       0go --init zsh|bash

Options:
  --hidden       Show hidden directories initially
  --init SHELL   Print shell integration (zsh or bash)
  --version      Print version
  -h, --help     Show help

Keys:
  ↑/↓, k/j       Move selection
  Tab, Shift+Tab Switch columns in a long list
  v              Switch tree/list view (tree is the default)
  →, l           Open highlighted directory
  ←, h, Backspace  Go to parent
  Enter          Select highlighted directory and exit
  /              Filter directory names (Enter/Esc finishes filtering)
  .              Toggle hidden directories
  ~              Go to home directory
  Ctrl+R         Refresh
  ?              Toggle keyboard help (Esc closes help)
  Esc, q, Ctrl+C Cancel without changing directory

The first row (.) selects the current directory.
Tree view also lists files; they cannot be opened or selected as a destination.
Without shell integration, stdout contains only the selected absolute path,
without a trailing newline. Cancellation produces no output and exits 130.
Connect the binary to your shell: eval "$(/path/to/0go --init zsh)"
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
		io.WriteString(stderr, "0go: expected at most one starting directory\n")
		return 2
	}
	if opts.initShell != "" {
		if opts.initShell != "zsh" && opts.initShell != "bash" {
			io.WriteString(stderr, "0go: --init supports zsh and bash\n")
			return 2
		}
		executable, err := os.Executable()
		if err != nil {
			io.WriteString(stderr, "0go: locate executable: "+err.Error()+"\n")
			return 1
		}
		return writeResult(stdout, stderr, shellInit(executable))
	}
	if opts.showVersion {
		return writeResult(stdout, stderr, "0go "+version+"\n")
	}
	start := "."
	if len(opts.args) == 1 {
		start = opts.args[0]
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		io.WriteString(stderr, "0go: resolve starting directory: "+err.Error()+"\n")
		return 1
	}
	m, err := newBrowser(abs, opts.hidden)
	if err != nil {
		io.WriteString(stderr, "0go: "+err.Error()+"\n")
		return 1
	}
	// Command substitution redirects stdout, but normally leaves stdin and
	// stderr connected to the terminal. Fall back to /dev/tty when necessary.
	input, output := os.Stdin, os.Stderr
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			io.WriteString(stderr, "0go: an interactive terminal is required: "+err.Error()+"\n")
			return 1
		}
		defer tty.Close()
		input, output = tty, tty
	}
	if err := runTerminal(&m, input, output); err != nil {
		if errors.Is(err, errCancelled) {
			return 130
		}
		io.WriteString(stderr, "0go: terminal session: "+err.Error()+"\n")
		return 1
	}
	selected := m.selected
	if selected == "" {
		return 130
	}
	return writeResult(stdout, stderr, selected)
}

func writeResult(stdout, stderr io.Writer, value string) int {
	if _, err := io.WriteString(stdout, value); err != nil {
		io.WriteString(stderr, "0go: write output: "+err.Error()+"\n")
		return 1
	}
	return 0
}
