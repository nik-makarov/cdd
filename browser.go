package main

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
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
	query     string
	message   string
	selected  string
	width     int
	height    int
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
		return fmt.Errorf("read directory %q: %w", path, err)
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

func (m browser) pageSize() int { return max(1, m.height-8) }

func (m browser) view() string {
	var lines []string
	lines = append(lines, "0go  /  choose a directory", "Path: "+display(m.path), "")
	start := max(0, m.cursor-m.pageSize()+1)
	end := min(len(m.visible), start+m.pageSize())
	for i := start; i < end; i++ {
		label := display(m.visible[i].name) + "/"
		if i == 0 {
			label = ".  (current directory)"
		}
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		lines = append(lines, prefix+label)
	}
	hidden := "off"
	if m.hidden {
		hidden = "on"
	}
	lines = append(lines, "", fmt.Sprintf("%d directories  ·  hidden: %s  ·  %d/%d", len(m.visible)-1, hidden, m.cursor+1, len(m.visible)))
	if m.filtering || m.query != "" {
		lines = append(lines, "Filter: /"+display(m.query))
	} else {
		lines = append(lines, "↑↓/jk move  →/l open  ←/h parent  Enter select")
	}
	if m.message != "" {
		lines = append(lines, display(m.message))
	} else {
		lines = append(lines, "/ filter  . hidden  ~ home  Ctrl+R refresh  Esc/q cancel")
	}
	for i, line := range lines {
		lines[i] = truncate(line, max(1, m.width))
	}
	return strings.Join(lines[:min(len(lines), max(1, m.height))], "\n")
}

// Reserve two cells for every non-ASCII rune. This conservative bound avoids
// wrapping wide characters without a Unicode-width dependency. Combining marks
// and emoji sequences can leave unused space, but filenames stay valid UTF-8.
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
