package main

import "strconv"

type pastePrompt struct {
	entry  *entryCopy
	busy   bool
	scroll int
}

func (m *browser) clipboardEntry(key string) {
	if len(m.visible) == 0 {
		return
	}
	m.message, m.notice = "", ""
	if key == "copy-entry" {
		m.action = externalAction{kind: key, path: m.focusedPath()}
		return
	}
	m.showHelp = false
	m.paste = &pastePrompt{}
	m.action = externalAction{kind: "prepare-paste", path: m.highlightedDirectory()}
}

func (m *browser) handlePaste(key keyEvent) {
	prompt := m.paste
	if prompt.busy {
		return
	}
	switch key.name {
	case "esc":
		m.paste, m.message = nil, ""
	case "enter":
		if prompt.entry == nil || m.height < 3 || len(m.pastePathRows()) == 0 {
			return
		}
		prompt.busy = true
		m.action = externalAction{kind: "paste-entry", paste: prompt.entry}
	case "up", "down":
		if prompt.entry != nil {
			if key.name == "up" {
				prompt.scroll = max(0, prompt.scroll-1)
			} else {
				prompt.scroll = min(prompt.scroll+1, max(0, len(m.pastePathRows())-max(1, m.height-2)))
			}
		}
	}
}

func (m *browser) finishPaste(p *entryCopy, err error) {
	m.paste = nil
	if err != nil {
		m.message = err.Error()
		return
	}
	if m.mode == outlineView {
		m.refreshOutline()
	} else {
		focus := m.focusedPath()
		items, err := readEntries(m.path)
		if err != nil {
			m.message = err.Error()
		} else {
			m.entries = items
			m.refilter()
			m.focusPath(focus)
			m.refreshContext()
		}
	}
	m.notice = "Pasted: " + p.destination
}

// Quote control characters and wrap long paths so confirmation never hides a
// parent directory, a filename suffix, or a literal newline.
func (m browser) pastePathRows() []viewLine {
	width, _ := m.viewWidth()
	path := strconv.Quote(m.paste.entry.destination)
	var rows []viewLine
	for path != "" {
		line := truncate(path, width)
		if line == "" { // The terminal cannot fit even one wide character.
			return nil
		}
		rows = append(rows, viewLine{{line, m.styles.accent}})
		path = path[len(line):]
	}
	return rows
}

func (m browser) pasteView() string {
	width, margin := m.viewWidth()
	prompt := m.paste
	if prompt.entry == nil {
		return m.modalView([]viewLine{{{"Reading file clipboard...", m.styles.strong}}, m.hints([]keyHint{{"Esc", "cancel"}})}, width, margin)
	}
	pathRows := m.pastePathRows()
	if m.height < 3 || len(pathRows) == 0 {
		return m.modalView([]viewLine{{{"Enlarge terminal to review the destination", m.styles.strong}}, m.hints([]keyHint{{"Esc", "cancel"}})}, width, margin)
	}
	title := "Paste to"
	if prompt.busy {
		title = "Copying to"
	}
	visible := m.height - 2
	if len(pathRows) > visible {
		title += " (Up/Down to scroll)"
		start := min(prompt.scroll, len(pathRows)-visible)
		pathRows = pathRows[start : start+visible]
	}
	rows := []viewLine{{{title, m.styles.strong}}}
	rows = append(rows, pathRows...)
	if prompt.busy {
		rows = append(rows, m.hints([]keyHint{{"Ctrl+C", "exit"}}))
	} else {
		rows = append(rows, m.hints([]keyHint{{"Enter", "paste"}, {"Esc", "cancel"}}))
	}
	// Show the source when the complete destination and controls already fit.
	if len(rows)+2 <= m.height {
		source := "From " + strconv.Quote(prompt.entry.source)
		rows = append([]viewLine{{{tail(source, width), m.styles.muted}}, nil}, rows...)
	}
	return m.modalView(rows, width, margin)
}
