package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadViewPreference(t *testing.T) {
	for _, tt := range []struct {
		name, value string
		want        viewMode
	}{
		{"panes", "panes\n", panesView},
		{"tree", "tree\n", outlineView},
		{"removed list", "list\n", outlineView},
		{"whitespace", " tree\r\n", outlineView},
		{"unknown", "removed-view\n", outlineView},
		{"old ordinal", "1\n", outlineView},
		{"empty", "", outlineView},
		{"malformed", "tree\x00list", outlineView},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "last-view")
			if err := os.WriteFile(path, []byte(tt.value), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := readViewPreference(path); got != tt.want {
				t.Fatalf("mode = %s, want %s", got, tt.want)
			}
		})
	}
	for _, path := range []string{"", t.TempDir(), filepath.Join(t.TempDir(), "missing")} {
		if readViewPreference(path) != outlineView {
			t.Fatal("missing or unreadable preferences must use the default view")
		}
	}
}

func TestViewPreferenceSurvivesCancellationAndRestart(t *testing.T) {
	root := fixture(t)
	t.Setenv("HOME", root)
	path := filepath.Join(t.TempDir(), "config", "cdd", "last-view")
	start := func() browser {
		m, err := newBrowser(root, false)
		if err != nil {
			t.Fatal(err)
		}
		m.restoreViewPreference(path)
		return m
	}
	m := start()
	if m.mode != outlineView {
		t.Fatal("first run must use tree")
	}
	for _, want := range []viewMode{panesView, outlineView} {
		m = press(m, "v", "v")
		if m.mode != want || readViewPreference(path) != want {
			t.Fatalf("switch to %s was not persisted immediately", want)
		}
		if !m.handle(keyEvent{name: "esc"}) || m.selected != "" {
			t.Fatal("saving a view must preserve cancellation without a selected path")
		}
		m = start()
		if m.mode != want || m.focusedPath() != root || m.origin != root {
			t.Fatalf("restart did not restore %s at the starting location", want)
		}
		if want == outlineView && m.treeRoot != root {
			t.Fatal("restored tree must initialize its home root and branches")
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "tree\n" {
		t.Fatal("preference must store only a stable mode name")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatal("preference file must be private")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 || entries[0].Name() != "last-view" {
		t.Fatal("atomic replacement left temporary files behind")
	}
}

func TestFailedModeSwitchPreservesPreference(t *testing.T) {
	root := fixture(t)
	t.Setenv("HOME", root)
	path := filepath.Join(t.TempDir(), "last-view")
	if err := writeViewPreference(path, outlineView); err != nil {
		t.Fatal(err)
	}
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	m.restoreViewPreference(path)
	m = focus(t, m, "regular-file")
	if m.mode != outlineView {
		t.Fatal("tree mode did not load")
	}
	// Tree-to-panes must reopen the selected entry's containing directory.
	// Removing that directory makes the transition fail before it changes mode.
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	m = press(m, "v", "v")
	if m.mode != outlineView || m.message == "" || readViewPreference(path) != outlineView {
		t.Fatal("failed transitions must not overwrite the last successfully used mode")
	}
}

func TestViewPreferenceWriteFailureKeepsNavigationUsable(t *testing.T) {
	root := fixture(t)
	t.Setenv("HOME", root)
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	m.restoreViewPreference(filepath.Join(blocked, "last-view"))
	m = press(m, "v", "v")
	if m.mode != panesView || !strings.Contains(m.message, "Could not save view") || m.selected != "" {
		t.Fatal("save failure must report an error without undoing the selected mode")
	}
	if !m.handle(keyEvent{name: "enter"}) || m.selected != root {
		t.Fatal("save failure must not prevent directory confirmation")
	}
}

func TestUnavailableSavedTreeFallsBackToPanes(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "project")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	path := filepath.Join(t.TempDir(), "last-view")
	if err := writeViewPreference(path, outlineView); err != nil {
		t.Fatal(err)
	}
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(home); err != nil {
		t.Fatal(err)
	}
	m.restoreViewPreference(path)
	if m.mode != panesView || !strings.Contains(m.message, "Could not restore tree view") {
		t.Fatal("unavailable saved tree must leave the default view usable")
	}
	if readViewPreference(path) != outlineView {
		t.Fatal("startup failure must not overwrite the saved preference")
	}
}
