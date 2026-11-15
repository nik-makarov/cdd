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

func TestConfiguredEditorKeyAndWindowBehavior(t *testing.T) {
	capture := fakeActionCLI(t, "zed")
	root := fixture(t)
	for _, name := range []string{"Alpha", "regular-file"} {
		m := focus(t, loadOutline(t, root), name)
		m.config.Keys.OpenInEditor = "e"
		if m.handle(keyEvent{name: "z", text: "z"}) || m.action.kind != "" {
			t.Fatal("old binding must not stay hardcoded")
		}
		if m.handle(keyEvent{name: "e", text: "e"}) || m.selected != "" || m.action != (externalAction{kind: "editor", path: m.focusedPath(), command: m.selectedEditorCommand()}) {
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
		flag := "--new"
		if name == "Alpha" {
			flag = "--classic"
		}
		if !slices.Equal(args, []string{flag, "--", m.focusedPath()}) {
			t.Fatalf("wrong Zed arguments: %q", args)
		}
		if !strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "e Open in editor") {
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
	if err := runExternalAction(t.Context(), externalAction{kind: "editor", path: filepath.Join(t.TempDir(), "missing"), command: m.selectedEditorCommand()}, "linux"); err != nil {
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
					if name == "Alpha" {
						flag = "--classic"
					}
					if programs[0] == "code" {
						flag = "--new-window"
					}
					assertEditorArgs(t, capture, []string{flag, "--", before})
				}
			}
			t.Setenv("CDD_TEST_ACTION", "fail")
			action := externalAction{kind: "editor", path: root, command: defaultConfiguration().EditorDirCommand}
			if err := runExternalAction(t.Context(), action, runtime.GOOS); err == nil || !strings.HasPrefix(err.Error(), filepath.Join(dir, programs[0])+":") {
				t.Fatalf("an installed editor's failure must not launch a fallback: %v", err)
			}
		})
	}
}

func TestCustomEditorCommandAndFailures(t *testing.T) {
	capture := fakeActionCLI(t, "custom editor")
	m := focus(t, loadOutline(t, fixture(t)), "regular-file")
	m.config.EditorFileCommand = `missing-editor || "custom editor" --new-window --profile 'Work Profile'`
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
	m.config.EditorFileCommand = ""
	m = press(m, "z", "z")
	if m.action.kind != "" || strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "Open in editor") {
		t.Fatal("empty editor command must disable the action and hint")
	}
}

func TestSeparateEditorCommandsFollowSelection(t *testing.T) {
	capture := fakeActionCLI(t, "folder-editor")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(executable, filepath.Join(filepath.Dir(capture), "file-editor")); err != nil {
		t.Fatal(err)
	}
	root := fixture(t)
	for name, target := range map[string]string{"dir-link": "Alpha", "file-link": "regular-file"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, name := range []string{"Alpha", "dir-link", "regular-file", "file-link"} {
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, name)
			m.config.EditorDirCommand = "folder-editor --project"
			m.config.EditorFileCommand = "file-editor --single"
			isDir := name == "Alpha" || name == "dir-link"
			command, flag := "file-editor --single", "--single"
			if isDir {
				command, flag = "folder-editor --project", "--project"
			}
			m = press(m, "z", "z")
			if m.action.command != command || m.selected != "" {
				t.Fatalf("%s/%s dispatched the wrong command: %+v", mode, name, m.action)
			}
			if err := runExternalAction(t.Context(), m.action, runtime.GOOS); err != nil {
				t.Fatal(err)
			}
			assertEditorArgs(t, capture, []string{flag, "--", filepath.Join(root, name)})
			m.action = externalAction{}
			if isDir {
				m.config.EditorFileCommand = ""
			} else {
				m.config.EditorDirCommand = ""
			}
			m = press(m, "z", "z")
			if m.action.command != command {
				t.Fatal("disabling the other entry type must not affect this action")
			}
			m.action = externalAction{}
			m.config.EditorDirCommand, m.config.EditorFileCommand = "", ""
			m = press(m, "z", "z")
			if m.action.kind != "" || strings.Contains(plainFrame(m.helpLines(120)[0].render(120)), "Open in editor") {
				t.Fatal("disabled command must hide the hint and leave the key inert")
			}
		}
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
