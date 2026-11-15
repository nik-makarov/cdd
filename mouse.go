package main

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

type mouseClick struct {
	path string
	x, y int
	when time.Time
}

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

// mouseEntry maps screen cells to the same rows and columns used by view.
// Side previews, dividers, margins, headers, footers, and empty rows are inert.
func (m browser) mouseEntry(x, y int) int {
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
	width, margin := m.viewWidth()
	widths := m.paneWidths(width, m.columns())
	left := margin
	for column, width := range widths {
		if x >= left && x < left+width {
			if !m.compact && len(widths) == 3 && column != 1 {
				return -1
			}
			index := m.pageStart() + row
			if m.compact {
				index += column * rows
			}
			if index >= len(m.visible) {
				return -1
			}
			return index
		}
		left += width + 3
	}
	return -1
}

func (m browser) selectButtonAt(x, y int) bool {
	button := m.selectButton()
	_, margin := m.viewWidth()
	if len(button) == 0 || y < 0 || y >= m.height {
		return false
	}
	left := margin + textWidth(button[0][0].text)
	if x < left || x >= left+textWidth(button[0][1].text) {
		return false
	}
	_, footer := m.frame()
	top := m.footerTop(footer)
	for i, line := range footer {
		if y == top+i {
			return slices.ContainsFunc(button, func(row viewLine) bool { return slices.Equal(line, row) })
		}
	}
	return false
}

func (m *browser) handleMouse(key keyEvent) bool {
	if m.creating {
		m.lastClick = mouseClick{}
		return false
	}
	if key.name == "mouse-release" {
		return false
	}
	previous := m.lastClick
	m.lastClick = mouseClick{}
	if key.name == "mouse-left" && m.selectButtonAt(key.x, key.y) {
		return m.handle(keyEvent{name: "confirm"})
	}
	index := m.mouseEntry(key.x, key.y)
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
		m.lastClick = mouseClick{path: m.visible[index].path, x: key.x, y: key.y, when: key.when}
		elapsed := key.when.Sub(previous.when)
		drift := key.x - previous.x
		if !previous.when.IsZero() && previous.path == m.lastClick.path && drift >= -2 && drift <= 2 && previous.y == key.y && elapsed >= 0 && elapsed <= 400*time.Millisecond {
			m.lastClick = mouseClick{}
			if m.visible[index].name == "." {
				m.filtering = false
				return m.handle(keyEvent{name: "confirm"})
			}
			m.openSelected()
		}
	}
	return false
}
