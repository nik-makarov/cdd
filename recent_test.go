package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRecentPlacesOrderLimitAndLiteralPaths(t *testing.T) {
	file := filepath.Join(t.TempDir(), "cdd", "recent-dirs")
	want := make([]string, 0, recentLimit+1)
	for i := range recentLimit + 1 {
		path := fmt.Sprintf("/projects/%02d", i)
		if err := recordRecent(file, path); err != nil {
			t.Fatal(err)
		}
		want = append(want, path)
	}
	slices.Reverse(want)
	want = want[:recentLimit]
	assertPlaces := func() {
		t.Helper()
		got, err := readRecent(file)
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("recent places: %q, want %q (%v)", got, want, err)
		}
	}
	assertPlaces()
	revisit := want[20]
	if err := recordRecent(file, revisit+"/."); err != nil {
		t.Fatal(err)
	}
	want = append([]string{revisit}, slices.Delete(want, 20, 21)...)
	assertPlaces()
	literal := "/projects/'$()\n\u754c\xff"
	if err := recordRecent(file, literal); err != nil {
		t.Fatal(err)
	}
	want = append([]string{literal}, want[:recentLimit-1]...)
	assertPlaces()
	if info, err := os.Stat(file); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("expected a private history file: %v, %v", info, err)
	}
}

func TestConcurrentRecentSelections(t *testing.T) {
	file := filepath.Join(t.TempDir(), "recent-dirs")
	const sessions = 12
	done := make(chan error, sessions)
	for i := range sessions {
		go func() { done <- recordRecent(file, fmt.Sprintf("/projects/%d", i)) }()
	}
	for range sessions {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	got, err := readRecent(file)
	if err != nil || len(got) != sessions {
		t.Fatalf("lost a concurrent selection: %q, %v", got, err)
	}
}

func TestRecentPathAnchors(t *testing.T) {
	const home = "/users/ann"
	for _, tc := range []struct {
		path  string
		width int
		want  string
	}{
		{"/", 20, "/"}, {home, 20, "~"}, {home + "/src", 20, "~/src"},
		{"/users/anna", 20, "/users/anna"},
		{home + "/very-long-parent/project", 13, "~/.../project"},
		{"/other/very-long-parent/project", 12, "/.../project"},
		{home + "/long", 1, "~"}, {"/other", 1, "/"},
		{home + "/\x1bdir", 12, "~/\ufffddir"},
	} {
		a := addressFor(tc.path, home)
		if got := a.label(tc.width); got != tc.want {
			t.Errorf("%q at width %d = %q, want %q", tc.path, tc.width, got, tc.want)
		}
	}
	a := addressFor(home+"/very-long-parent/project", home)
	if a.breadcrumb(0, 13) != home || a.breadcrumb(2, 13) != "" || a.breadcrumb(12, 13) != home+"/very-long-parent/project" {
		t.Fatal("abbreviated header lost its literal breadcrumb targets")
	}
}

func TestRecentPanelNavigationAndResize(t *testing.T) {
	root := fixture(t)
	m := focus(t, loadOutline(t, root), "Alpha")
	m.width, m.height = 80, 18
	m.recent.paths = []string{filepath.Join(root, "beta"), "/", filepath.Join(root, "Alpha")}
	treePath := m.focusedPath()
	view := plainFrame(m.view())
	if !strings.Contains(view, "Recent places") || !strings.Contains(view, "~/beta") || !strings.Contains(view, "Alpha") {
		t.Fatal("expected both panels at 80 columns", view)
	}
	width, _ := m.viewWidth()
	if left, right := m.outlineWidths(width); left != 42 || right != 30 {
		t.Fatalf("split at 80 columns: %d/%d", left, right)
	}
	m = press(m, "tab", "")
	m = press(m, "end", "")
	for _, key := range []string{"trash", "copy-entry", "paste-entry", "n", "r"} {
		m = press(m, key, "")
		if m.mutation != nil || m.paste != nil || m.creating || m.action.kind != "" {
			t.Fatalf("%s acted on the tree while history had focus", key)
		}
	}
	m.width = 79
	if view := plainFrame(m.view()); !strings.Contains(view, "> ~/Alpha") || strings.Contains(view, "(starting dir)") {
		t.Fatal("resizing did not preserve active history", view)
	}
	m = press(m, "backtab", "")
	if m.focusedPath() != treePath || strings.Contains(plainFrame(m.view()), "~/beta") {
		t.Fatal("returning to a narrow tree lost its selection or showed history")
	}
	m = press(m, "tab", "")
	m = press(m, "home", "")
	if !m.handle(keyEvent{name: "enter"}) || m.selected != filepath.Join(root, "beta") {
		t.Fatal("Enter must choose the absolute history path")
	}
}

func TestRecentPagesMouseAndStalePaths(t *testing.T) {
	root := fixture(t)
	m := loadOutline(t, root)
	for i := range recentLimit {
		m.recent.paths = append(m.recent.paths, filepath.Join(root, fmt.Sprintf("place-%02d", i)))
	}
	m.width, m.height = 100, 14
	m = press(m, "tab", "")
	m = press(m, "pgdown", "")
	if m.recent.cursor != m.pageSize() {
		t.Fatal("Page Down did not move by one viewport")
	}
	m = press(m, "end", "")
	if view := plainFrame(m.view()); !strings.Contains(view, "> ~/place-49") {
		t.Fatal("last recent place was not reachable", view)
	}
	if m.handle(keyEvent{name: "enter"}) || m.selected != "" || m.message == "" {
		t.Fatal("stale path returned to the shell")
	}
	m.recent.paths[recentLimit-1] = filepath.Join(root, "regular-file")
	if m.handle(keyEvent{name: "confirm"}) || m.selected != "" {
		t.Fatal("replaced directory returned a file to the shell")
	}
	m = press(m, "home", "")
	m = press(m, "tab", "")
	m.recent.paths[1] = filepath.Join(root, "beta")
	width, margin := m.viewWidth()
	left, _ := m.outlineWidths(width)
	header, _ := m.frame()
	key := keyEvent{name: "mouse-left", x: margin + left + 3, y: len(header) + 1, when: time.Unix(100, 0)}
	if m.handle(key) || !m.recent.focused || m.recent.cursor != 1 {
		t.Fatal("a single click must focus the visible history row")
	}
	key.when = key.when.Add(100 * time.Millisecond)
	if !m.handle(key) || m.selected != filepath.Join(root, "beta") {
		t.Fatal("a double click must choose the clicked directory")
	}
}
