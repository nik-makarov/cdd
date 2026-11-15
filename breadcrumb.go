package main

import (
	"path/filepath"
	"strings"
)

// Match the same sanitized and clipped text used by pathLine. The ellipsis
// and separators are inert; visible name fragments retain their original path.
func breadcrumbPath(path string, x, width int) string {
	if path == "" || x < 0 || x >= width {
		return ""
	}
	full := display(path)
	shown := tail(full, width)
	if x >= textWidth(shown) {
		return ""
	}
	if textWidth(full) > width && width >= 4 {
		if x < 3 {
			return ""
		}
		x += textWidth(full) - textWidth(shown)
	}
	if x == 0 {
		return string(filepath.Separator)
	}
	position, current := 1, string(filepath.Separator)
	for _, name := range strings.Split(strings.TrimPrefix(path, current), current) {
		current = filepath.Join(current, name)
		end := position + textWidth(display(name))
		if x >= position && x < end {
			return current
		}
		position = end + 1
	}
	return ""
}

func (m browser) breadcrumbAt(x, y int) string {
	if m.height <= 1 {
		return ""
	}
	row := 0
	if m.height >= 8 {
		row = 1
		if m.mode == panesView {
			row = 2
		}
	}
	if y != row {
		return ""
	}
	width, margin := m.viewWidth()
	widths := m.paneWidths(width)
	paths := m.paneLocations(widths)
	left := margin
	for i, cells := range widths {
		if x >= left && x < left+cells {
			return breadcrumbPath(paths[i], x-left, cells)
		}
		left += cells + 3
	}
	return ""
}
