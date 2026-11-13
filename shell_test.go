package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The test executable doubles as a picker to exercise the real shell function
// without a TTY. This includes failure after partial output.
func TestMain(m *testing.M) {
	if os.Getenv("CDD_TEST_HELPER") == "1" {
		if os.Getenv("CDD_TEST_MODE") == "cli" {
			os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
		}
		for _, arg := range os.Args[1:] {
			if arg == "--help" {
				fmt.Print("HELP\n")
				os.Exit(0)
			}
		}
		switch os.Getenv("CDD_TEST_MODE") {
		case "empty":
			os.Exit(0)
		case "cancel":
			os.Exit(130)
		case "error":
			fmt.Print(os.Getenv("CDD_TEST_TARGET"))
			os.Exit(1)
		}
		fmt.Print(os.Getenv("CDD_TEST_TARGET"))
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestShellInformationalArguments(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, shell := range []string{"zsh", "bash"} {
		t.Run(shell, func(t *testing.T) {
			binary, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("%s is not installed", shell)
			}
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			for _, tt := range []struct {
				name, kind string
				args       []string
				code       int
			}{
				{"version equals", "version", []string{"--version=true"}, 0},
				{"version alias", "version", []string{"-version"}, 0},
				{"last version value wins", "version", []string{"--version=false", "--version=1"}, 0},
				{"help alias", "help", []string{"--h"}, 0},
				{"help after false version", "help", []string{"--version", "--version=false", "--help"}, 0},
				{"init alias", "init", []string{"-init", "zsh"}, 0},
				{"init equals", "init", []string{"--init=bash"}, 0},
				{"invalid boolean", "error", []string{"--version=invalid"}, 2},
				{"missing init value", "error", []string{"--init"}, 2},
				{"unknown before help", "error", []string{"--unknown", "--help"}, 2},
			} {
				t.Run(tt.name, func(t *testing.T) {
					command := "cdd"
					for _, arg := range tt.args {
						command += " " + shellQuote(arg)
					}
					script := shellInit(executable) + "\n" + command + `
_cdd_test_code=$?
printf '%s\000%s' "$_cdd_test_code" "$PWD"
`
					cmd := exec.Command(binary, "-f", "-c", script)
					cmd.Dir = base
					cmd.Env = append(os.Environ(), "CDD_TEST_HELPER=1", "CDD_TEST_MODE=cli")
					var stderr bytes.Buffer
					cmd.Stderr = &stderr
					out, err := cmd.Output()
					if err != nil {
						t.Fatalf("shell failed: %v, %q", err, stderr.String())
					}
					suffix := strconv.Itoa(tt.code) + "\x00" + base
					if !bytes.HasSuffix(out, []byte(suffix)) {
						t.Fatalf("informational arguments changed directory or exit code: stdout %q, stderr %q", out, stderr.String())
					}
					output := string(out[:len(out)-len(suffix)])
					switch tt.kind {
					case "version":
						if output != "cdd "+version+"\n" || stderr.Len() != 0 {
							t.Fatalf("version output changed: stdout %q, stderr %q", output, stderr.String())
						}
					case "init":
						if !strings.HasPrefix(output, "cdd() {") || !strings.HasSuffix(output, "}\n") || stderr.Len() != 0 {
							t.Fatalf("init output changed: stdout %q, stderr %q", output, stderr.String())
						}
					case "help", "error":
						if output != "" || !strings.Contains(stderr.String(), "Usage:") {
							t.Fatalf("usage output changed: stdout %q, stderr %q", output, stderr.String())
						}
					}
				})
			}
		})
	}
}

func TestShellIntegration(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, shell := range []string{"zsh", "bash"} {
		t.Run(shell, func(t *testing.T) {
			binary, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("%s is not installed", shell)
			}
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			// A hostile-looking executable name also exercises shellInit quoting.
			helper := filepath.Join(base, "picker ' $(touch INJECTED) `id`")
			if err := os.Symlink(executable, helper); err != nil {
				t.Fatal(err)
			}
			for _, tt := range []struct {
				name, mode, target, args, prefix, code string
			}{
				{"plain", "select", "plain", "", "", "0"},
				{"spaces", "select", "with spaces", "", "", "0"},
				{"metacharacters", "select", "quote'\" $() `touch INJECTED` ; * [x]", "", "", "0"},
				{"unicode", "select", "\u0420\u0443\u0441\u0441\u043a\u0438\u0439 \u65e5\u672c\u8a9e", "", "", "0"},
				{"trailing newlines", "select", "new\nline\n\n", "", "", "0"},
				{"terminal controls", "select", "escape\x1b[31m\t\rname", "", "", "0"},
				{"trailing dot", "select", "trailing.", "", "", "0"},
				{"current directory", "select", ".", "", "", "0"},
				{"cancel", "cancel", "cancelled", "", "", "130"},
				{"empty selection", "empty", "unused-empty", "", "", "0"},
				{"failed cd", "missing", "missing-directory", "", "", "1"},
				{"partial output on error", "error", "failed", "", "", "1"},
				{"help", "select", "unused", "--hidden --help", "HELP\n", "0"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					target := filepath.Join(base, tt.target)
					if target != base && tt.mode != "missing" {
						if err := os.Mkdir(target, 0o755); err != nil {
							t.Fatal(err)
						}
					}
					script := shellInit(helper) + "\ncdd " + tt.args + `
_cdd_test_code=$?
printf '%s\000%s' "$_cdd_test_code" "$PWD"
`
					cmd := exec.Command(binary, "-f", "-c", script)
					cmd.Dir = base
					cmd.Env = append(os.Environ(), "CDD_TEST_HELPER=1", "CDD_TEST_MODE="+tt.mode, "CDD_TEST_TARGET="+target)
					var stderr bytes.Buffer
					cmd.Stderr = &stderr
					out, err := cmd.Output()
					if err != nil {
						t.Fatalf("shell failed: %v: stdout %q, stderr %q", err, out, stderr.String())
					}
					wantPath := target
					if tt.mode != "select" || tt.args != "" {
						wantPath = base
					}
					want := tt.prefix + tt.code + "\x00" + wantPath
					if string(out) != want {
						t.Fatalf("got %q, want %q", out, want)
					}
					message := stderr.String()
					moved := wantPath != base
					if moved {
						if !strings.HasPrefix(message, "cdd: "+base+" -> ") || !strings.HasSuffix(message, "\n") || strings.Count(message, "\n") != 1 || strings.ContainsAny(message, "\x1b\r\t") {
							t.Fatalf("missing or unsafe move message: %q", message)
						}
						if tt.name == "plain" || tt.name == "trailing dot" {
							if want := "cdd: " + base + " -> " + target + "\n"; message != want {
								t.Fatalf("move message = %q, want %q", message, want)
							}
						}
					} else if tt.mode == "missing" {
						if message == "" || strings.HasPrefix(message, "cdd: ") {
							t.Fatalf("failed cd must report its error without a success message: %q", message)
						}
					} else if message != "" {
						t.Fatalf("no directory change, but received message: %q", message)
					}
					if _, err := os.Stat(filepath.Join(base, "INJECTED")); !os.IsNotExist(err) {
						t.Fatal("shell executed filename contents")
					}
				})
			}
		})
	}
}
