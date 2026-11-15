package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type browser struct {
	path      string
	entries   []entry
	visible   []entry
	cursor    int
	scroll    int
	compact   bool
	lastClick mouseClick
	hidden    bool
	filtering bool
	creating  bool
	showHelp  bool
	query     string
	newName   string
	message   string
	selected  string
	width     int
	height    int
	styles    viewStyles
	parent    []entry
	children  []entry
	parentErr string
	childErr  string
	preview   string
}

func newBrowser(path string, hidden bool) (browser, error) {
	m := browser{hidden: hidden, width: 80, height: 24}
	if err := m.open(path, ""); err != nil {
		return browser{}, err
	}
	return m, nil
}

func (m *browser) open(path, focus string) error {
	entries, err := readEntries(path)
	if err != nil {
		return err
	}
	m.path, m.entries = path, entries
	m.query, m.message, m.filtering = "", "", false
	m.newName, m.creating = "", false
	m.refilter()
	for i, dir := range m.visible {
		if dir.name == focus {
			m.cursor = i
			break
		}
	}
	m.refreshContext()
	return nil
}

func (m *browser) refilter() {
	m.message = ""
	m.scroll, m.lastClick = 0, mouseClick{}
	m.visible = nil
	if parent := filepath.Dir(m.path); parent != m.path {
		m.visible = append(m.visible, entry{name: "..", path: parent})
	}
	m.cursor = len(m.visible)
	m.visible = append(m.visible, entry{name: ".", path: m.path})
	firstMatch := len(m.visible)
	for _, dir := range m.entries {
		if m.compact && !dir.isDir() {
			continue
		}
		if !m.hidden && strings.HasPrefix(dir.name, ".") {
			continue
		}
		if strings.Contains(strings.ToLower(dir.name), strings.ToLower(m.query)) {
			m.visible = append(m.visible, dir)
		}
	}
	if m.query != "" && len(m.visible) > firstMatch {
		m.cursor = firstMatch
	}
}

// handle returns true when the user selects a directory or cancels.
func (m *browser) handle(key keyEvent) bool {
	if key.name == "ctrl+c" || key.name == "ctrl+d" {
		return true
	}
	m.scroll = m.pageStart()
	defer func() {
		m.scroll = m.pageStart()
		m.refreshPreview(false)
	}()
	if strings.HasPrefix(key.name, "mouse-") {
		return m.handleMouse(key)
	}
	m.lastClick = mouseClick{}
	if m.creating {
		switch key.name {
		case "esc":
			m.creating, m.newName, m.message = false, "", ""
		case "enter":
			m.createDirectory()
		case "backspace":
			runes := []rune(m.newName)
			if len(runes) > 0 {
				m.newName = string(runes[:len(runes)-1])
			}
			m.message = ""
		default:
			if key.text != "" {
				m.newName += key.text
				m.message = ""
			}
		}
		return false
	}
	if m.filtering {
		switch key.name {
		case "esc", "enter":
			m.filtering = false
		case "backspace":
			runes := []rune(m.query)
			if len(runes) > 0 {
				m.query = string(runes[:len(runes)-1])
			}
			m.refilter()
		case "down", "up", "tab", "backtab":
			m.move(key.name)
		default:
			if key.text != "" {
				m.query += key.text
				m.refilter()
			}
		}
		return false
	}
	if m.showHelp && key.name == "esc" {
		m.showHelp = false
		return false
	}
	if key.name == "enter" && m.visible[m.cursor].name == "." {
		key.name = "confirm"
	}
	switch key.name {
	case "esc", "q":
		return true
	case "up", "k", "down", "j", "home", "end", "pgup", "pgdown", "tab", "backtab":
		m.move(key.name)
	case "right", "l", "enter":
		m.openSelected()
	case "left", "h", "backspace":
		parent := slices.IndexFunc(m.visible, func(item entry) bool { return item.name == ".." })
		if parent < 0 {
			m.message = ""
			break
		}
		if m.cursor != parent {
			m.cursor = parent
			m.message = ""
			break
		}
		m.openSelected()
	case "confirm":
		if !m.visible[m.cursor].isDir() {
			m.message = "Only directories can be chosen as a destination"
			break
		}
		path := m.visible[m.cursor].path
		info, err := os.Stat(path)
		if err != nil {
			m.message = err.Error()
			break
		}
		if !info.IsDir() {
			m.message = "Directory no longer exists; press Ctrl+R to refresh"
			break
		}
		m.selected = path
		return true
	case "/":
		m.showHelp = false
		m.filtering = true
		m.query = ""
		m.refilter()
	case "n":
		m.creating, m.showHelp = true, false
		m.newName, m.message = "", ""
	case ".":
		m.hidden = !m.hidden
		m.refilter()
	case "~":
		home, err := os.UserHomeDir()
		if err != nil {
			m.message = err.Error()
			break
		}
		m.navigate(home, "")
	case "ctrl+r":
		m.navigate(m.path, m.visible[m.cursor].name)
	case "?":
		m.showHelp = !m.showHelp
	case "v":
		focus := m.visible[m.cursor].path
		m.compact = !m.compact
		m.refilter()
		for i, item := range m.visible {
			if item.path == focus {
				m.cursor = i
				break
			}
		}
		m.message = ""
		m.refreshContext()
	}
	return false
}

func (m *browser) createDirectory() {
	name := m.newName
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\x00") {
		m.message = "Enter a single directory name, not a path, . or .."
		return
	}
	if err := os.Mkdir(filepath.Join(m.path, name), 0o755); err != nil {
		m.message = wrapError("create directory", err).Error()
		return
	}
	m.creating, m.newName = false, ""
	if strings.HasPrefix(name, ".") {
		m.hidden = true
	}
	if err := m.open(m.path, name); err != nil {
		m.message = "Directory created, but refreshing failed: " + err.Error()
	}
}

func (m *browser) refreshContext() {
	if m.compact {
		return
	}
	m.parent, m.parentErr = nil, ""
	if parent := filepath.Dir(m.path); parent != m.path {
		var err error
		m.parent, err = readEntries(parent)
		if err != nil {
			m.parentErr = err.Error()
		}
	} else {
		m.parentErr = "Filesystem root"
	}
	m.refreshPreview(true)
}

func (m *browser) refreshPreview(force bool) {
	if m.compact || len(m.visible) == 0 {
		return
	}
	item := m.visible[m.cursor]
	if !force && m.preview == item.path {
		return
	}
	m.preview, m.children, m.childErr = item.path, nil, ""
	if !item.isDir() {
		m.childErr = "File"
		return
	}
	if item.isDotDir() || item.path == m.path {
		return
	}
	var err error
	m.children, err = readEntries(item.path)
	if err != nil {
		m.childErr = err.Error()
	}
}

func (m *browser) navigate(path, focus string) {
	if err := m.open(path, focus); err != nil {
		m.message = err.Error()
	}
}

func (m *browser) openSelected() {
	item := m.visible[m.cursor]
	if !item.isDir() {
		m.message = "Files cannot be opened; choose a directory"
		return
	}
	focus := ""
	if item.name == ".." {
		focus = filepath.Base(m.path)
	}
	m.navigate(item.path, focus)
}

func (m *browser) move(key string) {
	m.message = ""
	switch key {
	case "up", "k":
		m.cursor--
	case "down", "j":
		m.cursor++
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.visible) - 1
	case "pgup":
		m.cursor -= m.pageSize()
	case "pgdown":
		m.cursor += m.pageSize()
	case "tab", "backtab":
		if m.columns() == 2 {
			rows := m.pageRows()
			pageStart := m.cursor / (rows * 2) * (rows * 2)
			if m.cursor-pageStart < rows {
				if pageStart+rows < len(m.visible) {
					m.cursor = min(m.cursor+rows, len(m.visible)-1)
				}
			} else {
				m.cursor -= rows
			}
		}
	}
	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
}
