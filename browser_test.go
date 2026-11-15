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
	m = press(m, "left", "")
	if m.path != root || m.visible[m.cursor].name != "Alpha" {
		t.Fatalf("parent did not preserve focus: %+v", m)
	}
	m = press(m, "confirm", "")
	if m.selected != filepath.Join(root, "Alpha") {
		t.Fatalf("selected %q", m.selected)
	}
	m = press(loadBrowser(t, root), "enter", "")
	if m.selected != root {
		t.Fatalf("current directory selection = %q", m.selected)
	}
}

func TestParentEntry(t *testing.T) {
	for _, compact := range []bool{false, true} {
		root := fixture(t)
		m := loadBrowser(t, root)
		m.compact = compact
		for _, query := range []string{"", "Alpha", "missing"} {
			for _, hidden := range []bool{false, true} {
				m.query, m.hidden = query, hidden
				m.refilter()
				if m.visible[0].name != ".." || m.visible[1].name != "." || m.visible[0].path != filepath.Dir(root) {
					t.Fatal("parent must precede current and both rows must bypass filters")
				}
				if query != "Alpha" && m.visible[m.cursor].name != "." {
					t.Fatal("current directory must remain the default selection")
				}
				if query == "Alpha" && m.visible[m.cursor].name != "Alpha" {
					t.Fatal("filter must focus the first matching child, not the parent")
				}
			}
		}
		m = focus(t, m, "..")
		if !m.handle(keyEvent{name: "confirm"}) || m.selected != filepath.Dir(root) {
			t.Fatal("confirmation on the parent row must choose the parent directory")
		}
		for _, key := range []string{"right", "l", "enter"} {
			m = loadBrowser(t, filepath.Join(root, "Alpha"))
			m.compact = compact
			m = focus(t, m, "..")
			if m.handle(keyEvent{name: key}) || m.path != root || m.visible[m.cursor].name != "Alpha" || m.selected != "" {
				t.Fatal("opening .. must restore parent focus without completing selection")
			}
		}
	}
	m := loadBrowser(t, string(filepath.Separator))
	if slices.ContainsFunc(m.visible, func(item entry) bool { return item.name == ".." }) {
		t.Fatal("filesystem root must not offer a parent row")
	}
	for _, key := range []string{"left", "h", "backspace"} {
		if m.handle(keyEvent{name: key}) || m.path != string(filepath.Separator) || m.cursor < 0 || m.selected != "" || m.message != "" {
			t.Fatal("parent navigation at the filesystem root must remain in the picker")
		}
	}
}

func TestCreateDirectory(t *testing.T) {
	for _, mode := range []string{"tree", "list"} {
		for _, tt := range []struct{ name, directory string }{
			{"plain", "new-directory"},
			{"spaces", " with spaces "},
			{"unicode", "\u041d\u043e\u0432\u0430\u044f \u65e5\u672c\u8a9e"},
			{"hidden", ".new-hidden"},
			{"leading dash", "-new"},
			{"backslash", `back\slash`},
			{"metacharacters", "quote'\" $() `touch INJECTED` ; * [x]"},
			{"shortcut letters", "nqvjhkl.?~"},
		} {
			t.Run(mode+"/"+tt.name, func(t *testing.T) {
				root := fixture(t)
				m := loadBrowser(t, root)
				m.compact = mode == "list"
				m = press(m, "/", "/")
				m = press(m, "text", "Alpha")
				m = press(m, "enter", "")
				m = press(m, "?", "?")
				m = press(m, "n", "n")
				if !m.creating || m.showHelp || m.query != "Alpha" {
					t.Fatal("new-directory input must close help and preserve the listing until confirmation")
				}
				for _, r := range tt.directory {
					m = press(m, string(r), string(r))
				}
				target := filepath.Join(root, tt.directory)
				if _, err := os.Stat(target); !os.IsNotExist(err) {
					t.Fatal("typing a name must not create the directory")
				}
				if m.handle(keyEvent{name: "enter"}) {
					t.Fatal("creating a directory must keep the picker open")
				}
				info, err := os.Stat(target)
				if err != nil || !info.IsDir() {
					t.Fatalf("directory was not created: %v", err)
				}
				if m.creating || m.newName != "" || m.path != root || m.query != "" || m.message != "" || m.selected != "" || m.visible[m.cursor].path != target {
					t.Fatalf("new directory is not ready to choose: %+v", m)
				}
				if strings.HasPrefix(tt.directory, ".") && !m.hidden {
					t.Fatal("a newly created hidden directory must become visible")
				}
				if !m.compact && (m.preview != target || len(m.children) != 0 || m.childErr != "") {
					t.Fatal("preview did not refresh to the new empty directory")
				}
				if !m.handle(keyEvent{name: "confirm"}) || m.selected != target {
					t.Fatal("confirmation must select the new directory")
				}
				if _, err := os.Stat(filepath.Join(root, "INJECTED")); !os.IsNotExist(err) {
					t.Fatal("directory name was interpreted as a shell command")
				}
			})
		}
	}
}

func TestCreateDirectoryCancellation(t *testing.T) {
	for _, key := range []string{"esc", "ctrl+c", "ctrl+d"} {
		t.Run(key, func(t *testing.T) {
			root := fixture(t)
			m := loadBrowser(t, root)
			m.query = "Alpha"
			m.refilter()
			cursor := m.cursor
			m = press(m, "n", "n")
			m = press(m, "text", "cancelled")
			if done := m.handle(keyEvent{name: key}); done != (key != "esc") {
				t.Fatalf("unexpected exit behavior for %s", key)
			}
			if m.path != root || m.cursor != cursor || m.query != "Alpha" || m.hidden || m.selected != "" {
				t.Fatal("cancellation changed the browsing state")
			}
			if key == "esc" && (m.creating || m.newName != "") {
				t.Fatal("Esc did not dismiss the name input")
			}
			if _, err := os.Stat(filepath.Join(root, "cancelled")); !os.IsNotExist(err) {
				t.Fatal("cancellation created a directory")
			}
		})
	}
}

func TestCreateDirectoryRejectsInvalidNamesAndExistingEntries(t *testing.T) {
	for _, tt := range []struct{ name, directory string }{
		{"empty", ""}, {"current", "."}, {"parent", ".."},
		{"nested", "Alpha/nested-new"}, {"outside", "../outside"},
		{"trailing slash", "Alpha/"}, {"nul", "bad\x00name"},
		{"existing directory", "Alpha"}, {"existing file", "regular-file"},
		{"directory link", "directory-link"}, {"file link", "file-link"},
		{"broken link", "broken-link"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := fixture(t)
			for name, target := range map[string]string{"directory-link": "Alpha", "file-link": "regular-file", "broken-link": "missing"} {
				if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
					t.Fatal(err)
				}
			}
			m := focus(t, loadBrowser(t, root), "Alpha")
			cursor := m.cursor
			m = press(m, "n", "n")
			m = press(m, "text", tt.directory)
			if m.handle(keyEvent{name: "enter"}) || !m.creating || m.message == "" || m.newName != tt.directory || m.path != root || m.cursor != cursor || m.selected != "" {
				t.Fatalf("invalid creation did not retain editable input: %+v", m)
			}
			content, err := os.ReadFile(filepath.Join(root, "regular-file"))
			if err != nil || string(content) != "text" {
				t.Fatal("existing file was changed")
			}
			for _, path := range []string{filepath.Join(root, "Alpha", "nested-new"), filepath.Join(root, "..", "outside"), filepath.Join(root, "missing")} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("invalid name created %q", path)
				}
			}
			m = press(m, "esc", "")
			if m.creating || m.message != "" {
				t.Fatal("Esc must dismiss failed creation")
			}
		})
	}
}

func TestCreateDirectoryEditingAndRetry(t *testing.T) {
	root := fixture(t)
	m := loadBrowser(t, root)
	m = press(m, "n", "n")
	m = press(m, "text", "Alpha")
	m = press(m, "enter", "")
	if m.message == "" {
		t.Fatal("existing directory must report a creation error")
	}
	m = press(m, "text", "\u754c")
	if m.message != "" {
		t.Fatal("editing must clear the previous error")
	}
	m = press(m, "backspace", "")
	if m.newName != "Alpha" {
		t.Fatal("Backspace split a Unicode rune")
	}
	for range 6 {
		m = press(m, "backspace", "")
	}
	if m.newName != "" || m.path != root {
		t.Fatal("Backspace on empty input must not navigate")
	}
	m = press(m, "text", "retry")
	for _, key := range []string{"up", "down", "left", "right", "home", "end", "pgup", "pgdown", "tab", "backtab", "ctrl+r"} {
		m = press(m, key, "")
	}
	if !m.creating || m.newName != "retry" || m.path != root {
		t.Fatal("navigation keys changed the creation input or location")
	}
	m = press(m, "enter", "")
	if m.creating || m.message != "" || m.visible[m.cursor].name != "retry" {
		t.Fatal("corrected name was not created and focused")
	}
}

func TestCreateDirectoryAfterParentDisappears(t *testing.T) {
	root := t.TempDir()
	m := loadBrowser(t, root)
	m = press(m, "n", "n")
	m = press(m, "text", "child")
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	m = press(m, "enter", "")
	if !m.creating || m.message == "" || m.selected != "" {
		t.Fatal("filesystem error must keep the name editable")
	}
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	m = press(m, "enter", "")
	if m.creating || m.message != "" || m.visible[m.cursor].name != "child" {
		t.Fatal("creation could not recover after the parent returned")
	}
}

func TestCreateDirectoryThroughLogicalSymlink(t *testing.T) {
	base := t.TempDir()
	real, link := filepath.Join(base, "real"), filepath.Join(base, "link")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	m := loadBrowser(t, link)
	m = press(m, "n", "n")
	m = press(m, "text", "child")
	m = press(m, "enter", "")
	m = press(m, "confirm", "")
	if m.selected != filepath.Join(link, "child") {
		t.Fatal("creation lost the logical symlink path")
	}
	if info, err := os.Stat(filepath.Join(real, "child")); err != nil || !info.IsDir() {
		t.Fatalf("directory missing from symlink target: %v", err)
	}
}

func TestLeftFocusesParentBeforeOpening(t *testing.T) {
	for _, key := range []string{"left", "h", "backspace"} {
		for _, compact := range []bool{false, true} {
			mode := "tree"
			if compact {
				mode = "list"
			}
			for _, target := range []string{".", "Alpha", "regular-file", "link"} {
				if compact && target == "regular-file" {
					continue
				}
				t.Run(key+"/"+mode+"/"+target, func(t *testing.T) {
					root := fixture(t)
					if err := os.Symlink("Alpha", filepath.Join(root, "link")); err != nil {
						t.Fatal(err)
					}
					m, err := newBrowser(root, false)
					if err != nil {
						t.Fatal(err)
					}
					m.compact = compact
					m.refilter()
					m = focus(t, m, target)
					m.refreshPreview(false)
					m.message = "Previous error"
					if m.handle(keyEvent{name: key}) || m.path != root || m.visible[m.cursor].name != ".." || m.message != "" || m.selected != "" {
						t.Fatalf("first %s must focus .. without leaving or selecting: path %q, cursor %d", key, m.path, m.cursor)
					}
					if !compact && (len(m.children) != 0 || m.childErr != "" || m.preview != filepath.Dir(root)) {
						t.Fatal("focusing .. must clear the child preview")
					}
					if m.handle(keyEvent{name: key}) || m.path != filepath.Dir(root) || m.visible[m.cursor].path != root {
						t.Fatalf("second %s must open the parent and focus the previous directory", key)
					}
				})
			}
		}
	}
}

func TestLeftPreservesFilterUntilParentNavigation(t *testing.T) {
	for _, mode := range []string{"tree", "list"} {
		t.Run(mode, func(t *testing.T) {
			root := fixture(t)
			m := loadBrowser(t, root)
			m.compact = mode == "list"
			m = press(m, "/", "/")
			m = press(m, "text", "Al")
			m = press(m, "esc", "")
			m = press(m, "left", "")
			if m.path != root || m.visible[m.cursor].name != ".." || m.query != "Al" || len(m.visible) != 3 {
				t.Fatal("focusing .. must preserve the completed filter and listing")
			}
			opened := press(m, "enter", "")
			if opened.path != filepath.Dir(root) || opened.selected != "" || opened.query != "" {
				t.Fatal("Enter after Left must browse the parent directory")
			}
			m = press(m, "left", "")
			if m.path != filepath.Dir(root) || m.query != "" {
				t.Fatal("opening the parent must clear the old directory's filter")
			}
		})
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
	want := []string{"..", ".", "Alpha", "beta", "link", "with spaces", "\u0420\u0443\u0441\u0441\u043a\u0438\u0439"}
	if !slices.Equal(names, want) {
		t.Fatalf("directories = %v, want %v", names, want)
	}
	m = press(m, ".", ".")
	m = focus(t, m, ".hidden")
	m = press(m, "/", "/")
	m = press(m, "\u0420\u0423\u0421", "\u0420\u0423\u0421")
	if len(m.visible) != 3 || m.visible[m.cursor].name != "\u0420\u0443\u0441\u0441\u043a\u0438\u0439" {
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
	m = press(m, "confirm", "")
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
	m = press(m, "confirm", "")
	if m.selected != "" {
		t.Fatalf("selected a removed directory: %q", m.selected)
	}
}

func TestEmptyFilterAndSmallTerminal(t *testing.T) {
	m := loadBrowser(t, fixture(t))
	m = press(m, "/", "/")
	m = press(m, "no-match", "no-match")
	if len(m.visible) != 2 || m.visible[m.cursor].name != "." || m.visible[0].name != ".." {
		t.Fatalf("no-match filter must allow selecting current directory: %+v", m)
	}
	if m.handle(keyEvent{name: "enter"}) || m.filtering || m.selected != "" {
		t.Fatal("Enter while filtering must finish input without selecting .")
	}
	current := m
	if !current.handle(keyEvent{name: "enter"}) || current.selected != m.path {
		t.Fatal("Enter after finishing the filter must confirm .")
	}
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
	if m.visible[m.cursor].name != "Alpha" {
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
	m = press(m, "n", "n")
	if m.creating || m.query != "?n" || !m.filtering {
		t.Fatal("n must remain literal input in filter mode")
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
	m.width, m.height = 80, 16
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
	partial.visible = partial.visible[:page+rows]
	partial.cursor = page
	partial = press(partial, "tab", "")
	if partial.cursor != page {
		t.Fatal("switching to an empty column must preserve selection")
	}
	partial = m
	partial.visible = partial.visible[:page+rows+2]
	partial.cursor = page + rows - 1
	partial = press(partial, "tab", "")
	if partial.cursor != len(partial.visible)-1 {
		t.Fatal("switching to a shorter column should select its last entry")
	}
	m.cursor = 2
	m = press(m, "tab", "")
	wantPath := m.visible[m.cursor].path
	parent := press(m, "left", "")
	if parent.path != root || parent.visible[parent.cursor].name != ".." {
		t.Fatal("Left from the second column must focus .. in the same directory")
	}
	selected := press(m, "confirm", "")
	if selected.selected != wantPath {
		t.Fatal("right column selected the wrong directory")
	}
	m = press(m, "right", "")
	if m.path != wantPath {
		t.Fatal("Right must still open the directory")
	}
	m = press(m, "left", "")
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
	for _, key := range []string{"right", "l", "enter", "confirm"} {
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
	selected := press(m, "confirm", "")
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
	m = focus(t, m, "nested")
	m = press(m, "left", "")
	if m.path != filepath.Join(root, "link") || m.visible[m.cursor].name != ".." {
		t.Fatal("first Left must focus .. without resolving or leaving the directory link")
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

func TestTreePreviewClearsOnCurrentDirectory(t *testing.T) {
	root := fixture(t)
	if err := os.Mkdir(filepath.Join(root, "Alpha", "nested", "leaf"), 0o755); err != nil {
		t.Fatal(err)
	}
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	checkEmpty := func(name string) {
		t.Helper()
		if m.visible[m.cursor].name != "." || len(m.children) != 0 || m.childErr != "" {
			t.Errorf("%s: current-directory row must clear Children: %+v, %q", name, m.children, m.childErr)
		}
	}
	checkEmpty("startup")
	m = focus(t, m, "Alpha")
	m.refreshPreview(false)
	if len(m.children) != 1 || m.children[0].name != "nested" {
		t.Fatal("highlighted directory must still show its children")
	}
	m = press(m, "right", "")
	if m.path != filepath.Join(root, "Alpha") {
		t.Fatal("directory was not opened")
	}
	checkEmpty("enter directory")
	m = press(m, "up", "")
	if m.visible[m.cursor].name != ".." || len(m.children) != 0 {
		t.Fatal("parent row must leave the children preview empty")
	}
	m = press(m, "down", "")
	checkEmpty("return from parent row")
	m = press(m, "down", "")
	if len(m.children) != 1 || m.children[0].name != "leaf" {
		t.Fatal("highlighting a child directory must populate its preview")
	}
	m = press(m, "left", "")
	if m.visible[m.cursor].name != ".." || len(m.children) != 0 {
		t.Fatal("Left must focus the parent and clear the children preview")
	}
	m = press(m, "down", "")
	checkEmpty("return to current row")
	m = press(m, "ctrl+r", "")
	checkEmpty("refresh")
	m = press(m, "v", "v")
	m = press(m, "v", "v")
	checkEmpty("view round trip")
}
