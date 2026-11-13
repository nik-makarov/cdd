package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The test executable doubles as a picker to exercise the real shell function
// without a TTY. This includes failure after partial output.
func TestMain(m *testing.M) {
	if os.Getenv("ZERO_GO_TEST_HELPER") == "1" {
		for _, arg := range os.Args[1:] {
			if arg == "--help" {
				fmt.Print("HELP\n")
				os.Exit(0)
			}
		}
		switch os.Getenv("ZERO_GO_TEST_MODE") {
		case "cancel":
			os.Exit(130)
		case "error":
			fmt.Print(os.Getenv("ZERO_GO_TEST_TARGET"))
			os.Exit(1)
		}
		fmt.Print(os.Getenv("ZERO_GO_TEST_TARGET"))
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
				{"spaces", "select", "with spaces", "", "", "0"},
				{"metacharacters", "select", "quote'\" $() `touch INJECTED` ; * [x]", "", "", "0"},
				{"unicode", "select", "Русский 日本語", "", "", "0"},
				{"trailing newlines", "select", "new\nline\n\n", "", "", "0"},
				{"trailing dot", "select", "trailing.", "", "", "0"},
				{"cancel", "cancel", "cancelled", "", "", "130"},
				{"partial output on error", "error", "failed", "", "", "1"},
				{"help", "select", "unused", "--hidden --help", "HELP\n", "0"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					target := filepath.Join(base, tt.target)
					if err := os.Mkdir(target, 0o755); err != nil {
						t.Fatal(err)
					}
					script := shellInit(helper) + "\n0go " + tt.args + `
_0go_test_code=$?
printf '%s\000%s' "$_0go_test_code" "$PWD"
`
					cmd := exec.Command(binary, "-f", "-c", script)
					cmd.Dir = base
					cmd.Env = append(os.Environ(), "ZERO_GO_TEST_HELPER=1", "ZERO_GO_TEST_MODE="+tt.mode, "ZERO_GO_TEST_TARGET="+target)
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("shell failed: %v: %s", err, out)
					}
					wantPath := target
					if tt.mode != "select" || tt.args != "" {
						wantPath = base
					}
					want := tt.prefix + tt.code + "\x00" + wantPath
					if string(out) != want {
						t.Fatalf("got %q, want %q", out, want)
					}
					if _, err := os.Stat(filepath.Join(base, "INJECTED")); !os.IsNotExist(err) {
						t.Fatal("shell executed filename contents")
					}
				})
			}
		})
	}
}
