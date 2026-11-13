package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"Alpha/nested", "beta", ".hidden", "\u0420\u0443\u0441\u0441\u043a\u0438\u0439", "with spaces"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "regular-file"), []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func press(m browser, name, text string) browser {
	m.handle(keyEvent{name: name, text: text})
	return m
}

func loadBrowser(t *testing.T, root string) browser {
	t.Helper()
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func focus(t *testing.T, m browser, name string) browser {
	t.Helper()
	i := slices.IndexFunc(m.visible, func(d directory) bool { return d.name == name })
	if i < 0 {
		t.Fatalf("directory %q missing from %v", name, m.visible)
	}
	m.cursor = i
	return m
}

func TestBrowseAndSelect(t *testing.T) {
	root := fixture(t)
	m := focus(t, loadBrowser(t, root), "Alpha")
	m = press(m, "right", "")
	if m.path != filepath.Join(root, "Alpha") {
		t.Fatalf("opened %q", m.path)
	}
	m = press(m, "left", "")
	if m.path != root || m.visible[m.cursor].name != "Alpha" {
		t.Fatalf("parent did not preserve focus: %+v", m)
	}
	m = press(m, "enter", "")
	if m.selected != filepath.Join(root, "Alpha") {
		t.Fatalf("selected %q", m.selected)
	}
	m = press(loadBrowser(t, root), "enter", "")
	if m.selected != root {
		t.Fatalf("current directory selection = %q", m.selected)
	}
}

func TestFilterHiddenAndSymlinks(t *testing.T) {
	root := fixture(t)
	for name, target := range map[string]string{"link": "Alpha", "broken": "missing", "file-link": "regular-file"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	m := loadBrowser(t, root)
	var names []string
	for _, d := range m.visible {
		names = append(names, d.name)
	}
	want := []string{".", "Alpha", "beta", "link", "with spaces", "\u0420\u0443\u0441\u0441\u043a\u0438\u0439"}
	if !slices.Equal(names, want) {
		t.Fatalf("directories = %v, want %v", names, want)
	}
	m = press(m, ".", ".")
	m = focus(t, m, ".hidden")
	m = press(m, "/", "/")
	m = press(m, "\u0420\u0423\u0421", "\u0420\u0423\u0421")
	if len(m.visible) != 2 || m.visible[m.cursor].name != "\u0420\u0443\u0441\u0441\u043a\u0438\u0439" {
		t.Fatalf("unicode filter: %v", m.visible)
	}
	m = press(m, "backspace", "")
	if m.query != "\u0420\u0423" {
		t.Fatalf("backspace split a rune: %q", m.query)
	}
	m = press(m, "esc", "")
	if m.filtering || m.selected != "" {
		t.Fatalf("escape should finish filtering: %+v", m)
	}
	m = press(m, "enter", "")
	if m.selected != filepath.Join(root, "\u0420\u0443\u0441\u0441\u043a\u0438\u0439") {
		t.Fatalf("filtered selection = %q", m.selected)
	}
}

func TestCancellationAndFailedNavigation(t *testing.T) {
	root := fixture(t)
	for _, key := range []keyEvent{{name: "esc"}, {name: "q", text: "q"}, {name: "ctrl+c"}} {
		m := loadBrowser(t, root)
		if !m.handle(key) || m.selected != "" {
			t.Fatalf("cancellation failed for %v", key)
		}
	}
	m := focus(t, loadBrowser(t, root), "beta")
	if err := os.Remove(filepath.Join(root, "beta")); err != nil {
		t.Fatal(err)
	}
	m = press(m, "right", "")
	if m.path != root || m.message == "" {
		t.Fatalf("failed navigation lost state: %+v", m)
	}
	m = press(m, "enter", "")
	if m.selected != "" {
		t.Fatalf("selected a removed directory: %q", m.selected)
	}
}

func TestEmptyFilterAndSmallTerminal(t *testing.T) {
	m := loadBrowser(t, fixture(t))
	m = press(m, "/", "/")
	m = press(m, "no-match", "no-match")
	if len(m.visible) != 1 || m.cursor != 0 {
		t.Fatalf("no-match filter must allow selecting current directory: %+v", m)
	}
	m = press(m, "esc", "")
	m = press(m, "down", "")
	m.width, m.height = 12, 5
	lines := strings.Split(m.view(), "\n")
	if len(lines) > 5 {
		t.Fatalf("view overflows height: %v", lines)
	}
	for _, line := range lines {
		if len([]rune(line)) > 12 {
			t.Fatalf("view overflows width: %q", line)
		}
	}
	if got := display("bad\x1b[31m\nname"); strings.ContainsAny(got, "\x1b\n") {
		t.Fatalf("control characters leaked into view: %q", got)
	}
}

func TestKeyboardHelpPreservesNavigationAndFilterInput(t *testing.T) {
	m := loadBrowser(t, fixture(t))
	if m.handle(keyEvent{name: "?", text: "?"}) || !m.showHelp {
		t.Fatal("help should open without exiting")
	}
	m = press(m, "down", "")
	if m.cursor != 1 {
		t.Fatal("help should leave navigation available")
	}
	if m.handle(keyEvent{name: "esc"}) || m.showHelp || m.selected != "" {
		t.Fatal("Esc should close help without cancelling or selecting")
	}
	m = press(m, "?", "?")
	m = press(m, "?", "?")
	if m.showHelp {
		t.Fatal("question mark should toggle help off")
	}
	m = press(m, "?", "?")
	m = press(m, "/", "/")
	m = press(m, "?", "?")
	if m.showHelp || m.query != "?" || !m.filtering {
		t.Fatal("question mark must remain literal input in filter mode")
	}
}
