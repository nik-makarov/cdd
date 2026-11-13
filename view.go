package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type viewStyles struct {
	muted, strong, accent, directory, file, link, failure, selected string
}

// Use the terminal's own palette so light and dark themes keep their contrast.
// NO_COLOR suppresses colors, while emphasis and reverse selection remain useful.
func terminalStyles(tty bool, terminal, noColor string) viewStyles {
	if !tty || terminal == "dumb" {
		return viewStyles{}
	}
	s := viewStyles{muted: "\x1b[2m", strong: "\x1b[1m", selected: "\x1b[7m"}
	s.file = s.muted
	s.accent, s.failure = s.strong, s.strong
	if noColor == "" {
		s.accent, s.failure = "\x1b[1;36m", "\x1b[1;31m"
		s.directory = "\x1b[36m"
		s.link = "\x1b[35m"
	}
	return s
}

type span struct{ text, style string }
type viewLine []span

func columnPosition(column int) string { return "\x1b[" + strconv.Itoa(column) + "G" }

// Clip printable text before applying trusted SGR sequences. Never truncate ANSI
// bytes or allow a filename to inject terminal controls into the styled frame.
func (line viewLine) render(width int) string {
	var out strings.Builder
	for _, part := range line {
		printable := display(part.text)
		if printable == "" {
			continue
		}
		text := truncate(printable, width)
		if text == "" {
			break
		}
		out.WriteString(part.style)
		out.WriteString(text)
		if part.style != "" {
			out.WriteString("\x1b[0m")
		}
		width -= textWidth(text)
		// A wide rune that did not fit must not be skipped in favor of a later span.
		if text != printable {
			break
		}
	}
	return out.String()
}

type keyHint struct{ key, action string }

var helpHints = []keyHint{
	{"↑↓ / j k", "move"}, {"→ / l", "open"}, {"← / h / Backspace", "current, then parent"},
	{"Enter", "select"}, {"Home / End", "first / last"}, {"PgUp / PgDn", "page"},
	{"Tab / Shift+Tab", "switch column"},
	{"v", "switch view"},
	{"/", "filter"}, {".", "hidden"}, {"~", "home"}, {"Ctrl+R", "refresh"},
	{"Esc / q / Ctrl+C / Ctrl+D", "cancel"},
}

func (m browser) hints(hints []keyHint) viewLine {
	var line viewLine
	for i, hint := range hints {
		if i > 0 {
			line = append(line, span{"  ", ""})
		}
		line = append(line, span{hint.key, m.styles.strong}, span{" " + hint.action, m.styles.muted})
	}
	return line
}

func (m browser) shortHelp(width int) viewLine {
	if m.filtering {
		return m.hints([]keyHint{{"Enter / Esc", "finish filter"}, {"↑↓", "move"}})
	}
	leftAction := "parent"
	if m.cursor != 0 {
		leftAction = "current"
	}
	hints := []keyHint{{"↑↓", "move"}, {"→", "open"}, {"←", leftAction}, {"Enter", "select"}, {"/", "filter"}, {"v", "view"}, {"?", "help"}, {"Esc", "cancel"}}
	// Preserve selection, help and exit, then opening a directory when space allows.
	for _, key := range []string{"↑↓", "←", "→", "/", "v"} {
		if textWidth(hintText(hints)) <= width {
			break
		}
		for i, hint := range hints {
			if hint.key == key {
				hints = append(hints[:i], hints[i+1:]...)
				break
			}
		}
	}
	return m.hints(hints)
}

func hintText(hints []keyHint) string {
	var parts []string
	for _, hint := range hints {
		parts = append(parts, hint.key+" "+hint.action)
	}
	return strings.Join(parts, "  ")
}

func (m browser) helpLines(width int) []viewLine {
	var lines []viewLine
	var row []keyHint
	for _, hint := range helpHints {
		candidate := append(append([]keyHint(nil), row...), hint)
		if len(row) > 0 && textWidth(hintText(candidate)) > width {
			lines = append(lines, m.hints(row))
			row = nil
		}
		row = append(row, hint)
	}
	if len(row) > 0 {
		lines = append(lines, m.hints(row))
	}
	if !m.compact {
		lines = append(lines, viewLine{{"folder/  ", m.styles.directory}, {"file  ", m.styles.file}, {"directory-link@/", m.styles.link}})
	}
	return append(lines, m.hints([]keyHint{{"? / Esc", "close help"}}))
}

func (m browser) viewWidth() (width, margin int) {
	// Leave the final terminal column untouched to avoid automatic line wrapping.
	width = max(1, m.width-1)
	if width >= 20 {
		margin = 2
	}
	return width - 2*margin, margin
}

func (m browser) pathLine(path string, width int, active bool) viewLine {
	text := tail(display(path), width)
	if !active {
		return viewLine{{text, m.styles.muted}}
	}
	base := display(filepath.Base(path))
	parent := strings.TrimSuffix(text, base)
	if parent != text {
		return viewLine{{parent, ""}, {base, m.styles.accent}}
	}
	return viewLine{{text, m.styles.accent}}
}

func (m browser) frame() (header, footer []viewLine) {
	width, _ := m.viewWidth()
	pathLine := m.pathLine(m.path, width, true)
	header = []viewLine{pathLine}
	if m.height >= 8 {
		header = []viewLine{{{"cdd", m.styles.muted}}, pathLine, nil}
	}
	footer = []viewLine{m.shortHelp(width)}
	if m.showHelp {
		footer = m.helpLines(width)
	}
	if m.filtering || m.query != "" {
		query := "/ " + display(m.query)
		if m.filtering {
			query += "_"
		}
		footer = append([]viewLine{{{query, m.styles.accent}}}, footer...)
	}
	if m.message != "" {
		footer = append([]viewLine{{{"Error: " + display(m.message), m.styles.failure}}}, footer...)
	}
	// Even a tiny terminal must retain one row for the highlighted directory.
	maxFooter := max(0, m.height-len(header)-1)
	if m.height >= 8 {
		maxFooter = max(0, maxFooter-3)
	}
	if len(footer) > maxFooter {
		if m.showHelp && maxFooter > 0 {
			footer = append(footer[:maxFooter-1], footer[len(footer)-1])
		} else {
			footer = footer[:maxFooter]
		}
	}
	return header, footer
}

func (m browser) pageRows() int {
	header, footer := m.frame()
	extra := 0
	if m.height >= 8 {
		extra = 3 // Separator, summary and bottom margin.
	}
	return max(1, m.height-len(header)-len(footer)-extra)
}

func (m browser) columns() int {
	width, _ := m.viewWidth()
	// Leave at least 24 cells per column and three cells for the divider.
	if m.compact && width >= 2*24+3 && len(m.visible) > m.pageRows() {
		return 2
	}
	return 1
}

func (m browser) pageSize() int { return m.pageRows() * m.columns() }

func (m browser) paneWidths(width, columns int) []int {
	if !m.compact {
		return treeWidths(width)
	}
	if columns == 1 {
		return []int{width}
	}
	left := (width - 3) / 2
	return []int{left, width - 3 - left}
}

func (m browser) panePaths(widths []int, margin int) string {
	paths := []string{m.path}
	if m.compact && len(widths) == 2 {
		paths = append(paths, m.path)
	}
	if !m.compact && len(widths) == 3 {
		parent, child := "", ""
		if path := filepath.Dir(m.path); path != m.path {
			parent = path
		}
		if m.cursor >= 0 && m.cursor < len(m.visible) {
			item := m.visible[m.cursor]
			if item.isDir() && item.path != m.path {
				child = item.path
			}
		}
		paths = []string{parent, m.path, child}
	}
	parts := make([]viewLine, len(paths))
	for i, path := range paths {
		parts[i] = m.pathLine(path, widths[i], path == m.path)
	}
	return m.paneRow(parts, widths, margin)
}

func (m browser) entryLine(index, width int) viewLine {
	if index >= len(m.visible) {
		return nil
	}
	return m.itemLine(m.visible[index], width, index == m.cursor)
}

func (m browser) itemLine(item entry, width int, selected bool) viewLine {
	label := item.label()
	style := m.styles.directory
	if !item.isDir() {
		style = m.styles.file
	} else if item.kind == folderLink {
		style = m.styles.link
	}
	row := viewLine{{"  " + label, style}}
	if item.name == "." {
		label = ".  current directory"
		row = viewLine{{"  .", ""}, {"  current directory", m.styles.muted}}
	}
	if selected {
		text := truncate("> "+display(label), width)
		text += strings.Repeat(" ", max(0, width-textWidth(text)))
		row = viewLine{{text, m.styles.selected + style}}
	}
	return row
}

func (m browser) view() string {
	width, margin := m.viewWidth()
	header, footer := m.frame()
	rows, columns := m.pageRows(), m.columns()
	widths := m.paneWidths(width, columns)
	capacity := rows * columns
	start := max(0, m.cursor-rows+1)
	mode := "tree"
	if m.compact {
		mode = "list"
	}
	if m.height >= 8 {
		header[0] = viewLine{{"cdd  " + mode, m.styles.muted}, {"   v", m.styles.strong}, {" view", m.styles.muted}}
	}
	if columns == 2 {
		start = m.cursor / capacity * capacity
		if m.height >= 8 {
			header[0] = append(header[0], span{"   Tab", m.styles.strong}, span{" switch column", m.styles.muted})
		}
	}
	end := min(len(m.visible), start+capacity)
	prefix := strings.Repeat(" ", margin)
	render := func(line viewLine) string { return prefix + line.render(width) }
	var lines []string
	for _, line := range header {
		lines = append(lines, render(line))
	}
	pathRow := 0
	if m.height >= 8 {
		pathRow = 1
		if !m.compact {
			lines[1] = m.treeHeading(width, margin)
			pathRow = 2
		}
	}
	lines[pathRow] = m.panePaths(widths, margin)
	if !m.compact {
		lines = append(lines, m.treeLines(rows, width, margin)...)
	}
	for row := 0; m.compact && row < min(rows, end-start); row++ {
		if columns == 1 {
			lines = append(lines, render(m.entryLine(start+row, width)))
			continue
		}
		parts := []viewLine{m.entryLine(start+row, widths[0]), m.entryLine(start+rows+row, widths[1])}
		lines = append(lines, m.paneRow(parts, widths, margin))
	}
	if m.height >= 8 {
		for len(lines) < m.height-len(footer)-3 {
			lines = append(lines, render(nil))
		}
		hidden := "off"
		if m.hidden {
			hidden = "on"
		}
		directories, files := 0, 0
		for _, item := range m.visible[1:] {
			if item.isDir() {
				directories++
			} else {
				files++
			}
		}
		count := strconv.Itoa(directories) + " directories"
		if directories == 1 {
			count = "1 directory"
		}
		if !m.compact {
			if files == 1 {
				count += " / 1 file"
			} else {
				count += " / " + strconv.Itoa(files) + " files"
			}
		}
		if m.query != "" {
			count = strconv.Itoa(len(m.visible)-1) + " matches"
			if len(m.visible) == 2 {
				count = "1 match"
			}
		}
		if start > 0 {
			count += "  |  more above"
		}
		if end < len(m.visible) {
			count += "  |  more below"
		}
		lines = append(lines, render(nil), render(viewLine{{count + "  |  hidden " + hidden, m.styles.muted}}))
	}
	for _, line := range footer {
		lines = append(lines, render(line))
	}
	if m.height >= 8 {
		lines = append(lines, render(nil))
	}
	// In a one-line terminal the selection is more useful than the path header.
	if m.height <= 1 {
		lines = lines[len(header):]
	}
	return strings.Join(lines[:min(len(lines), max(1, m.height))], "\n")
}

// Reserve two cells for every non-ASCII rune. This conservative bound avoids
// wrapping wide characters without a Unicode-width dependency. Combining marks
// and emoji sequences can leave unused space, but filenames stay valid UTF-8.
func textWidth(s string) int {
	width := 0
	for _, r := range s {
		width++
		if r > 127 {
			width++
		}
	}
	return width
}

func truncate(s string, width int) string {
	cells := 0
	for i, r := range s {
		cost := 1
		if r > 127 {
			cost = 2
		}
		if cells+cost > width {
			return s[:i]
		}
		cells += cost
	}
	return s
}

func tail(s string, width int) string {
	if textWidth(s) <= width {
		return s
	}
	if width < 4 {
		return truncate(s, width)
	}
	runes := []rune(s)
	start, cells := len(runes), 0
	for start > 0 {
		cost := textWidth(string(runes[start-1]))
		if cells+cost > width-3 {
			break
		}
		cells += cost
		start--
	}
	return "..." + string(runes[start:])
}

// Filenames may contain terminal escape sequences. Only render printable text;
// keep the original bytes in the selected path passed back to the shell.
func display(s string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return '�'
		}
		return r
	}, s)
}
