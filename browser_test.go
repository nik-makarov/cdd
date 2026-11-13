package main

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
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
	// Existing directory-only cases exercise the compact list mode.
	m.compact = true
	m.refilter()
	return m
}

func focus(t *testing.T, m browser, name string) browser {
	t.Helper()
	i := slices.IndexFunc(m.visible, func(d entry) bool { return d.name == name })
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

func TestColumnNavigation(t *testing.T) {
	root := fixture(t)
	for i := range 40 {
		if err := os.Mkdir(filepath.Join(root, "dir-"+strconv.Itoa(i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	m := loadBrowser(t, root)
	m.width, m.height = 80, 12
	rows, page := m.pageRows(), m.pageSize()
	for _, tt := range []struct {
		name, key   string
		start, want int
	}{
		{"switch right", "tab", 2, rows + 2},
		{"switch left", "tab", rows + 2, 2},
		{"reverse switch", "backtab", 2, rows + 2},
		{"reverse return", "backtab", rows + 2, 2},
		{"down enters second column", "down", rows - 1, rows},
		{"up returns to first column", "up", rows, rows - 1},
		{"next page", "pgdown", 2, page + 2},
		{"previous page", "pgup", page + 2, 2},
		{"first entry", "home", page + 2, 0},
		{"last entry", "end", 0, len(m.visible) - 1},
		{"empty other column", "tab", (len(m.visible) - 1) / page * page, (len(m.visible) - 1) / page * page},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := m
			m.cursor = tt.start
			m = press(m, tt.key, "")
			if m.cursor != tt.want || m.path != root {
				t.Fatalf("cursor %d, path %q; want cursor %d in %q", m.cursor, m.path, tt.want, root)
			}
		})
	}
	partial := m
	partial.visible = partial.visible[:page+rows+2]
	partial.cursor = page + rows - 1
	partial = press(partial, "tab", "")
	if partial.cursor != len(partial.visible)-1 {
		t.Fatal("switching to a shorter column should select its last entry")
	}
	m.cursor = 2
	m = press(m, "tab", "")
	wantPath := m.visible[m.cursor].path
	selected := press(m, "enter", "")
	if selected.selected != wantPath {
		t.Fatal("right column selected the wrong directory")
	}
	m = press(m, "right", "")
	if m.path != wantPath {
		t.Fatal("Right must still open the directory")
	}
	m = press(m, "left", "")
	if m.path != root || m.visible[m.cursor].path != wantPath {
		t.Fatal("Left must restore parent focus")
	}
	m = press(m, "/", "/")
	m = press(m, "dir-", "dir-")
	before, query := m.cursor, m.query
	m = press(m, "tab", "")
	if m.cursor != before+m.pageRows() || m.query != query || !m.filtering {
		t.Fatal("filter-mode Tab should switch columns without changing the query")
	}
	m.width = 40
	before = m.cursor
	m = press(m, "tab", "")
	if m.cursor != before {
		t.Fatal("Tab in a single column should keep selection")
	}
}

func TestDefaultTreeViewAndModeSwitch(t *testing.T) {
	root := fixture(t)
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if m.compact {
		t.Fatal("tree must be the default view")
	}
	m = focus(t, m, "regular-file")
	for _, key := range []string{"right", "l", "enter"} {
		t.Run("file/"+key, func(t *testing.T) {
			copy := m
			if copy.handle(keyEvent{name: key}) || copy.path != root || copy.selected != "" || copy.message == "" {
				t.Fatalf("file must remain highlighted without navigation or selection: %+v", copy)
			}
			copy = press(copy, "/", "/")
			if copy.message != "" {
				t.Fatal("a new filter should clear the previous file-action message")
			}
		})
	}
	m = press(m, "v", "v")
	if !m.compact || m.visible[m.cursor].path != root {
		t.Fatal("switching from a file should focus current directory in compact mode")
	}
	for _, item := range m.visible {
		if !item.isDir() {
			t.Fatal("compact mode must retain directory-only behavior")
		}
	}
	m = focus(t, m, "Alpha")
	m = press(m, "v", "v")
	if m.compact || m.visible[m.cursor].name != "Alpha" {
		t.Fatal("mode switch lost directory focus")
	}
	m = press(m, "tab", "")
	if m.visible[m.cursor].name != "Alpha" {
		t.Fatal("tree view must keep the middle column active")
	}
	m = press(m, "/", "/")
	m = press(m, "v", "v")
	if m.compact || m.query != "v" || !m.filtering {
		t.Fatal("v must remain literal while filtering")
	}
}

func TestTreeContextAndSymlinkNavigation(t *testing.T) {
	root := fixture(t)
	if err := os.Symlink("Alpha", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	m = focus(t, m, "link")
	m.refreshPreview(false)
	if len(m.children) != 1 || m.children[0].name != "nested" || m.preview != filepath.Join(root, "link") {
		t.Fatalf("link preview missing: %+v", m.children)
	}
	selected := press(m, "enter", "")
	if selected.selected != filepath.Join(root, "link") {
		t.Fatal("directory symlink must be selectable using its logical path")
	}
	m = press(m, "right", "")
	if m.path != filepath.Join(root, "link") {
		t.Fatal("directory symlink was not opened")
	}
	if !slices.ContainsFunc(m.parent, func(item entry) bool { return item.path == m.path }) {
		t.Fatal("parent pane lost the current directory")
	}
	m = press(m, "left", "")
	if m.path != root || m.visible[m.cursor].name != "link" {
		t.Fatal("logical parent or focus was not preserved")
	}
	m = focus(t, m, "regular-file")
	m.refreshPreview(false)
	if len(m.children) != 0 || m.childErr == "" {
		t.Fatal("files must not expose a directory preview")
	}
	m = focus(t, m, "Alpha")
	m.refreshPreview(false)
	if err := os.RemoveAll(filepath.Join(root, "Alpha")); err != nil {
		t.Fatal(err)
	}
	m.refreshPreview(true)
	if m.childErr == "" || m.path != root || m.selected != "" {
		t.Fatal("preview failure must stay in the side pane without changing location")
	}
	m = press(m, "ctrl+r", "")
	if slices.ContainsFunc(m.visible, func(item entry) bool { return item.name == "Alpha" }) {
		t.Fatal("refresh retained a removed directory")
	}
}
