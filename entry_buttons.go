package main

import (
	"runtime"
	"slices"
	"strings"
)

func entryActions() []keyHint {
	return []keyHint{{"r", "[ Rename  r ]"}, {"trash", "[ Delete  " + deleteShortcut(runtime.GOOS) + " ]"}}
}

func (m browser) entryButtons() []viewLine {
	width, _ := m.viewWidth()
	if m.creating || m.filtering || m.mutation != nil || !m.mutationAllowed() {
		return nil
	}
	actions := entryActions()
	var rows []viewLine
	for _, action := range actions {
		label, style := action.action, m.styles.directory
		if action.key == "trash" {
			style = m.styles.failure
			if textWidth(label) > width {
				label = "[ Delete ]"
			}
		}
		if textWidth(label) <= width {
			rows = append(rows, viewLine{{label, style}})
		}
	}
	if len(rows) == 2 && textWidth(rows[0][0].text)+textWidth(rows[1][0].text)+2 <= width {
		rows = []viewLine{append(append(rows[0], span{"  ", ""}), rows[1]...)}
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
