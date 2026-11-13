package main

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type directory struct {
	name string
	path string
}

type browser struct {
	path      string
	dirs      []directory
	visible   []directory
	cursor    int
	hidden    bool
	filtering bool
	showHelp  bool
	query     string
	message   string
	selected  string
	width     int
	height    int
	styles    viewStyles
}

func newBrowser(path string, hidden bool) (browser, error) {
	m := browser{hidden: hidden, width: 80, height: 24}
	if err := m.open(path, ""); err != nil {
		return browser{}, err
	}
	return m, nil
}

func (m *browser) open(path, focus string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return wrapError("read directory "+strconv.Quote(path), err)
	}
	dirs := make([]directory, 0, len(entries))
	for _, entry := range entries {
		isDir := entry.IsDir()
		if entry.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(filepath.Join(path, entry.Name()))
			// Broken or inaccessible symlinks are not navigable directories.
			isDir = err == nil && info.IsDir()
		}
		if isDir {
			dirs = append(dirs, directory{entry.Name(), filepath.Join(path, entry.Name())})
		}
	}
	slices.SortFunc(dirs, func(a, b directory) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name)), strings.Compare(a.name, b.name))
	})
	m.path, m.dirs = path, dirs
	m.query, m.message, m.filtering = "", "", false
	m.refilter()
	for i, dir := range m.visible {
		if dir.name == focus {
			m.cursor = i
			break
		}
	}
	return nil
}

func (m *browser) refilter() {
	m.visible = []directory{{".", m.path}}
	for _, dir := range m.dirs {
		if !m.hidden && strings.HasPrefix(dir.name, ".") {
			continue
		}
		if strings.Contains(strings.ToLower(dir.name), strings.ToLower(m.query)) {
			m.visible = append(m.visible, dir)
		}
	}
	m.cursor = 0
	if m.query != "" && len(m.visible) > 1 {
		m.cursor = 1
	}
}

// handle returns true when the user selects a directory or cancels.
func (m *browser) handle(key keyEvent) bool {
	if key.name == "ctrl+c" || key.name == "ctrl+d" {
		return true
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
		case "down", "up":
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
	switch key.name {
	case "esc", "q":
		return true
	case "up", "k", "down", "j", "home", "end", "pgup", "pgdown":
		m.move(key.name)
	case "right", "l":
		m.navigate(m.visible[m.cursor].path, "")
	case "left", "h", "backspace":
		m.navigate(filepath.Dir(m.path), filepath.Base(m.path))
	case "enter":
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
	}
	return false
}

func (m *browser) navigate(path, focus string) {
	if err := m.open(path, focus); err != nil {
		m.message = err.Error()
	}
}

func (m *browser) move(key string) {
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
	}
	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
}
