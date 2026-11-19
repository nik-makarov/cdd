package main

import (
	"runtime"
	"slices"
	"strings"
)

func (m browser) entryActions(platform string) []keyHint {
	var actions []keyHint
	if len(m.visible) > 0 {
		actions = append(actions, keyHint{"y", "[ Copy Path  y ]"})
		if platform == "darwin" {
			actions = append(actions, keyHint{"copy-entry", "[ Copy ]"}, keyHint{"paste-entry", "[ Paste ]"})
		}
	}
	if platform == "darwin" && len(m.visible) > 0 && m.visible[m.cursor].isDir() {
		actions = append(actions, keyHint{"f", "[ Finder  f ]"})
	}
	if m.mutationAllowed() {
		actions = append(actions, keyHint{"r", "[ Rename  r ]"}, keyHint{"trash", "[ Delete ]"})
	}
	return actions
}

func (m browser) entryButtons() []viewLine {
	width, _ := m.viewWidth()
	if m.recent.focused || m.creating || m.filtering || m.mutation != nil || m.filePreview != nil || m.paste != nil {
		return nil
	}
	actions := m.entryActions(runtime.GOOS)
	var rows []viewLine
	var row viewLine
	used := 0
	for _, action := range actions {
		label := action.action
		size := textWidth(label)
		if size > width {
			continue
		}
		if len(row) > 0 && used+2+size > width {
			rows = append(rows, row)
			row, used = nil, 0
		}
		if len(row) > 0 {
			row = append(row, span{"  ", ""})
			used += 2
		}
		row = append(row, span{label, m.styles.directory})
		used += size
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	for i, row := range rows {
		used := 0
		for _, part := range row {
			used += textWidth(part.text)
		}
		rows[i] = append(viewLine{{strings.Repeat(" ", (width-used)/2), ""}}, row...)
	}
	return rows
}

func (m browser) entryButtonAt(x, y int) string {
	buttons := m.entryButtons()
	_, footer := m.frame()
	_, margin := m.viewWidth()
	for i, line := range footer {
		if y != m.footerTop(footer)+i || !slices.ContainsFunc(buttons, func(row viewLine) bool { return slices.Equal(line, row) }) {
			continue
		}
		left := margin
		for _, part := range line {
			if x >= left && x < left+textWidth(part.text) {
				switch part.text {
				case "[ Copy ]":
					return "copy-entry"
				case "[ Paste ]":
					return "paste-entry"
				}
				if strings.HasPrefix(part.text, "[ Copy Path") {
					return "y"
				}
				if strings.HasPrefix(part.text, "[ Finder") {
					return "f"
				}
				if strings.HasPrefix(part.text, "[ Rename") {
					return "r"
				}
				if strings.HasPrefix(part.text, "[ Delete") {
					return "trash"
				}
			}
			left += textWidth(part.text)
		}
	}
	return ""
}
