package main

import (
	"strings"
	"testing"
)

func TestTreeViewKeepsOnlyMiddleColumnActive(t *testing.T) {
	m := browser{path: "/projects/current", width: 100, height: 18, styles: terminalStyles(true, "xterm", "")}
	m.visible = []entry{{name: ".", path: m.path}, {name: "selected", path: m.path + "/selected"}}
	m.cursor = 1
	m.parent = []entry{{name: "current", path: m.path}, {name: "sibling", path: "/projects/sibling"}}
	m.children = []entry{{name: "nested"}, {name: "readme.txt", kind: file}}
	view := m.view()
	plain := plainFrame(view)
	for _, text := range []string{"Parent", "[Current]", "Children", "sibling/", "nested/", "readme.txt", "> selected/"} {
		if !strings.Contains(plain, text) {
			t.Errorf("tree missing %q: %q", text, plain)
		}
	}
	if strings.Count(view, "\x1b[7m") != 1 || strings.Count(plain, ">") != 1 {
		t.Fatal("only the active middle row should have a selection bar")
	}
	width, margin := m.viewWidth()
	widths := treeWidths(width)
	for _, line := range strings.Split(plain, "\n") {
		if index := strings.Index(line, ">"); index >= 0 && index != margin+widths[0]+3 {
			t.Fatalf("selection moved outside the middle column: %q", line)
		}
	}
	m.width = 40
	plain = plainFrame(m.view())
	if !strings.Contains(plain, "> selected/") || strings.Contains(plain, "sibling/") || strings.Contains(plain, "nested/") || m.compact {
		t.Fatal("narrow tree view must retain the active list without changing view mode")
	}
}

func TestCurrentDirectoryLeavesChildrenPaneBlank(t *testing.T) {
	m := browser{path: "/projects/current", width: 100, height: 18}
	m.visible = []entry{{name: ".", path: m.path}, {name: "nested", path: m.path + "/nested"}}
	width, margin := m.viewWidth()
	widths := treeWidths(width)
	childStart := margin + widths[0] + widths[1] + 6
	for _, line := range m.treeLines(m.pageRows(), width, margin) {
		plain := plainFrame(line)
		if len(plain) >= childStart && strings.TrimSpace(plain[childStart:]) != "" {
			t.Fatalf("current-directory row left text in Children: %q", plain)
		}
	}
	m.cursor = 1
	first := plainFrame(m.treeLines(m.pageRows(), width, margin)[0])
	if len(first) <= childStart || strings.TrimSpace(first[childStart:]) != "(empty)" {
		t.Fatal("an empty child directory should still display its empty-preview label")
	}
}
