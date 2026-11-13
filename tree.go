package main

import "strings"

func treeWidths(width int) []int {
	if width < 54 {
		return []int{width}
	}
	side := (width - 6) / 4
	return []int{side, width - 6 - 2*side, side}
}

// All positioning is generated here; filesystem text is clipped and sanitized
// separately, before any trusted terminal controls are written.
func (m browser) paneRow(parts []viewLine, widths []int, margin int) string {
	var out strings.Builder
	out.WriteString(strings.Repeat(" ", margin))
	column := margin
	for i, part := range parts {
		if i > 0 {
			out.WriteString(columnPosition(column + 1))
			out.WriteString((viewLine{{" | ", m.styles.muted}}).render(3))
			column += 3
		}
		out.WriteString(part.render(widths[i]))
		column += widths[i]
	}
	return out.String()
}

func (m browser) treeHeading(width, margin int) string {
	widths := treeWidths(width)
	parts := []viewLine{{{"[Current]", m.styles.accent}}}
	if len(widths) == 3 {
		parts = []viewLine{{{"Parent", m.styles.muted}}, {{"[Current]", m.styles.accent}}, {{"Children", m.styles.muted}}}
	}
	return m.paneRow(parts, widths, margin)
}

func (m browser) contextEntries(items []entry) []entry {
	if m.hidden {
		return items
	}
	visible := make([]entry, 0, len(items))
	for _, item := range items {
		if !strings.HasPrefix(item.name, ".") {
			visible = append(visible, item)
		}
	}
	return visible
}

func (m browser) contextLine(items []entry, index, width int, message string) viewLine {
	if len(items) == 0 {
		if index > 0 {
			return nil
		}
		if message == "" {
			message = "(empty)"
		}
		return viewLine{{message, m.styles.muted}}
	}
	if index >= len(items) {
		return nil
	}
	return m.itemLine(items[index], width, false, items[index].path == m.path)
}

func (m browser) treeLines(rows, width, margin int) []string {
	widths := treeWidths(width)
	start := max(0, m.cursor-rows+1)
	parent, children := m.contextEntries(m.parent), m.contextEntries(m.children)
	parentStart := 0
	for i, item := range parent {
		if item.path == m.path {
			parentStart = max(0, i-rows+1)
			break
		}
	}
	lines := make([]string, 0, rows)
	for row := range rows {
		parts := []viewLine{m.entryLine(start+row, widths[0])}
		if len(widths) == 3 {
			var child viewLine
			if m.visible[m.cursor].path != m.path {
				child = m.contextLine(children, row, widths[2], m.childErr)
			}
			parts = []viewLine{
				m.contextLine(parent, parentStart+row, widths[0], m.parentErr),
				m.entryLine(start+row, widths[1]),
				child,
			}
		}
		lines = append(lines, m.paneRow(parts, widths, margin))
	}
	return lines
}
