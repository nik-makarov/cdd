package main

import (
	"path/filepath"
	"strings"
)

func (m browser) mutationView() string {
	width, margin := m.viewWidth()
	edit := m.mutation
	title, label, action := "Rename", "New name: ", "rename"
	if edit.kind == "trash" {
		title, label, action = "Move to Trash", "Confirm name: ", "move to Trash"
	}
	rows := []viewLine{
		{{"Name: " + tail(display(filepath.Base(edit.path)), max(1, width-6)), m.styles.strong}},
		{{label, m.styles.strong}, {tail(display(edit.input)+"_", max(1, width-textWidth(label))), m.styles.accent}},
	}
	if edit.busy {
		rows[1] = viewLine{{"Moving to Trash...", m.styles.accent}}
		rows = append(rows, m.hints([]keyHint{{"Ctrl+C", "exit"}}))
	} else {
		rows = append(rows, m.hints([]keyHint{{"Enter", action}, {"Esc", "cancel"}}))
	}
	if m.message != "" {
		rows = append(rows, viewLine{{"Error: " + display(m.message), m.styles.failure}})
	}
	if m.height >= len(rows)+3 {
		rows = append([]viewLine{{{title, m.styles.strong}}, {{tail(display(edit.path), width), m.styles.muted}}, nil}, rows...)
	}
	rows = rows[:min(len(rows), max(1, m.height))]
	lines := make([]string, max(0, (m.height-len(rows))/2))
	for _, row := range rows {
		lines = append(lines, strings.Repeat(" ", margin)+row.render(width))
	}
	for len(lines) < max(1, m.height) {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
