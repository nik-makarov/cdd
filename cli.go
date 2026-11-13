package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
  →, l           Open highlighted directory
  ←, h, Backspace  Go to parent
  Enter          Select highlighted directory and exit
  /              Filter directory names (Enter/Esc finishes filtering)
  .              Toggle hidden directories
  ~              Go to home directory
  Ctrl+R         Refresh
  Esc, q, Ctrl+C Cancel without changing directory

The first row (.) selects the current directory.
Without shell integration, stdout contains only the selected absolute path,
without a trailing newline. Cancellation produces no output and exits 130.
Connect the binary to your shell: eval "$(/path/to/0go --init zsh)"
`

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("0go", flag.ContinueOnError)
	flags.SetOutput(stderr)
	hidden := flags.Bool("hidden", false, "show hidden directories")
	initShell := flags.String("init", "", "print shell integration")
	showVersion := flags.Bool("version", false, "print version")
	flags.Usage = func() { fmt.Fprint(stderr, help) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 1 {
		fmt.Fprintln(stderr, "0go: expected at most one starting directory")
		return 2
	}
	if *initShell != "" {
		if *initShell != "zsh" && *initShell != "bash" {
			fmt.Fprintln(stderr, "0go: --init supports zsh and bash")
			return 2
		}
		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintf(stderr, "0go: locate executable: %v\n", err)
			return 1
		}
		return writeResult(stdout, stderr, shellInit(executable))
	}
	if *showVersion {
		return writeResult(stdout, stderr, "0go "+version+"\n")
	}
	start := "."
	if flags.NArg() == 1 {
		start = flags.Arg(0)
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		fmt.Fprintf(stderr, "0go: resolve starting directory: %v\n", err)
		return 1
	}
	m, err := newBrowser(abs, *hidden)
	if err != nil {
		fmt.Fprintf(stderr, "0go: %v\n", err)
		return 1
	}
	// Command substitution redirects stdout, but normally leaves stdin and
	// stderr connected to the terminal. Fall back to /dev/tty when necessary.
	input, output := os.Stdin, os.Stderr
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			fmt.Fprintf(stderr, "0go: an interactive terminal is required: %v\n", err)
			return 1
		}
		defer tty.Close()
		input, output = tty, tty
	}
	if err := runTerminal(&m, input, output); err != nil {
		if errors.Is(err, errCancelled) {
			return 130
		}
		fmt.Fprintf(stderr, "0go: terminal session: %v\n", err)
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
		fmt.Fprintf(stderr, "0go: write output: %v\n", err)
		return 1
	}
	return 0
}
