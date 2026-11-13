package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The test executable doubles as a picker to exercise the real shell function
// without a TTY. This includes failure after partial output.
func TestMain(m *testing.M) {
	if os.Getenv("CDD_TEST_HELPER") == "1" {
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
