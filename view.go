package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

type viewStyles struct {
	muted, strong, accent, directory, failure, selected string
}

// Use the terminal's own palette so light and dark themes keep their contrast.
// NO_COLOR suppresses colors, while emphasis and reverse selection remain useful.
func terminalStyles(tty bool, terminal, noColor string) viewStyles {
	if !tty || terminal == "dumb" {
		return viewStyles{}
	}
	s := viewStyles{muted: "\x1b[2m", strong: "\x1b[1m", selected: "\x1b[7m"}
	s.accent, s.failure = s.strong, s.strong
	if noColor == "" {
		s.accent, s.failure = "\x1b[1;36m", "\x1b[1;31m"
		s.directory = "\x1b[36m"
	}
	return s
}

type span struct{ text, style string }
type viewLine []span

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
	{"↑↓ / j k", "move"}, {"→ / l", "open"}, {"← / h / Backspace", "parent"},
	{"Enter", "select"}, {"Home / End", "first / last"}, {"PgUp / PgDn", "page"},
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
	hints := []keyHint{{"↑↓", "move"}, {"→", "open"}, {"←", "parent"}, {"Enter", "select"}, {"/", "filter"}, {"?", "help"}, {"Esc", "cancel"}}
	// Preserve selection, help and exit, then opening a directory when space allows.
	for _, key := range []string{"↑↓", "←", "→", "/"} {
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

func (m browser) frame() (header, footer []viewLine) {
	width, _ := m.viewWidth()
	path := tail(display(m.path), width)
	base := filepath.Base(m.path)
	parent := strings.TrimSuffix(path, display(base))
	pathLine := viewLine{{path, m.styles.accent}}
	if parent != path {
		pathLine = viewLine{{parent, m.styles.muted}, {display(base), m.styles.accent}}
	}
	header = []viewLine{pathLine}
	if m.height >= 8 {
		header = []viewLine{{{"0go", m.styles.muted}}, pathLine, nil}
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

func (m browser) pageSize() int {
	header, footer := m.frame()
	extra := 0
	if m.height >= 8 {
		extra = 3 // Separator, summary and bottom margin.
	}
	return max(1, m.height-len(header)-len(footer)-extra)
}

func (m browser) view() string {
	width, margin := m.viewWidth()
	header, footer := m.frame()
	start := max(0, m.cursor-m.pageSize()+1)
	end := min(len(m.visible), start+m.pageSize())
	lines := append([]viewLine(nil), header...)
	for i := start; i < end; i++ {
		label := m.visible[i].name + "/"
		row := viewLine{{"  " + m.visible[i].name, ""}, {"/", m.styles.directory}}
		if i == 0 {
			label = ".  current directory"
			row = viewLine{{"  .", ""}, {"  current directory", m.styles.muted}}
		}
		if i == m.cursor {
			text := truncate("> "+display(label), width)
			text += strings.Repeat(" ", max(0, width-textWidth(text)))
			row = viewLine{{text, m.styles.selected}}
		}
		lines = append(lines, row)
	}
	if m.height >= 8 {
		for len(lines) < m.height-len(footer)-3 {
			lines = append(lines, nil)
		}
		hidden := "off"
		if m.hidden {
			hidden = "on"
		}
		count := fmt.Sprintf("%d directories", len(m.visible)-1)
		if len(m.visible) == 2 {
			count = "1 directory"
		}
		if m.query != "" {
			count = fmt.Sprintf("%d matches", len(m.visible)-1)
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
		lines = append(lines, nil, viewLine{{count + "  |  hidden " + hidden, m.styles.muted}})
	}
	lines = append(lines, footer...)
	if m.height >= 8 {
		lines = append(lines, nil)
	}
	// In a one-line terminal the selection is more useful than the path header.
	if m.height <= 1 {
		lines = lines[len(header):]
	}
	var rendered []string
	for _, line := range lines[:min(len(lines), max(1, m.height))] {
		rendered = append(rendered, strings.Repeat(" ", margin)+line.render(width))
	}
	return strings.Join(rendered, "\n")
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
