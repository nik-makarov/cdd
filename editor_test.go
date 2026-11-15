package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestConfiguredEditorKeyAndNewWindow(t *testing.T) {
	capture := fakeActionCLI(t, "zed")
	root := fixture(t)
	for _, name := range []string{"Alpha", "regular-file"} {
		m := focus(t, loadOutline(t, root), name)
		m.config.Keys.OpenInEditor = "e"
		if m.handle(keyEvent{name: "z", text: "z"}) || m.action.kind != "" {
			t.Fatal("old binding must not stay hardcoded")
		}
		if m.handle(keyEvent{name: "e", text: "e"}) || m.selected != "" || m.action != (externalAction{kind: "editor", path: m.focusedPath(), command: m.config.EditorCommand}) {
			t.Fatal("configured key must request the editor without choosing a shell directory")
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
		if !strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "e Editor: new window") {
			t.Fatal("help must use the configured key")
		}
	}
	for _, input := range []string{"/", "n", "N"} {
		m := loadOutline(t, root)
		m.config.Keys.OpenInEditor = "e"
		m = press(m, input, input)
		m = press(m, "e", "e")
		if m.action.kind != "" || (m.query != "e" && m.newName != "e") {
			t.Fatal("configured key must remain literal during input")
		}
	}
}

func TestMissingEditorsAreIgnored(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	m := loadOutline(t, fixture(t))
	m.message = "existing message"
	if m.handle(keyEvent{name: m.config.Keys.OpenInEditor}) || m.action.kind != "" || m.message != "existing message" || m.selected != "" {
		t.Fatal("missing editors must make the binding inert")
	}
	if err := runExternalAction(t.Context(), externalAction{kind: "editor", path: filepath.Join(t.TempDir(), "missing"), command: m.config.EditorCommand}, "linux"); err != nil {
		t.Fatal("missing CLI must remain ignored at execution time")
	}
	file := focus(t, m, "regular-file")
	file.previewFile("linux")
	file = press(file, file.config.Keys.OpenInEditor, "")
	if file.filePreview == nil || file.action.kind != "" {
		t.Fatal("missing editors must not dismiss the built-in preview")
	}
}

func TestEditorFallbackPriorityAndLiteralPaths(t *testing.T) {
	for _, programs := range [][]string{{"zed", "zeditor", "code"}, {"zeditor", "code"}, {"code"}} {
		t.Run(programs[0], func(t *testing.T) {
			capture := fakeActionCLI(t, programs[0])
			dir := filepath.Dir(capture)
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range programs[1:] {
				if err := os.Symlink(executable, filepath.Join(dir, name)); err != nil {
					t.Fatal(err)
				}
			}
			root := fixture(t)
			file := "-\u754c ' $() `touch INJECTED`; file"
			if err := os.WriteFile(filepath.Join(root, file), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"Alpha", file} {
				m := focus(t, loadOutline(t, root), name)
				before := m.focusedPath()
				m = press(m, "z", "z")
				if m.action.kind != "editor" || m.selected != "" || m.focusedPath() != before {
					t.Fatal("editor action must preserve browsing state")
				}
				for _, platform := range []string{"linux", "darwin"} {
					program, _, err := actionCommand(m.action, platform)
					if err != nil || program != filepath.Join(dir, programs[0]) {
						t.Fatalf("wrong editor priority: %q, %v", program, err)
					}
					if err := runExternalAction(t.Context(), m.action, platform); err != nil {
						t.Fatal(err)
					}
					flag := "--new"
					if programs[0] == "code" {
						flag = "--new-window"
					}
					assertEditorArgs(t, capture, []string{flag, "--", before})
				}
			}
			t.Setenv("CDD_TEST_ACTION", "fail")
			action := externalAction{kind: "editor", path: root, command: defaultConfiguration().EditorCommand}
			if err := runExternalAction(t.Context(), action, runtime.GOOS); err == nil || !strings.HasPrefix(err.Error(), filepath.Join(dir, programs[0])+":") {
				t.Fatalf("an installed editor's failure must not launch a fallback: %v", err)
			}
		})
	}
}

func TestCustomEditorCommandAndFailures(t *testing.T) {
	capture := fakeActionCLI(t, "custom editor")
	m := focus(t, loadOutline(t, fixture(t)), "regular-file")
	m.config.EditorCommand = `missing-editor || "custom editor" --new-window --profile 'Work Profile'`
	m.previewFile("linux")
	m = press(m, "z", "z")
	if m.filePreview != nil || m.action.kind != "editor" {
		t.Fatal("editor key must work from preview")
	}
	if err := runExternalAction(t.Context(), m.action, runtime.GOOS); err != nil {
		t.Fatal(err)
	}
	assertEditorArgs(t, capture, []string{"--new-window", "--profile", "Work Profile", "--", m.focusedPath()})
	t.Setenv("CDD_TEST_ACTION", "fail")
	if err := runExternalAction(t.Context(), m.action, runtime.GOOS); err == nil || len(err.Error()) > 4400 {
		t.Fatal("installed editor failures must be reported with bounded diagnostics")
	}
	t.Setenv("CDD_TEST_ACTION", "wait")
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if err := runExternalAction(ctx, m.action, runtime.GOOS); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("editor cancellation not propagated: %v", err)
	}
	if err := os.Remove(m.focusedPath()); err != nil {
		t.Fatal(err)
	}
	if err := runExternalAction(t.Context(), m.action, runtime.GOOS); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale paths must be rejected before launch: %v", err)
	}
	m.action = externalAction{}
	m.config.EditorCommand = ""
	m = press(m, "z", "z")
	if m.action.kind != "" || strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "Editor: new window") {
		t.Fatal("empty editor command must disable the action and hint")
	}
}

func assertEditorArgs(t *testing.T, capture string, want []string) {
	t.Helper()
	data, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil || !slices.Equal(args, want) {
		t.Fatalf("editor arguments=%q, want=%q, error=%v", args, want, err)
	}
}

func TestParseEditorCommand(t *testing.T) {
	for _, tt := range []struct {
		command string
		want    [][]string
	}{
		{"", nil},
		{" \t ", nil},
		{"zed --new||code --new-window", [][]string{{"zed", "--new"}, {"code", "--new-window"}}},
		{`'/opt/My Editor/code' --profile "Work Space" ''`, [][]string{{"/opt/My Editor/code", "--profile", "Work Space", ""}}},
		{`custom\ editor --flag='literal $() || ; *'`, [][]string{{"custom editor", "--flag=literal $() || ; *"}}},
		{`editor "a\"b" 'a\b'`, [][]string{{"editor", "a\"b", `a\b`}}},
	} {
		got, err := parseEditorCommand(tt.command)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("%q: got=%q, want=%q, error=%v", tt.command, got, tt.want, err)
		}
	}
	for _, command := range []string{"zed ||", "|| code", "zed || || code", "''", "zed || ''", `zed "`, `zed \`, "zed | code", "zed && code", "zed; code", "zed > file", "zed $(code)", "zed `code`", "zed\ncode", "\n", "zed\x00"} {
		if _, err := parseEditorCommand(command); err == nil {
			t.Fatalf("accepted invalid command %q", command)
		}
	}
}
