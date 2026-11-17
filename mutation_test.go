package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenameEntries(t *testing.T) {
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, name := range []string{"Alpha", "regular-file", "dir-link", "broken-link"} {
			root := fixture(t)
			for link, target := range map[string]string{"dir-link": "Alpha", "broken-link": "absent"} {
				if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
					t.Fatal(err)
				}
			}
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, name)
			before, err := os.Lstat(m.focusedPath())
			if err != nil {
				t.Fatal(err)
			}
			m = press(m, "r", "r")
			if m.mutation == nil || m.mutation.input != name {
				t.Fatal("rename must start with the selected name")
			}
			newName := ".\u754c quote' $() `literal`; renamed"
			m = press(m, "text", newName)
			m = press(m, "enter", "")
			path := filepath.Join(root, newName)
			after, err := os.Lstat(path)
			if err != nil || !os.SameFile(before, after) || m.message != "" || m.mutation != nil || m.focusedPath() != path || !m.hidden || m.selected != "" {
				t.Fatalf("%s/%s: rename failed: %v, %s", mode, name, err, m.message)
			}
			if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
				t.Fatal("old entry remained")
			}
			if name == "dir-link" {
				if _, err := os.Stat(filepath.Join(root, "Alpha", "nested")); err != nil {
					t.Fatal("renaming a link affected its target")
				}
			}
		}
	}
}

func TestRenameRejectsOverwriteAndStaleEntry(t *testing.T) {
	root := fixture(t)
	if err := os.Symlink("absent", filepath.Join(root, "broken")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", ".", "..", "../escape", "a/b", "a\x00b", "Alpha", "broken"} {
		m := focus(t, loadBrowser(t, root), "regular-file")
		m = press(m, "r", "r")
		m.mutation.input = name
		m = press(m, "enter", "")
		if m.mutation == nil || m.message == "" {
			t.Fatalf("accepted invalid or existing name %q", name)
		}
		data, err := os.ReadFile(filepath.Join(root, "regular-file"))
		if err != nil || string(data) != "text" {
			t.Fatal("failed rename modified the source")
		}
	}
	m := focus(t, loadBrowser(t, root), "regular-file")
	m = press(m, "r", "r")
	if err := os.Rename(filepath.Join(root, "regular-file"), filepath.Join(root, "saved")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "regular-file"), []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	m = press(m, "text", "renamed")
	m = press(m, "enter", "")
	if m.mutation == nil || !strings.Contains(m.message, "Entry changed") {
		t.Fatal("changed source must invalidate the prompt")
	}
}

func TestRenameButtonCancellationAndProtection(t *testing.T) {
	root := fixture(t)
	m := focus(t, loadBrowser(t, root), "Alpha")
	found := false
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.entryButtonAt(x, y) == "r" {
				m.handle(keyEvent{name: "mouse-left", x: x, y: y})
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found || m.mutation == nil {
		t.Fatal("visible Rename button must start input")
	}
	m = press(m, "backspace", "")
	if m.mutation.input != "" {
		t.Fatal("backspace must clear the initially selected name")
	}
	m = press(m, "text", "\u754c\x1b[31m")
	if strings.ContainsRune(plainFrame(m.view()), '\x1b') {
		t.Fatal("rename input leaked terminal controls")
	}
	m = press(m, "esc", "")
	if m.mutation != nil || m.focusedPath() != filepath.Join(root, "Alpha") {
		t.Fatal("cancel must preserve selection")
	}
	for _, name := range []string{".", ".."} {
		m = focus(t, loadBrowser(t, root), name)
		m = press(m, "r", "r")
		if m.mutation != nil || m.message == "" || strings.Contains(m.view(), "[ Rename") || strings.Contains(m.view(), "[ Delete") {
			t.Fatal("navigation rows must be protected")
		}
	}
	m = loadOutline(t, root)
	m.focusPath(root)
	m = press(m, "r", "r")
	if m.mutation != nil || m.message == "" {
		t.Fatal("starting dir must be protected in the tree")
	}
	for _, input := range []string{"/", "n", "N"} {
		m = press(loadBrowser(t, root), input, input)
		m = press(m, "r", "r")
		if m.mutation != nil || m.query != "r" && m.newName != "r" {
			t.Fatal("rename shortcut must remain literal during text input")
		}
	}
}
