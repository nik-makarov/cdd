package main

import (
	"slices"
	"strings"
)

func (m browser) entryButtons() []viewLine {
	width, _ := m.viewWidth()
	if m.creating || m.filtering || m.mutation != nil || !m.mutationAllowed() {
		return nil
	}
	label := "[ Rename  r ]"
	if width < textWidth(label) {
		return nil
	}
	return []viewLine{{{strings.Repeat(" ", (width-textWidth(label))/2), ""}, {label, m.styles.directory}}}
}

func (m browser) entryButtonAt(x, y int) string {
	buttons := m.entryButtons()
	_, footer := m.frame()
	_, margin := m.viewWidth()
	for i, line := range footer {
		if y != m.footerTop(footer)+i || !slices.ContainsFunc(buttons, func(row viewLine) bool { return slices.Equal(line, row) }) {
			continue
		}
		left := margin + textWidth(line[0].text)
		if x >= left && x < left+textWidth(line[1].text) {
			return "r"
		}
	}
	return ""
}
