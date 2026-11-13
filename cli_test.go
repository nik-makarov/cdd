package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIWithoutTerminal(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		code int
	}{
		{"help", []string{"--help"}, 0},
		{"version", []string{"--version"}, 0},
		{"init zsh", []string{"--init", "zsh"}, 0},
		{"init bash", []string{"--init", "bash"}, 0},
		{"unknown shell", []string{"--init", "fish"}, 2},
		{"unknown flag", []string{"--unknown"}, 2},
		{"extra args", []string{"a", "b"}, 2},
		{"missing directory", []string{filepath.Join(t.TempDir(), "missing")}, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if got := run(tt.args, &out, &errOut); got != tt.code {
				t.Fatalf("exit = %d, want %d; stderr: %s", got, tt.code, &errOut)
			}
			if tt.code != 0 && out.Len() != 0 {
				t.Fatalf("error contaminated stdout: %q", &out)
			}
			if tt.name == "help" && (!strings.Contains(errOut.String(), "Usage:") || out.Len() != 0) {
				t.Fatal("help missing or leaked into stdout")
			}
			if tt.name == "version" && out.String() != "0go "+version+"\n" {
				t.Fatalf("version output: %q", &out)
			}
		})
	}
}
