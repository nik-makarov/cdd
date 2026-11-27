package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func loadOutline(t *testing.T, root string) browser {
	t.Helper()
	t.Setenv("HOME", root)
	m, err := newBrowser(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.prepareOutline(root); err != nil {
		t.Fatal(err)
	}
	return m
}

func visiblePath(m browser, path string) bool {
	return slices.ContainsFunc(m.visible, func(item entry) bool { return item.path == path })
}

func TestOutlineRootButtonRevealsAncestors(t *testing.T) {
	root := fixture(t)
	m := loadOutline(t, root)
	m = focus(t, m, "Alpha")
	m = press(m, "right", "")
	m = press(m, "right", "")
	focus := m.focusedPath()
	width, margin := m.viewWidth()
	if !strings.Contains(plainFrame(m.view()), rootButton) {
		t.Fatal("root button must be visible in the rendered header")
	}
	if m.handle(keyEvent{name: "mouse-left", x: margin + width - 2, y: 0}) {
		t.Fatal("root button must not confirm a destination")
	}
	if m.treeRoot != "/" || m.focusedPath() != focus || !visiblePath(m, filepath.Dir(root)) || m.origin != root || m.selected != "" {
		t.Fatal("root button must reveal ancestors and retain the nested selection")
	}
	m = press(m, "~", "~")
	if m.treeRoot != "/" || m.path != root {
		t.Fatal("home navigation must retain the expanded root scope")
	}
	m = loadOutline(t, root)
	m = press(m, "\\", "\\")
	if m.treeRoot != "/" || m.focusedPath() != root {
		t.Fatal("keyboard root action must retain home focus")
	}
	m.height = 7
	if m.rootButtonAt(margin+width-2, 0) {
		t.Fatal("hidden root button must not accept clicks")
	}
}

func TestOutlineNavigationAndConfirmation(t *testing.T) {
	root := fixture(t)
	m := loadOutline(t, root)
	if m.mode != outlineView || m.focusedPath() != root || m.treeRoot != root || len(m.branches) != 1 {
		t.Fatal("tree must focus the starting dir and load its home root")
	}
	alpha, nested := filepath.Join(root, "Alpha"), filepath.Join(root, "Alpha", "nested")
	m = focus(t, m, "Alpha")
	m = press(m, "right", "")
	if m.path != root || m.focusedPath() != alpha || !visiblePath(m, nested) || m.selected != "" {
		t.Fatal("Right must expand in place without moving focus or choosing a destination")
	}
	m = press(m, "right", "")
	if m.focusedPath() != nested {
		t.Fatal("Right on an expanded branch must focus its first child")
	}
	m = press(m, "left", "")
	if m.focusedPath() != alpha || !visiblePath(m, nested) {
		t.Fatal("Left on a collapsed node must focus its parent")
	}
	m = press(m, "left", "")
	if m.focusedPath() != alpha || visiblePath(m, nested) {
		t.Fatal("Left on an expanded branch must collapse it without moving focus")
	}
	m = press(m, "right", "")
	m = press(m, "right", "")
	if m.focusedPath() != nested || m.selected != "" {
		t.Fatal("Right must expand and then focus a child")
	}
	if !m.handle(keyEvent{name: "enter"}) || m.selected != nested {
		t.Fatal("confirmation must return the nested logical destination")
	}
}

func TestOutlineModeOrderAndNestedFocus(t *testing.T) {
	root := fixture(t)
	m := loadOutline(t, root)
	m = focus(t, m, "Alpha")
	m = press(m, "right", "")
	m = press(m, "right", "")
	nested := m.focusedPath()
	for _, mode := range []viewMode{panesView, outlineView} {
		m = press(m, "v", "v")
		if m.mode != mode || m.focusedPath() != nested || m.selected != "" {
			t.Fatalf("switching to %s lost nested focus", mode)
		}
	}
	// Cycling without changing the tree root retains expanded branches.
	m = focus(t, m, "nested")
	m = press(m, "right", "")
	for range 2 {
		m = press(m, "v", "v")
	}
	if _, loaded := m.branches[nested]; !loaded || m.mode != outlineView {
		t.Fatal("view cycling discarded the cached branch")
	}
}

func TestOutlineFilterHiddenAndRefresh(t *testing.T) {
	root := fixture(t)
	m := focus(t, loadOutline(t, root), "Alpha")
	m = press(m, "right", "")
	m = press(m, "/", "/")
	m = press(m, "text", "NEST")
	for _, name := range []string{filepath.Base(root), "Alpha", "nested"} {
		if !slices.ContainsFunc(m.visible, func(item entry) bool { return item.name == name }) {
			t.Fatalf("filter lost matching branch or ancestor %q", name)
		}
	}
	if visiblePath(m, filepath.Join(root, "beta")) || len(m.branches) != 2 {
		t.Fatal("filter included a nonmatch or loaded a collapsed branch")
	}
	if m.visible[m.cursor].name != "nested" {
		t.Fatal("filter must focus its first matching entry, not a retained ancestor")
	}
	m = press(m, "esc", "")
	m = focus(t, m, "nested")
	m = press(m, "left", "")
	if m.query != "" || m.focusedPath() != filepath.Join(root, "Alpha") {
		t.Fatal("tree navigation must clear the filter and focus the parent")
	}
	m = press(m, ".", ".")
	if !visiblePath(m, filepath.Join(root, ".hidden")) {
		t.Fatal("hidden toggle did not expose hidden entries")
	}
	newPath := filepath.Join(root, "Alpha", "new-child")
	if err := os.Mkdir(newPath, 0o700); err != nil {
		t.Fatal(err)
	}
	m = press(m, "ctrl+r", "")
	if !visiblePath(m, newPath) || m.focusedPath() != filepath.Join(root, "Alpha") {
		t.Fatal("refresh must reload expanded branches while keeping focus")
	}
	m = focus(t, m, "new-child")
	if err := os.Remove(newPath); err != nil {
		t.Fatal(err)
	}
	m = press(m, "ctrl+r", "")
	if visiblePath(m, newPath) || m.focusedPath() != filepath.Join(root, "Alpha") {
		t.Fatal("refresh must move removed-node focus to its nearest visible ancestor")
	}
}

func TestOutlineLazyLinksAndGuards(t *testing.T) {
	root := fixture(t)
	for _, link := range []struct{ name, target string }{{"link", "Alpha"}, {"cycle", "."}} {
		if err := os.Symlink(link.target, filepath.Join(root, link.name)); err != nil {
			t.Fatal(err)
		}
	}
	m := focus(t, loadOutline(t, root), "link")
	m = press(m, "right", "")
	logical := filepath.Join(root, "link", "nested")
	if !visiblePath(m, logical) || len(m.branches) != 2 {
		t.Fatal("symlinks must retain their logical paths and load only the expanded level")
	}
	m = focus(t, m, "cycle")
	m = press(m, "right", "")
	if len(m.branches) != 3 || !visiblePath(m, filepath.Join(root, "cycle", "cycle")) {
		t.Fatal("a cyclic symlink must not trigger recursive traversal")
	}
	for _, key := range []string{"right", "confirm"} {
		t.Run("file/"+key, func(t *testing.T) {
			m := focus(t, loadOutline(t, root), "regular-file")
			if m.handle(keyEvent{name: key}) || m.selected != "" || m.path != root || m.message == "" {
				t.Fatal("tree file action bypassed the directory-only guard")
			}
		})
	}
	m = focus(t, m, "beta")
	if err := os.Remove(filepath.Join(root, "beta")); err != nil {
		t.Fatal(err)
	}
	m = press(m, "right", "")
	if m.message == "" || m.focusedPath() != filepath.Join(root, "beta") || m.selected != "" {
		t.Fatal("failed expansion must leave focus in place and report the error")
	}
	if m.handle(keyEvent{name: "confirm"}) || m.selected != "" {
		t.Fatal("a stale directory must not be confirmed")
	}
}

func TestOutlineParentCreationAndEmptyLeaf(t *testing.T) {
	root := fixture(t)
	m := focus(t, loadOutline(t, root), "Alpha")
	m = press(m, "n", "n")
	if !strings.Contains(plainFrame(m.view()), "New in:") {
		t.Fatal("creation must show which directory receives the new child")
	}
	m = press(m, "text", "new dir; $x")
	m = press(m, "enter", "")
	want := filepath.Join(root, "Alpha", "new dir; $x")
	if m.creating || m.path != root || m.focusedPath() != want {
		t.Fatal("creation must reveal and focus a child inside the highlighted directory")
	}
	if info, err := os.Stat(want); err != nil || !info.IsDir() {
		t.Fatal("nested directory was not created literally")
	}
	m = press(m, "right", "")
	if !strings.Contains(plainFrame(m.view()), "(empty)") || m.message != "" {
		t.Fatal("empty branches must show a quiet empty state")
	}
	m = press(m, "left", "")
	if m.focusedPath() != filepath.Join(root, "Alpha") {
		t.Fatal("Left on a known empty leaf must focus its parent")
	}
	m.focusPath(root)
	m = press(m, "left", "")
	m = press(m, "left", "")
	if m.treeRoot != root || m.focusedPath() != root || m.mode != outlineView {
		t.Fatal("Left on the home root must keep the tree anchored at home")
	}
}

func TestOutlineMouseDisclosureAndConfirmation(t *testing.T) {
	for _, filtering := range []bool{false, true} {
		t.Run(fmt.Sprintf("filter=%t", filtering), func(t *testing.T) {
			root := fixture(t)
			m := loadOutline(t, root)
			if filtering {
				m = press(m, "/", "/")
				m = press(m, "text", "Alpha")
			}
			click := mouseOn(t, m, "Alpha", "mouse-left", time.Unix(100, 0))
			if m.handle(click) || m.visible[m.cursor].name != "Alpha" || visiblePath(m, filepath.Join(root, "Alpha", "nested")) {
				t.Fatal("clicking a label must only focus it")
			}
			click.when = click.when.Add(200 * time.Millisecond)
			if m.handle(click) || !visiblePath(m, filepath.Join(root, "Alpha", "nested")) || m.filtering || m.selected != "" {
				t.Fatal("double click must expand in place and finish filtering")
			}
			click = mouseOn(t, m, "Alpha", "mouse-left", click.when.Add(time.Second))
			click.x -= 2
			if m.handle(click) || visiblePath(m, filepath.Join(root, "Alpha", "nested")) {
				t.Fatal("a single click on the disclosure must collapse the branch")
			}
			if !m.handle(keyEvent{name: "confirm"}) || m.selected != filepath.Join(root, "Alpha") {
				t.Fatal("tree confirmation must return the highlighted directory")
			}
			m = loadOutline(t, root)
			click = mouseOn(t, m, "(starting dir)", "mouse-left", time.Unix(200, 0))
			if m.handle(click) {
				t.Fatal("single click on current directory must not confirm")
			}
			click.when = click.when.Add(200 * time.Millisecond)
			if m.handle(click) || m.selected != "" || m.branches[root].expanded {
				t.Fatal("double click on the starting dir must toggle its branch")
			}
			if !m.handle(keyEvent{name: "enter"}) || m.selected != root {
				t.Fatal("Enter must confirm even the collapsed starting dir")
			}
		})
	}
}

func TestOutlineCentersSelectionWhileScrolling(t *testing.T) {
	root := t.TempDir()
	for i := range 24 {
		if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("dir-%02d", i)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	m := loadOutline(t, root)
	m.width = 160
	for _, height := range []int{15, 16, 40, 1} {
		t.Run(fmt.Sprintf("height=%d", height), func(t *testing.T) {
			m := m
			m.height = height
			for _, key := range []string{"down", "up"} {
				for i := range len(m.visible) {
					index := i
					if key == "up" {
						index = len(m.visible) - 1 - i
					}
					rows := m.pageRows()
					wantRow := (rows - 1) / 2
					if index < wantRow || len(m.visible) <= rows {
						wantRow = index
					} else if remaining := len(m.visible) - 1 - index; remaining < rows-1-wantRow {
						wantRow = rows - 1 - remaining
					}
					header, _ := m.frame()
					y := min(len(header), height-1) + wantRow
					line := strings.Split(plainFrame(m.view()), "\n")[y]
					if !strings.Contains(line, "> ") || !strings.Contains(line, m.visible[index].name) {
						t.Fatalf("%s at entry %d: expected selection on tree row %d, got %q", key, index, wantRow, line)
					}
					m.handle(keyEvent{name: key})
				}
			}
		})
	}
	t.Run("mouse", func(t *testing.T) {
		m.height = 16
		m.handle(keyEvent{name: "pgdown"})
		start := m.pageStart()
		target := m.visible[start+1].path
		click := mouseOn(t, m, filepath.Base(target), "mouse-left", time.Unix(100, 0))
		m.handle(click)
		if m.focusedPath() != target || m.mouseEntry(click.x, click.y) != m.cursor {
			t.Fatal("the clicked directory moved away from the pointer")
		}
		click.when = click.when.Add(200 * time.Millisecond)
		m.handle(click)
		if !m.branches[target].expanded {
			t.Fatal("double-click did not expand the clicked directory")
		}
		click.name = "mouse-down"
		m.handle(click)
		if m.cursor != start+2 || m.cursor-m.pageStart() != (m.pageRows()-1)/2 {
			t.Fatal("wheel movement did not center the next directory")
		}
	})
}

func TestOutlineRenderingAndHitTesting(t *testing.T) {
	root := fixture(t)
	for i := range 40 {
		if err := os.Mkdir(filepath.Join(root, "Alpha", fmt.Sprintf("child-%02d-\u754c", i)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, noColor := range []string{"", "1"} {
		for _, size := range [][2]int{{100, 24}, {55, 12}, {20, 8}, {4, 2}, {1, 1}} {
			t.Run(fmt.Sprintf("NO_COLOR=%s/%dx%d", noColor, size[0], size[1]), func(t *testing.T) {
				m := focus(t, loadOutline(t, root), "Alpha")
				m = press(m, "right", "")
				m = focus(t, m, "child-39-\u754c")
				m.width, m.height = size[0], size[1]
				m.styles = terminalStyles(true, "xterm", noColor)
				view := m.view()
				lines := strings.Split(plainFrame(view), "\n")
				if len(lines) != m.height || !utf8.ValidString(view) || !strings.Contains(view, m.styles.selected) {
					t.Fatal("tree lost its selected row or terminal bounds")
				}
				for _, line := range lines {
					if textWidth(line) > max(1, m.width-1) {
						t.Fatalf("tree row overflows terminal: %q", line)
					}
				}
				header, _ := m.frame()
				top := len(header)
				if m.height == 1 {
					top = 0
				}
				_, margin := m.viewWidth()
				y := top + m.cursor - m.pageStart()
				if m.mouseEntry(margin, y) != m.cursor {
					t.Fatal("tree mouse hit testing diverged from the scrolled row")
				}
			})
		}
	}
}

func TestOutlineDeepLabelsAndRootControls(t *testing.T) {
	m := browser{
		mode: outlineView, path: "/projects", width: 42, height: 12, cursor: 1,
		styles:  terminalStyles(true, "xterm", ""),
		visible: []entry{{name: "projects", path: "/projects"}, {name: "leaf-\u754c\x1b[2J", path: "/projects/leaf"}},
		outline: []outlineRow{{parent: -1}, {parent: 0, guide: strings.Repeat("\u2502  ", 30) + "\u2514\u2500 "}},
	}
	m.measureOutline()
	width, margin := m.viewWidth()
	line := plainFrame(m.entryLine(1, width).render(width))
	if !strings.Contains(line, "leaf-\u754c") || strings.ContainsRune(line, '\x1b') || textWidth(line) > width {
		t.Fatalf("deep tree clipped the label or allowed terminal controls: %q", line)
	}
	arrow := margin + 2 + textWidth(m.outlineGuide(1, width))
	if !m.outlineDisclosureAt(1, arrow) || m.outlineDisclosureAt(1, arrow+2) {
		t.Fatal("clipped guide and disclosure hit target differ")
	}
	m = browser{mode: outlineView, path: "/", width: 80, height: 24}
	m.rebuildOutline("/")
	m = press(m, "left", "")
	if len(m.visible) != 1 || m.visible[m.cursor].name != "/" || m.path != "/" {
		t.Fatal("Left at an empty filesystem root must stay at the root")
	}
	m = press(m, "right", "")
	if m.cursor != 0 || m.selected != "" {
		t.Fatal("Right on an empty leaf must not move or confirm")
	}
}

func TestOutlineRootChoice(t *testing.T) {
	for _, tt := range []struct{ name, path, home, want string }{
		{"descendant", "/users/alex/src/app", "/users/alex", "/users/alex"},
		{"home itself", "/users/alex", "/users/alex/", "/users/alex"},
		{"outside", "/work/app", "/users/alex", "/"},
		{"shared prefix", "/users/alexander/app", "/users/alex", "/"},
		{"missing home", "/work/app", "", "/"},
		{"relative home", "/work/app", "work", "/"},
		{"filesystem root", "/", "/users/alex", "/"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := treeRootFor(tt.path, tt.home); got != tt.want {
				t.Fatalf("root = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutlineRevealsStartingDirAndKeepsItsIdentity(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	origin := filepath.Join(home, ".work", "project")
	for _, path := range []string{origin, filepath.Join(home, "elsewhere", "child"), filepath.Join(home, ".unrelated")} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	m, err := newBrowser(origin, false)
	if err != nil {
		t.Fatal(err)
	}
	m.restoreViewPreference("")
	if m.treeRoot != home || m.visible[0].path != home || m.focusedPath() != origin || m.hidden {
		t.Fatal("tree must reveal and focus a hidden starting path below home")
	}
	if !visiblePath(m, filepath.Join(home, ".work")) || visiblePath(m, filepath.Join(home, ".unrelated")) {
		t.Fatal("revealing hidden ancestors must not reveal unrelated hidden entries")
	}
	for _, item := range m.visible {
		if item.isDotDir() {
			t.Fatal("tree must contain real parent nodes instead of synthetic navigation rows")
		}
	}
	for _, noColor := range []string{"", "1"} {
		m.styles = terminalStyles(true, "xterm", noColor)
		m.focusPath(origin)
		row := m.entryLine(m.cursor, 100).render(100)
		if !strings.Contains(row, m.styles.origin) || !strings.Contains(plainFrame(row), "project (starting dir)") {
			t.Fatal("starting dir must retain distinct styling and a textual marker")
		}
		m = press(m, "up", "")
		index := slices.IndexFunc(m.visible, func(item entry) bool { return item.path == origin })
		row = m.entryLine(index, 100).render(100)
		if !strings.Contains(row, m.styles.origin) || strings.Contains(row, m.styles.selected) {
			t.Fatal("starting-dir styling must remain distinct from keyboard focus")
		}
		if strings.Contains(row, "33m") || strings.Contains(row, m.styles.strong) {
			t.Fatal("starting dir must use a subdued foreground without a bright or bold accent")
		}
	}
	m.navigate(filepath.Join(home, "elsewhere"), "")
	if m.origin != origin || m.treeRoot != home || !visiblePath(m, origin) {
		t.Fatal("browsing must not move the starting-dir marker")
	}
	for range 2 {
		m = press(m, "v", "v")
	}
	if m.origin != origin || m.treeRoot != home || !visiblePath(m, origin) {
		t.Fatal("mode changes must preserve the original source")
	}
	if !strings.Contains(plainFrame(m.view()), "From:") {
		t.Fatal("tree header must show the starting location independently of focus")
	}
}

func TestOutlineEnterConfirmsEveryDirAndFilteredMatch(t *testing.T) {
	for _, target := range []string{"root", "Alpha", "nested"} {
		for _, filtering := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/filter=%t", target, filtering), func(t *testing.T) {
				root := fixture(t)
				m := loadOutline(t, root)
				name := target
				if target == "root" {
					name = filepath.Base(root)
				}
				if target == "nested" {
					m = focus(t, m, "Alpha")
					m = press(m, "right", "")
				}
				if filtering {
					m = press(m, "/", "/")
					m = press(m, "text", name)
				}
				m = focus(t, m, name)
				want := m.focusedPath()
				if !m.handle(keyEvent{name: "enter"}) || m.selected != want {
					t.Fatal("Enter must confirm the highlighted dir at any depth, even while filtering")
				}
			})
		}
	}
}

func TestOutlineOutsideHomeAndParentFocus(t *testing.T) {
	base := t.TempDir()
	home, outside := filepath.Join(base, "home"), filepath.Join(base, "outside")
	for _, path := range []string{home, outside} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	m, err := newBrowser(outside, false)
	if err != nil {
		t.Fatal(err)
	}
	m.restoreViewPreference("")
	if m.treeRoot != "/" || m.visible[0].path != "/" || m.focusedPath() != outside {
		t.Fatal("outside home, tree must start at / and reveal the starting path")
	}
	m, err = newBrowser(home, false)
	if err != nil {
		t.Fatal(err)
	}
	m = focus(t, m, "..")
	m = press(m, "v", "v")
	if m.treeRoot != "/" || m.focusedPath() != base {
		t.Fatal("entering the tree with home's parent selected must retain that destination")
	}
}
