package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestFinderOpensSelectedDirWithoutNavigating(t *testing.T) {
	root := fixture(t)
	link := filepath.Join(root, "dir-link")
	if err := os.Symlink(filepath.Join(root, "Alpha"), link); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, path := range []string{root, filepath.Dir(root), filepath.Join(root, "Alpha"), link} {
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
				if path == filepath.Dir(root) {
					m.showRootTree()
				}
			}
			m.focusPath(path)
			if m.focusedPath() != path {
				t.Fatalf("fixture did not focus %q", path)
			}
			before := m
			m.openInFinder("darwin")
			if m.action != (externalAction{kind: "finder", path: path}) || m.message != "" {
				t.Fatal("Finder must receive the highlighted dir, including navigation rows")
			}
			if m.selected != "" || m.path != before.path || m.cursor != before.cursor || m.scroll != before.scroll || m.query != before.query {
				t.Fatal("opening Finder must preserve browsing and shell selection")
			}
			m.action = externalAction{}
			m.openInFinder("linux")
			if m.action.kind != "" || slices.ContainsFunc(m.entryActions("linux"), func(h keyHint) bool { return h.key == "f" }) {
				t.Fatal("Finder must be unavailable on Linux")
			}
		}
	}
}

func TestFinderRejectsFilesAndChangedDirs(t *testing.T) {
	root := fixture(t)
	for _, name := range []string{"regular-file", "broken-link"} {
		if name == "broken-link" {
			if err := os.Symlink("missing", filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
		}
		m := focus(t, loadBrowser(t, root), name)
		m.openInFinder("darwin")
		if m.action.kind != "" || slices.ContainsFunc(m.entryActions("darwin"), func(h keyHint) bool { return h.key == "f" }) {
			t.Fatal("Finder must not be offered for files or broken links")
		}
	}
	m := focus(t, loadBrowser(t, root), "Alpha")
	path := m.focusedPath()
	m.openInFinder("darwin")
	action := m.action
	m.action = externalAction{}
	if err := os.Rename(path, filepath.Join(root, "saved-dir")); err != nil {
		t.Fatal(err)
	}
	for _, replaced := range []bool{false, true} {
		if replaced {
			if err := os.WriteFile(path, nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		m.openInFinder("darwin")
		if m.action.kind != "" || m.message == "" || validateActionPath(action) == nil {
			t.Fatal("missing dirs and dirs replaced by files must not open Finder")
		}
	}
}

func TestFinderCommandPreservesLiteralPath(t *testing.T) {
	capture := fakeActionCLI(t, "open")
	path := filepath.Join(t.TempDir(), "-dir' $() `literal`; \u754c")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	action := externalAction{kind: "finder", path: path}
	if err := runExternalAction(t.Context(), action, "darwin"); err != nil {
		t.Fatal(err)
	}
	assertEditorArgs(t, capture, []string{"-b", "com.apple.finder", path})
	if _, _, err := actionCommand(action, "linux"); err == nil {
		t.Fatal("Finder must not launch on Linux")
	}
	if _, _, err := actionCommand(externalAction{kind: "finder", path: "relative"}, "darwin"); err == nil {
		t.Fatal("relative paths must not launch Finder")
	}
	t.Setenv("CDD_TEST_ACTION", "fail")
	if err := runExternalAction(t.Context(), action, "darwin"); err == nil || len(err.Error()) > 4300 {
		t.Fatal("Finder failures must report bounded diagnostics")
	}
}

func TestFinderButtonAndShortcut(t *testing.T) {
	root := fixture(t)
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, width := range []int{20, 40, 80} {
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, "Alpha")
			m.width, m.height = width, 24
			path := m.focusedPath()
			found := false
			for y := 0; y < m.height && !found; y++ {
				for x := 0; x < m.width; x++ {
					if m.entryButtonAt(x, y) == "f" {
						if m.handle(keyEvent{name: "mouse-left", x: x, y: y}) {
							t.Fatal("Finder click exited cdd")
						}
						found = true
						break
					}
				}
			}
			if found != (runtime.GOOS == "darwin") {
				t.Fatal("Finder button visibility does not match the platform")
			}
			clicked := m.action
			m.action = externalAction{}
			m = press(m, "f", "f")
			if m.action != clicked || m.selected != "" || m.focusedPath() != path {
				t.Fatal("Finder button and shortcut must preserve the selected dir")
			}
			if runtime.GOOS == "darwin" && m.action != (externalAction{kind: "finder", path: path}) {
				t.Fatal("Finder button did not queue the selected dir")
			}
		}
	}
	for _, input := range []string{"/", "n", "N", "r"} {
		m := press(focus(t, loadBrowser(t, root), "Alpha"), input, input)
		m = press(m, "f", "f")
		if m.action.kind != "" || len(m.entryButtons()) != 0 {
			t.Fatal("Finder action must not interrupt text input")
		}
		if input == "r" && m.mutation.input != "f" || input == "/" && m.query != "f" || (input == "n" || input == "N") && m.newName != "f" {
			t.Fatal("f must remain literal during text input")
		}
	}
	m := focus(t, loadBrowser(t, root), "Alpha")
	m.width, m.height = 80, 2
	if strings.Contains(m.view(), "[ Finder") {
		t.Fatal("clipped button must not be rendered")
	}
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.entryButtonAt(x, y) == "f" {
				t.Fatal("clipped button must not be clickable")
			}
		}
	}
}
