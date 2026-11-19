package main

import (
	"path/filepath"
	"strings"
)

// Keep the anchor separate so clipping cannot turn an address into a basename.
type pathAddress struct {
	base, prefix, rest string
}

func addressFor(path, home string) pathAddress {
	if path == "" {
		return pathAddress{}
	}
	if filepath.IsAbs(home) {
		rel, err := filepath.Rel(home, path)
		if err == nil && filepath.IsLocal(rel) {
			if rel == "." {
				return pathAddress{base: home, prefix: "~"}
			}
			return pathAddress{home, "~/", rel}
		}
	}
	return pathAddress{"/", "/", strings.TrimPrefix(path, "/")}
}

func (a pathAddress) label(width int) string {
	return truncate(a.prefix, width) + tail(display(a.rest), max(0, width-len(a.prefix)))
}

// The ellipsis and separators are inert. Visible components retain their
// literal targets even when their displayed names contain replacement runes.
func (a pathAddress) breadcrumb(x, width int) string {
	if a.base == "" || x < 0 || x >= width {
		return ""
	}
	if x == 0 {
		return a.base
	}
	x -= len(a.prefix)
	width -= len(a.prefix)
	full := display(a.rest)
	shown := tail(full, width)
	if x < 0 || x >= textWidth(shown) {
		return ""
	}
	if textWidth(full) > width && width >= 4 {
		if x < 3 {
			return ""
		}
		x += textWidth(full) - textWidth(shown)
	}
	position, current := 0, a.base
	for name := range strings.SplitSeq(a.rest, "/") {
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
			return addressFor(paths[i], m.home).breadcrumb(x-left, cells)
		}
		left += cells + 3
	}
	return ""
}
