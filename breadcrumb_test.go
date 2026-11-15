package main

import (
	"path/filepath"
	"testing"
)

func TestBreadcrumbClippingAndLiteralPaths(t *testing.T) {
	for _, tt := range []struct {
		path     string
		x, width int
		want     string
	}{
		{"/one/two", 0, 30, "/"},
		{"/one/two", 2, 30, "/one"},
		{"/one/two", 4, 30, ""},
		{"/one/two", 6, 30, "/one/two"},
		{"/one/two", 9, 30, ""},
		{"/one/two", 2, 6, ""},
		{"/one/two", 3, 6, "/one/two"},
		{"/a/\u754c/$(literal)", 4, 40, "/a/\u754c"},
		{"/a/\x1bname", 4, 40, "/a/\x1bname"},
		{"/one", 0, 1, "/"},
	} {
		if got := breadcrumbPath(tt.path, tt.x, tt.width); got != tt.want {
			t.Fatalf("%q at %d/%d: got %q, want %q", tt.path, tt.x, tt.width, got, tt.want)
		}
	}
}

func TestBreadcrumbNavigatesWithoutConfirming(t *testing.T) {
	root := fixture(t)
	m := loadOutline(t, root)
	m = focus(t, m, "Alpha")
	m = press(m, "right", "")
	m = press(m, "right", "")
	width, margin := m.viewWidth()
	path := m.focusedPath()
	if m.paneLocations([]int{width})[0] != path {
		t.Fatal("tree breadcrumb must follow the highlighted dir")
	}
	for x := margin; x < margin+width; x++ {
		if m.breadcrumbAt(x, 1) != filepath.Dir(path) {
			continue
		}
		if m.handle(keyEvent{name: "mouse-left", x: x, y: 1}) || m.selected != "" || m.path != filepath.Dir(path) || m.origin != root {
			t.Fatal("breadcrumb clicks must browse without changing the shell destination")
		}
		return
	}
	t.Fatal("parent component had no clickable cells")
}
