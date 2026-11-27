package main

import (
	"strconv"
	"strings"
	"time"
)

type mouseClick struct {
	path string
	x, y int
	when time.Time
}

const doubleClickInterval = 400 * time.Millisecond

// SGR reports use one-based terminal cells: CSI < button ; x ; y M (or m).
func decodeMouse(sequence string) keyEvent {
	if len(sequence) < 2 || sequence[0] != '<' {
		return keyEvent{}
	}
	end := sequence[len(sequence)-1]
	if end != 'M' && end != 'm' {
		return keyEvent{}
	}
	fields := strings.Split(sequence[1:len(sequence)-1], ";")
	if len(fields) != 3 {
		return keyEvent{}
	}
	var values [3]int
	for i, field := range fields {
		for _, c := range field {
			if c < '0' || c > '9' {
				return keyEvent{}
			}
		}
		value, err := strconv.Atoi(field)
		if err != nil {
			return keyEvent{}
		}
		values[i] = value
	}
	if values[1] < 1 || values[2] < 1 {
		return keyEvent{}
	}
	return mouseKey(values[0], values[1]-1, values[2]-1, end == 'm')
}

func mouseKey(button, x, y int, released bool) keyEvent {
	key := keyEvent{name: "mouse-ignore", x: x, y: y}
	if released || button == 3 {
		key.name = "mouse-release"
		return key
	}
	switch button {
	case 0:
		key.name = "mouse-left"
	case 2:
		key.name = "mouse-right"
	case 64:
		key.name = "mouse-up"
	case 65:
		key.name = "mouse-down"
	}
	return key
}

// Mouse hit testing uses the same viewport as the rendered lists.
func (m browser) mouseRow(y int) int {
	if y < 0 || y >= max(1, m.height) {
		return -1
	}
	header, _ := m.frame()
	top := len(header)
	if m.height <= 1 {
		top = 0
	}
	row, rows := y-top, m.pageRows()
	if row < 0 || row >= rows {
		return -1
	}
	return row
}

func (m browser) mouseEntry(x, y int) int {
	row := m.mouseRow(y)
	if row < 0 {
		return -1
	}
	width, margin := m.viewWidth()
	widths := m.paneWidths(width)
	if m.mode == outlineView {
		widths[0], _ = m.outlineWidths(width)
	}
	left := margin
	for column, width := range widths {
		if x >= left && x < left+width {
			if m.mode == panesView && len(widths) == 3 && column != 1 {
				return -1
			}
			index := m.pageStart() + row
			if index >= len(m.visible) {
				return -1
			}
			return index
		}
		left += width + 3
	}
	return -1
}

func (m browser) mouseRecentEntry(x, y int) int {
	if m.mode != outlineView {
		return -1
	}
	width, margin := m.viewWidth()
	_, right := m.outlineWidths(width)
	if x < margin+width-right || x >= margin+width {
		return -1
	}
	row := m.mouseRow(y)
	if row < 0 {
		return -1
	}
	list, start := m.recentPage(right, m.pageRows())
	index := start + row
	if index >= len(list.rows) {
		return -1
	}
	return list.rows[index].entry
}

func (c mouseClick) doubleClick(path string, key keyEvent) bool {
	elapsed, drift := key.when.Sub(c.when), key.x-c.x
	return !c.when.IsZero() && c.path == path && c.y == key.y && drift >= -2 && drift <= 2 && elapsed >= 0 && elapsed <= doubleClickInterval
}

// expireMouseClick releases the held row and reports whether the list needs a redraw.
func (m *browser) expireMouseClick(now time.Time) bool {
	if m.lastClick.when.IsZero() || now.Sub(m.lastClick.when) <= doubleClickInterval {
		return false
	}
	m.lastClick = mouseClick{}
	if m.recent.focused {
		width, _ := m.viewWidth()
		_, start := m.recentPage(width, m.pageRows())
		changed := start != m.recent.scroll
		m.recent.scroll = start
		return changed
	}
	return m.mode == outlineView && m.pageStart() != m.scroll
}

func (m *browser) handleMouse(key keyEvent) bool {
	if m.creating {
		m.lastClick = mouseClick{}
		return false
	}
	if key.name == "mouse-release" {
		return false
	}
	// Resolve the displayed row before clearing the click that holds it in place.
	index := m.mouseEntry(key.x, key.y)
	recentIndex := m.mouseRecentEntry(key.x, key.y)
	previous := m.lastClick
	m.lastClick = mouseClick{}
	if key.name == "mouse-left" {
		if action := m.entryButtonAt(key.x, key.y); action != "" {
			return m.handle(keyEvent{name: action})
		}
	}
	if key.name == "mouse-left" && m.rootButtonAt(key.x, key.y) {
		m.showRootTree()
		return false
	}
	if key.name == "mouse-left" {
		if path := m.breadcrumbAt(key.x, key.y); path != "" {
			m.recent.focused = false
			m.navigate(path, "")
			return false
		}
	}
	if index := recentIndex; index >= 0 {
		switch key.name {
		case "mouse-left":
			path := m.recent.paths[index]
			reveal := m.recent.focused && previous.doubleClick(path, key)
			width, _ := m.viewWidth()
			if !m.recent.focused {
				row := m.mouseRow(key.y)
				m.recent.focused = true
				m.recent.scroll = max(0, m.recentLayout(width).first[index]-row)
			} else if previous.path == "" {
				_, m.recent.scroll = m.recentPage(width, m.pageRows())
			}
			m.recent.cursor = index
			m.filtering, m.showHelp, m.message = false, false, ""
			m.lastClick = mouseClick{path: path, x: key.x, y: key.y, when: key.when}
			if reveal {
				m.revealRecent()
			}
		case "mouse-up", "mouse-down":
			m.recent.focused, m.filtering, m.showHelp = true, false, false
			return m.handleRecent(key)
		}
		return false
	}
	if index < 0 {
		return false
	}
	switch key.name {
	case "mouse-up":
		m.move("up")
	case "mouse-down":
		m.move("down")
	case "mouse-left":
		m.cursor, m.message = index, ""
		if m.mode == outlineView && m.outlineDisclosureAt(index, key.x) {
			m.toggleOutline()
			return false
		}
		m.lastClick = mouseClick{path: m.visible[index].path, x: key.x, y: key.y, when: key.when}
		if previous.doubleClick(m.lastClick.path, key) {
			m.lastClick = mouseClick{}
			if m.visible[index].name == "." {
				m.filtering = false
				return m.handle(keyEvent{name: "confirm"})
			}
			if m.mode == outlineView && m.visible[index].isDir() && m.visible[index].name != ".." {
				m.toggleOutline()
				return false
			}
			m.openSelected()
		}
	}
	return false
}
