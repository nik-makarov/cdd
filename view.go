package main

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

type viewStyles struct {
	muted, strong, accent, directory, file, link, failure, selected, origin string
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
	s.origin = s.muted
	if noColor == "" {
		s.accent, s.failure = "\x1b[1;36m", "\x1b[1;31m"
		s.directory = "\x1b[36m"
		s.link = "\x1b[35m"
	}
	return s
}

type span struct{ text, style string }
type viewLine []span

func (line viewLine) cells() int {
	width := 0
	for _, part := range line {
		width += textWidth(display(part.text))
	}
	return width
}

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
	{"↑↓ / j k", "move"}, {"→ / l", "open"}, {"← / h / Backspace", "focus parent, then open"},
	{"Enter", "open file/dir; . changes dir"}, {changeShortcut(runtime.GOOS), "change dir"},
	{"Home / End", "first / last"}, {"PgUp / PgDn", "page"},
	{"n", "new dir"},
	{"Shift+N", "new file"},
	{"r", "rename file/dir"},
	{"y", "copy path"},
	{deleteShortcut(runtime.GOOS), "move to Trash"},
	{"Space", "preview file"},
	{"v", "switch view"},
	{"/", "filter"}, {".", "hidden"}, {"~", "home"}, {"Ctrl+R", "refresh"},
	{"Esc / q / Ctrl+C / Ctrl+D", "cancel"},
	{"Click", "highlight"}, {"Double click", "open"}, {"Wheel", "move"},
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
	if m.recent.focused {
		if len(m.recent.paths) == 0 {
			return m.hints([]keyHint{{"Tab", "tree"}, {"Esc", "cancel"}})
		}
		return m.hints([]keyHint{{"Enter", "show in tree"}, {"Tab", "tree"}, {"↑↓", "move"}, {"Esc", "cancel"}})
	}
	if m.mutation != nil {
		return m.hints([]keyHint{{"Enter", "rename"}, {"Esc", "cancel"}})
	}
	if m.creating {
		return m.hints([]keyHint{{"Enter", "create"}, {"Esc", "cancel"}})
	}
	if m.filtering {
		if m.mode == outlineView {
			hints := []keyHint{{"Esc", "finish filter"}, {"↑↓", "move"}}
			if len(m.visible) > 0 && m.visible[m.cursor].isDir() {
				hints = append([]keyHint{{"Enter", "change dir"}}, hints...)
			} else {
				hints = append([]keyHint{{"Enter", "open file"}}, hints...)
			}
			return m.hints(hints)
		}
		return m.hints([]keyHint{{"Enter / Esc", "finish filter"}, {"↑↓", "move"}})
	}
	leftAction, rightAction, enterAction := "parent", "open", "open"
	if m.mode == outlineView {
		rightAction, enterAction = "expand", "change dir"
		if len(m.visible) > 0 {
			branch := m.branches[m.focusedPath()]
			if branch.expanded && len(m.outlineEntries(branch.items)) > 0 {
				leftAction, rightAction = "collapse", "child"
			}
		}
	}
	if len(m.visible) > 0 {
		switch m.visible[m.cursor].name {
		case ".":
			enterAction = "change dir"
		case "..":
			leftAction, rightAction, enterAction = "go back", "go back", "go back"
		}
	}
	hints := []keyHint{{"↑↓", "move"}, {"←", leftAction}}
	if m.mode == outlineView {
		hints = append(hints, keyHint{"Tab", "recent"})
	}
	if len(m.visible) == 0 || m.visible[m.cursor].isDir() {
		hints = append(hints, keyHint{"→", rightAction}, keyHint{"Enter", enterAction})
	} else {
		hints = append(hints, keyHint{"Enter", "open file"}, keyHint{"Space", "preview"})
	}
	if key := m.config.Keys.OpenInEditor; key != "" && m.selectedEditorCommand() != "" {
		hints = append(hints, keyHint{key, "Editor"})
	}
	hints = append(hints, keyHint{"n/N", "new dir/file"}, keyHint{"/", "filter"}, keyHint{"v", "view"}, keyHint{"?", "help"}, keyHint{"Esc", "cancel"})
	// Preserve Enter, help and exit when space is limited.
	trimOrder := []string{"↑↓", "←", "→", "/", "v", "n/N"}
	if m.mode == outlineView {
		trimOrder = []string{"/", "n/N", "↑↓", "v", "→", "←"}
	}
	trimOrder = append([]string{m.config.Keys.OpenInEditor}, trimOrder...)
	for _, key := range trimOrder {
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
	if m.recent.focused {
		return []viewLine{
			m.shortHelp(width),
			m.hints([]keyHint{{"Home / End", "first / last"}, {"PgUp / PgDn", "page"}}),
			m.hints([]keyHint{{"? / Esc", "close help"}}),
		}
	}
	var lines []viewLine
	var row []keyHint
	hints := append([]keyHint(nil), helpHints...)
	if m.mode == outlineView {
		hints = append([]keyHint{{"Tab / Shift+Tab", "tree / recent places"}}, hints...)
	}
	if runtime.GOOS == "darwin" {
		hints = append([]keyHint{{"f", "Open dir in Finder"}, {"Cmd+C", "copy file/dir"}, {"Cmd+V", "paste file/dir"}}, hints...)
	}
	if key := m.config.Keys.OpenInEditor; key != "" && m.selectedEditorCommand() != "" {
		hints = append([]keyHint{{key, "Open in editor"}}, hints...)
	}
	for _, hint := range hints {
		if m.mode == outlineView {
			switch hint.key {
			case "~":
				lines = append(lines, m.hints([]keyHint{{"\\", "tree from /"}}))
			case "→ / l":
				hint.action = "expand / first child"
			case "← / h / Backspace":
				hint.action = "collapse / parent"
			case "Enter":
				hint.action = "change dir / open file"
			case "Double click":
				hint.action = "toggle branch"
			case "Click":
				hint.action = "highlight; arrow toggles branch"
			case "/":
				hint.action = "filter expanded branches"
			}
		}
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
	lines = append(lines, viewLine{{"folder  ", m.styles.directory}, {"file  ", m.styles.file}, {"dir-link@", m.styles.link}})
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
	text := addressFor(path, m.home).label(width)
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
		if m.mode == outlineView {
			left, _ := m.outlineWidths(width)
			header[2] = viewLine{{fromPrefix, m.styles.muted}, {addressFor(m.origin, m.home).label(max(0, left-len(fromPrefix))), m.styles.origin}}
		}
	}
	footer = []viewLine{m.shortHelp(width)}
	if m.showHelp {
		footer = m.helpLines(width)
	}
	var status []viewLine
	if m.mutation != nil {
		status = append(status, viewLine{{"Rename: " + tail(display(m.mutation.path), max(1, width-8)), m.styles.muted}})
		status = append(status, viewLine{{"New name: ", m.styles.strong}, {tail(display(m.mutation.input)+"_", max(1, width-10)), m.styles.accent}})
	} else if m.creating {
		if m.mode == outlineView || m.creatingFile {
			status = append(status, viewLine{{"New in: " + tail(display(m.creationPath()), max(1, width-8)), m.styles.muted}})
		}
		label := "New: "
		if m.creatingFile {
			label = "New file: "
		}
		name := tail(display(m.newName)+"_", max(1, width-textWidth(label)))
		status = append(status, viewLine{{label, m.styles.strong}, {name, m.styles.accent}})
	} else if !m.recent.focused && (m.filtering || m.query != "") {
		query := "/ " + display(m.query)
		if m.mode == outlineView {
			query = "Filter expanded: " + display(m.query)
		}
		if m.filtering {
			query += "_"
		}
		status = append(status, viewLine{{query, m.styles.accent}})
	}
	if m.notice != "" {
		status = append([]viewLine{{{display(m.notice), m.styles.accent}}}, status...)
	}
	if m.message != "" {
		status = append([]viewLine{{{"Error: " + display(m.message), m.styles.failure}}}, status...)
	}
	// Even a tiny terminal must retain one row for the highlighted directory.
	maxFooter := max(0, m.height-len(header)-1)
	if m.height >= 8 {
		maxFooter = max(0, maxFooter-3)
	}
	status = status[:min(len(status), maxFooter)]
	actions := m.entryButtons()
	if len(actions) >= maxFooter-len(status) {
		actions = nil
	}
	helpRows := maxFooter - len(status) - len(actions)
	if len(footer) > helpRows {
		if m.showHelp && helpRows > 0 {
			footer = append(footer[:helpRows-1], footer[len(footer)-1])
		} else {
			footer = footer[:helpRows]
		}
	}
	footer = append(append(status, actions...), footer...)
	return header, footer
}

func (m browser) footerTop(footer []viewLine) int {
	bottom := max(1, m.height)
	if m.height >= 8 {
		bottom--
	}
	return bottom - len(footer)
}

func (m browser) pageRows() int {
	header, footer := m.frame()
	extra := 0
	if m.height >= 8 {
		extra = 3 // Separator, summary and bottom margin.
	}
	return max(1, m.height-len(header)-len(footer)-extra)
}

func (m browser) pageSize() int { return m.pageRows() }

func (m browser) pageStart() int {
	rows := m.pageRows()
	start := m.scroll
	if m.mode == outlineView {
		// Keep a clicked tree row under the pointer for a double-click.
		if m.lastClick.path == "" || m.recent.focused {
			start = m.cursor - (rows-1)/2
		}
		start = min(start, len(m.visible)-rows)
	}
	return max(0, min(start, m.cursor), m.cursor-rows+1)
}

func (m browser) paneWidths(width int) []int {
	if m.mode == panesView {
		return treeWidths(width)
	}
	return []int{width}
}

func (m browser) paneLocations(widths []int) []string {
	paths := []string{m.path}
	if m.mode == outlineView {
		if m.recent.focused {
			if len(m.recent.paths) == 0 {
				return []string{""}
			}
			return []string{m.recent.paths[m.recent.cursor]}
		}
		paths[0] = m.focusedPath()
		if len(m.visible) > 0 && !m.visible[m.cursor].isDir() {
			paths[0] = filepath.Dir(paths[0])
		}
	}
	if m.mode == panesView && len(widths) == 3 {
		parent, child := "", ""
		if path := filepath.Dir(m.path); path != m.path {
			parent = path
		}
		if m.cursor >= 0 && m.cursor < len(m.visible) {
			item := m.visible[m.cursor]
			if item.isDir() && !item.isDotDir() && item.path != m.path {
				child = item.path
			}
		}
		paths = []string{parent, m.path, child}
	}
	return paths
}

func (m browser) panePaths(widths []int, margin int) string {
	paths := m.paneLocations(widths)
	parts := make([]viewLine, len(paths))
	for i, path := range paths {
		parts[i] = m.pathLine(path, widths[i], path == m.path || m.mode == outlineView)
	}
	return m.paneRow(parts, widths, margin)
}

func (m browser) entryLine(index, width int) viewLine {
	if index >= len(m.visible) {
		return nil
	}
	if m.mode == outlineView {
		return m.outlineEntryLine(index, width)
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
	if item.isDotDir() {
		name := item.name
		description := "  (current dir)"
		if item.name == ".." {
			name = "←"
			description = " Back to parent"
			style = m.styles.accent
		}
		label = name + description
		row = viewLine{{"  " + name, ""}, {description, m.styles.muted}}
		if item.name == ".." {
			row = viewLine{{label, style}}
		}
	}
	if selected {
		text := "> " + display(label)
		if item.name == ".." {
			text = display(label)
			if m.styles.selected == "" {
				text += " >"
			}
		}
		text = truncate(text, width)
		text += strings.Repeat(" ", max(0, width-textWidth(text)))
		row = viewLine{{text, m.styles.selected + style}}
	}
	return row
}

func (m browser) view() string {
	if m.paste != nil {
		return m.pasteView()
	}
	if m.mutation != nil {
		return m.mutationView()
	}
	if m.filePreview != nil {
		return m.filePreviewView()
	}
	width, margin := m.viewWidth()
	header, footer := m.frame()
	rows := m.pageRows()
	widths := m.paneWidths(width)
	capacity := rows
	start := m.pageStart()
	if m.height >= 8 {
		header[0] = viewLine{{"cdd  " + m.mode.String(), m.styles.muted}, {"   v", m.styles.strong}, {" tree / panes", m.styles.muted}}
		if m.mode == outlineView && !m.recent.focused && width >= 14 {
			title := truncate("cdd  tree   v tree / panes", width-len(rootButton)-1)
			header[0] = viewLine{{title + strings.Repeat(" ", width-len(rootButton)-textWidth(title)), m.styles.muted}, {rootButton, m.styles.accent}}
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
		if m.mode == panesView {
			lines[1] = m.treeHeading(width, margin)
			pathRow = 2
		}
	}
	lines[pathRow] = m.panePaths(widths, margin)
	if m.mode == panesView {
		lines = append(lines, m.treeLines(rows, width, margin)...)
	} else {
		left, right := m.outlineWidths(width)
		if m.height >= 8 {
			style := m.styles.muted
			if m.recent.focused {
				style = m.styles.strong
			}
			recentHeader := viewLine{{"Recent places", style}}
			lines[2] = m.outlinePanelRow(header[2], recentHeader, left, right, margin)
		}
		lines = append(lines, m.outlineLines(rows, width, margin)...)
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
		for i, item := range m.visible {
			if item.isDotDir() || (m.mode == outlineView && i == 0) {
				continue
			}
			if item.isDir() {
				directories++
			} else {
				files++
			}
		}
		count := strconv.Itoa(directories) + " dirs"
		if directories == 1 {
			count = "1 dir"
		}
		if files == 1 {
			count += " / 1 file"
		} else {
			count += " / " + strconv.Itoa(files) + " files"
		}
		if m.query != "" && m.mode != outlineView {
			count = strconv.Itoa(directories+files) + " matches"
			if directories+files == 1 {
				count = "1 match"
			}
		}
		if m.mode == outlineView {
			count += " shown"
		}
		if start > 0 {
			count += "  |  more above"
		}
		if end < len(m.visible) {
			count += "  |  more below"
		}
		count += "  |  hidden " + hidden
		if m.recent.focused {
			count = strconv.Itoa(len(m.recent.paths)) + " recent places"
		}
		lines = append(lines, render(nil), render(viewLine{{count, m.styles.muted}}))
	}
	for len(lines) < m.footerTop(footer) {
		lines = append(lines, render(nil))
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

// UI arrows and box drawing use one cell. Reserve two for other non-ASCII runes
// to avoid wrapping filenames without a Unicode-width dependency. Combining
// marks and emoji sequences can leave unused space, but stay valid UTF-8.
func runeWidth(r rune) int {
	switch r {
	case '←', '→', '↑', '↓', '│', '─', '╭', '╮', '╰', '╯', '├', '└', '▸', '▾':
		return 1
	}
	if r > 127 {
		return 2
	}
	return 1
}

func textWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

func truncate(s string, width int) string {
	cells := 0
	for i, r := range s {
		cost := runeWidth(r)
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
