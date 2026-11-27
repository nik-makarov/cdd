package main

import "unicode/utf8"

const fromPrefix = "From: "

// Show recent places full width on focus and as a sidebar when the tree fits.
func (m browser) outlineWidths(width int) (tree, recent int) {
	if m.recent.focused {
		return 0, width
	}
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

type recentRow struct {
	entry int
	text  string
}

type recentLayout struct {
	rows  []recentRow
	first []int // First row of each entry, followed by the end of the list.
}

func (m browser) recentLayout(width int) recentLayout {
	var list recentLayout
	gutter := "  "
	if width < 4 {
		gutter = ""
	}
	cells := width - len(gutter)
	for i, path := range m.recent.paths {
		list.first = append(list.first, len(list.rows))
		marker := gutter
		if gutter != "" && m.recent.focused && i == m.recent.cursor {
			marker = "> "
		}
		address := addressFor(path, m.home)
		if !m.recent.focused {
			list.rows = append(list.rows, recentRow{i, marker + address.label(cells)})
			continue
		}
		rest := address.prefix + display(address.rest)
		for {
			part := truncate(rest, cells)
			if part == "" && rest != "" {
				// A one-cell terminal cannot display a wide rune.
				_, n := utf8.DecodeRuneInString(rest)
				rest, part = rest[n:], "?"
			} else {
				rest = rest[len(part):]
			}
			list.rows = append(list.rows, recentRow{i, marker + part})
			if rest == "" {
				break
			}
			marker = gutter
		}
	}
	list.first = append(list.first, len(list.rows))
	return list
}

func (list recentLayout) pageStart(r recentPlaces, rows int) int {
	if len(list.rows) == 0 {
		return 0
	}
	first, end := list.first[r.cursor], list.first[r.cursor+1]
	if end-first > rows {
		return max(first, min(r.scroll, end-rows))
	}
	return max(0, end-rows, min(r.scroll, first, len(list.rows)-rows))
}

func (m browser) recentPage(width, rows int) (recentLayout, int) {
	if width == 0 {
		return recentLayout{}, 0
	}
	list := m.recentLayout(width)
	if m.recent.focused && m.lastClick.path != "" {
		// Keep the clicked row under the pointer while waiting for a double-click.
		return list, m.recent.scroll
	}
	return list, list.pageStart(m.recent, rows)
}

func (m browser) outlineLines(rows, width, margin int) []string {
	left, right := m.outlineWidths(width)
	start := m.pageStart()
	list, recentStart := m.recentPage(right, rows)
	lines := make([]string, 0, rows)
	for row := range rows {
		var tree, recent viewLine
		if left > 0 {
			tree = m.entryLine(start+row, left)
		}
		if index := recentStart + row; index < len(list.rows) {
			line := list.rows[index]
			style := m.styles.directory
			if m.recent.focused && line.entry == m.recent.cursor {
				style = m.styles.selected
			}
			recent = viewLine{{line.text, style}}
		} else if right > 0 && row == 0 {
			recent = viewLine{{"No recent places", m.styles.muted}}
		}
		lines = append(lines, m.outlinePanelRow(tree, recent, left, right, margin))
	}
	return lines
}
