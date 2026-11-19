package main

const fromPrefix = "From: "

// Give the tree its full width before allocating space for recent places.
func (m browser) outlineWidths(width int) (tree, recent int) {
	const minRecent, divider = 30, 3
	required := m.outlineWidth
	if m.height >= 8 {
		from := addressFor(m.origin, m.home)
		required = max(required, len(fromPrefix)+len(from.prefix)+textWidth(display(from.rest)))
	}
	available := width - divider
	if available-required >= minRecent {
		recent = min(available-required, max(minRecent, available*2/5))
		return available - recent, recent
	}
	if m.recent.focused {
		return 0, width
	}
	return width, 0
}

func (m browser) outlinePanelRow(tree, recent viewLine, left, right, margin int) string {
	if right == 0 {
		return m.paneRow([]viewLine{tree}, []int{left}, margin)
	}
	if left == 0 {
		return m.paneRow([]viewLine{recent}, []int{right}, margin)
	}
	return m.paneRow([]viewLine{tree, recent}, []int{left, right}, margin)
}

func (m browser) recentLine(index, width int) viewLine {
	r := m.recent
	if len(r.paths) == 0 && index == 0 {
		return viewLine{{"No recent places", m.styles.muted}}
	}
	if index >= len(r.paths) {
		return nil
	}
	marker, style := "  ", m.styles.directory
	if r.focused && index == r.cursor {
		marker, style = "> ", m.styles.selected
	}
	if width <= 2 {
		marker = ""
	}
	return viewLine{{marker + addressFor(r.paths[index], m.home).label(width-len(marker)), style}}
}

func (m browser) outlineLines(rows, width, margin int) []string {
	left, right := m.outlineWidths(width)
	start, recentStart := m.pageStart(), m.recent.pageStart(rows)
	lines := make([]string, 0, rows)
	for row := range rows {
		var tree, recent viewLine
		if left > 0 {
			tree = m.entryLine(start+row, left)
		}
		if right > 0 {
			recent = m.recentLine(recentStart+row, right)
		}
		lines = append(lines, m.outlinePanelRow(tree, recent, left, right, margin))
	}
	return lines
}
