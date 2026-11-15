package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestEnterOpensFileWithoutSelectingDestination(t *testing.T) {
	for _, mode := range []viewMode{outlineView, panesView} {
		root := fixture(t)
		m := loadOutline(t, root)
		if mode == panesView {
			m = press(m, "v", "v")
		}
		m = focus(t, m, "regular-file")
		path := m.focusedPath()
		if m.handle(keyEvent{name: "enter"}) || m.action != (externalAction{kind: "open", path: path}) || m.selected != "" || m.focusedPath() != path {
			t.Fatal("Enter must request opening without confirming a directory")
		}
		m.action = externalAction{}
		if m.handle(keyEvent{name: "confirm"}) || m.action.kind != "" || m.selected != "" {
			t.Fatal("directory confirmation must continue rejecting files")
		}
		m.message = ""
		m.previewFile("linux")
		if m.handle(keyEvent{name: "enter"}) || m.filePreview != nil || m.action.kind != "open" {
			t.Fatal("Enter in preview must open the file and return to browsing")
		}
		m.action = externalAction{}
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		m.handle(keyEvent{name: "enter"})
		if m.action.kind != "" || m.message == "" {
			t.Fatal("stale files must not launch an application")
		}
	}
}

func TestDefaultApplicationReceivesLiteralPath(t *testing.T) {
	for _, tt := range []struct{ platform, command string }{{"darwin", "open"}, {"linux", "xdg-open"}} {
		t.Run(tt.platform, func(t *testing.T) {
			capture := fakeActionCLI(t, tt.command)
			path := filepath.Join(t.TempDir(), "file' $() ; \u754c.txt")
			if err := os.WriteFile(path, []byte("text"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := runExternalAction(t.Context(), externalAction{kind: "open", path: path}, tt.platform); err != nil {
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
			if !slices.Equal(args, []string{path}) {
				t.Fatalf("file path was changed: %q", args)
			}
		})
	}
}
