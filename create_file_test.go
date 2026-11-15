package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateEmptyFileInSelectedDirectory(t *testing.T) {
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, selected := range []string{"Alpha", "regular-file", "link"} {
			for _, name := range []string{"note.txt", ".hidden-note", " with spaces ", "\u754c.txt", "-quote' $() `touch INJECTED` ; *"} {
				t.Run(mode.String()+"/"+selected+"/"+name, func(t *testing.T) {
					root := fixture(t)
					if err := os.Symlink("Alpha", filepath.Join(root, "link")); err != nil {
						t.Fatal(err)
					}
					m := loadOutline(t, root)
					if mode == panesView {
						m = press(m, "v", "v")
					}
					m = focus(t, m, selected)
					parent := m.focusedPath()
					if selected == "regular-file" {
						parent = root
					}
					m = press(m, "N", "N")
					if !m.creating || !m.creatingFile || m.creationPath() != parent {
						t.Fatal("file prompt must target the selected dir or file parent")
					}
					m = press(m, "text", name)
					path := filepath.Join(parent, name)
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatal("typing must not create a file")
					}
					if m.handle(keyEvent{name: "enter"}) {
						t.Fatal("creation must leave the browser open")
					}
					info, err := os.Stat(path)
					if err != nil || !info.Mode().IsRegular() || info.Size() != 0 {
						t.Fatalf("empty regular file not created: %v", err)
					}
					if m.creating || m.creatingFile || m.newName != "" || m.query != "" || m.message != "" || m.selected != "" || m.action.kind != "" || m.focusedPath() != path || m.visible[m.cursor].isDir() {
						t.Fatal("new file must be focused without being opened or selected for cd")
					}
					if strings.HasPrefix(name, ".") && !m.hidden {
						t.Fatal("new hidden files must become visible")
					}
					if m.handle(keyEvent{name: "confirm"}) || m.selected != "" {
						t.Fatal("new file must remain ineligible for cd")
					}
					if _, err := os.Stat(filepath.Join(parent, "INJECTED")); !os.IsNotExist(err) {
						t.Fatal("file name was interpreted by a shell")
					}
				})
			}
		}
	}
}

func TestCreateFileRejectsExistingEntriesAndInvalidNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../outside", "Alpha/nested", "bad\x00name", "regular-file", "Alpha", "file-link", "broken-link"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			for name, target := range map[string]string{"file-link": "regular-file", "broken-link": "missing"} {
				if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
					t.Fatal(err)
				}
			}
			m := focus(t, loadOutline(t, root), "regular-file")
			m = press(m, "N", "N")
			m = press(m, "text", name)
			if m.handle(keyEvent{name: "enter"}) || !m.creating || !m.creatingFile || m.message == "" || m.newName != name || m.selected != "" {
				t.Fatal("failed creation must preserve editable input")
			}
			data, err := os.ReadFile(filepath.Join(root, "regular-file"))
			if err != nil || string(data) != "text" {
				t.Fatal("existing file was overwritten")
			}
			if _, err := os.Stat(filepath.Join(root, "missing")); !os.IsNotExist(err) {
				t.Fatal("broken symlink target was created")
			}
			m.newName = "retry.txt"
			m = press(m, "enter", "")
			if m.creating || m.message != "" || m.visible[m.cursor].name != "retry.txt" {
				t.Fatal("corrected name must succeed")
			}
		})
	}
}

func TestCreateFileCancellationAndMissingParent(t *testing.T) {
	for _, key := range []string{"esc", "ctrl+c", "ctrl+d"} {
		root := fixture(t)
		m := focus(t, loadOutline(t, root), "Alpha")
		m = press(m, "N", "N")
		m = press(m, "text", "cancelled")
		if done := m.handle(keyEvent{name: key}); done != (key != "esc") {
			t.Fatal("incorrect cancellation behavior")
		}
		if _, err := os.Stat(filepath.Join(root, "Alpha", "cancelled")); !os.IsNotExist(err) {
			t.Fatal("cancellation created a file")
		}
		if key == "esc" {
			m = press(m, "n", "n")
			if m.creatingFile {
				t.Fatal("file mode leaked into directory creation")
			}
		}
	}
	root := fixture(t)
	m := focus(t, loadOutline(t, root), "beta")
	m = press(m, "N", "N")
	m = press(m, "text", "note")
	if err := os.Remove(filepath.Join(root, "beta")); err != nil {
		t.Fatal(err)
	}
	m = press(m, "enter", "")
	if !m.creating || m.message == "" {
		t.Fatal("removed parent must leave name editable")
	}
	if err := os.Mkdir(filepath.Join(root, "beta"), 0o700); err != nil {
		t.Fatal(err)
	}
	m = press(m, "enter", "")
	if m.creating || m.message != "" {
		t.Fatal("retry must work after the parent returns")
	}
}

func TestNewFilePromptAndLiteralInput(t *testing.T) {
	m := focus(t, loadBrowser(t, fixture(t)), "Alpha")
	m = press(m, "N", "N")
	m.newName = "note\x1b[31m\n.txt"
	frame := plainFrame(m.view())
	if !strings.Contains(frame, "New file:") || !strings.Contains(frame, "New in:") || !strings.Contains(frame, "Enter create") || strings.ContainsRune(frame, '\x1b') {
		t.Fatal("file prompt must show a safe name and destination")
	}
	m = press(m, "esc", "")
	m = press(m, "/", "/")
	m = press(m, "N", "N")
	if m.creating || m.query != "N" {
		t.Fatal("Shift+N must remain literal while filtering")
	}
}
