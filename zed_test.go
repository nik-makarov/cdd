package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestConfiguredZedKeyAndNewWindow(t *testing.T) {
	capture := fakeActionCLI(t, "zed")
	root := fixture(t)
	for _, name := range []string{"Alpha", "regular-file"} {
		m := focus(t, loadOutline(t, root), name)
		m.config.Keys.OpenInZed = "e"
		if m.handle(keyEvent{name: "z", text: "z"}) || m.action.kind != "" {
			t.Fatal("old binding must not stay hardcoded")
		}
		if m.handle(keyEvent{name: "e", text: "e"}) || m.selected != "" || m.action != (externalAction{kind: "zed", path: m.focusedPath()}) {
			t.Fatal("configured key must request Zed without choosing a shell directory")
		}
		if err := runExternalAction(t.Context(), m.action, runtime.GOOS); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(capture)
		if err != nil {
			t.Fatal(err)
		}
		var args []string
		if err := json.Unmarshal(data, &args); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(args, []string{"--new", "--", m.focusedPath()}) {
			t.Fatalf("wrong Zed arguments: %q", args)
		}
		if !strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "e Zed: new window") {
			t.Fatal("help must use the configured key")
		}
	}
	for _, input := range []string{"/", "n"} {
		m := loadOutline(t, root)
		m.config.Keys.OpenInZed = "e"
		m = press(m, input, input)
		m = press(m, "e", "e")
		if m.action.kind != "" || (m.query != "e" && m.newName != "e") {
			t.Fatal("configured key must remain literal during input")
		}
	}
}

func TestMissingZedIsIgnoredAndLinuxAliasWorks(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	m := loadOutline(t, fixture(t))
	m.message = "existing message"
	if m.handle(keyEvent{name: m.config.Keys.OpenInZed}) || m.action.kind != "" || m.message != "existing message" || m.selected != "" {
		t.Fatal("missing Zed must make the binding inert")
	}
	if err := runExternalAction(t.Context(), externalAction{kind: "zed", path: filepath.Join(t.TempDir(), "missing")}, "linux"); err != nil {
		t.Fatal("missing CLI must remain ignored at execution time")
	}
	capture := fakeActionCLI(t, "zeditor")
	if zedCLI("linux") != filepath.Join(filepath.Dir(capture), "zeditor") || zedCLI("darwin") != "" {
		t.Fatal("zeditor alias must be supported on Linux")
	}
}
