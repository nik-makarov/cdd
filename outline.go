package main

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type viewMode uint8

const (
	panesView viewMode = iota
	outlineView
)

func (v viewMode) String() string {
	switch v {
	case outlineView:
		return "tree"
	default:
		return "panes"
	}
}

type treeBranch struct {
	items    []entry
	expanded bool
}

type outlineRow struct {
	parent int
	guide  string
}

func (m browser) focusedPath() string {
	if m.cursor >= 0 && m.cursor < len(m.visible) {
		return m.visible[m.cursor].path
	}
	return m.path
}

func (m *browser) focusPath(path string) {
	for i, item := range m.visible {
		if item.path == path {
			m.cursor = i
			return
		}
	}
}

func (m *browser) cycleView() {
	previousMode := m.mode
	defer func() {
		if m.mode != previousMode && m.viewFile != "" {
			if err := writeViewPreference(m.viewFile, m.mode); err != nil {
				m.message = "Could not save view: " + err.Error()
			}
		}
	}()
	focus := m.visible[m.cursor]
	next := outlineView
	if m.mode == outlineView {
		next = panesView
	}
	if next == outlineView {
		if err := m.prepareOutline(focus.path); err != nil {
			m.message = err.Error()
		}
		return
	}
	if m.mode == outlineView {
		// Keep a nested destination in view when returning to a flat listing.
		path, name := filepath.Dir(focus.path), focus.name
		if focus.path == m.treeRoot || focus.path == m.path {
			path, name = focus.path, ""
		}
		previous := m.mode
		m.mode = next
		if err := m.open(path, name); err != nil {
			m.mode, m.message = previous, err.Error()
		}
		return
	}
	m.mode = next
	m.refilter()
	m.focusPath(focus.path)
	m.refreshContext()
}

func pathWithin(path, parent string) bool {
	if !filepath.IsAbs(parent) {
		return false
	}
	rel, err := filepath.Rel(parent, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func treeRootFor(path, home string) string {
	if pathWithin(path, home) {
		return filepath.Clean(home)
	}
	return filepath.VolumeName(path) + string(filepath.Separator)
}

// Load only the ancestor chains needed to reveal the browsing location and
// starting dir. Other branches remain lazy, including directory symlinks.
func (m *browser) prepareOutline(focus string) error {
	home, _ := os.UserHomeDir()
	root := treeRootFor(m.path, home)
	if m.rootTree {
		root = treeRootFor(m.path, "")
	}
	if !pathWithin(focus, root) {
		root = treeRootFor(focus, home)
	}
	branches := make(map[string]treeBranch)
	if m.treeRoot != "" && pathWithin(m.treeRoot, root) {
		branches = maps.Clone(m.branches)
		if branches == nil {
			branches = make(map[string]treeBranch)
		}
	}
	reveal := func(path string) error {
		var chain []string
		for dir := path; ; dir = filepath.Dir(dir) {
			chain = append(chain, dir)
			if dir == root {
				break
			}
		}
		slices.Reverse(chain)
		for _, dir := range chain {
			branch, loaded := branches[dir]
			if dir == m.path {
				branch.items = m.entries
			} else if !loaded {
				items, err := readEntries(dir)
				if err != nil {
					return err
				}
				branch.items = items
			}
			branch.expanded = true
			branches[dir] = branch
		}
		return nil
	}
	if err := reveal(m.path); err != nil {
		return err
	}
	if focus != root {
		if err := reveal(filepath.Dir(focus)); err != nil {
			return err
		}
	}
	if m.origin != root && pathWithin(m.origin, root) {
		// The source may have disappeared since launch; its header still explains
		// where the shell started even if that optional branch cannot be read.
		_ = reveal(filepath.Dir(m.origin))
	}
	m.treeRoot, m.branches, m.mode = root, branches, outlineView
	m.message, m.scroll, m.lastClick = "", 0, mouseClick{}
	m.rebuildOutline(focus)
	return nil
}

func (m *browser) showRootTree() {
	next := *m
	next.rootTree = true
	next.query, next.filtering = "", false
	if err := next.prepareOutline(m.focusedPath()); err != nil {
		m.message = err.Error()
		return
	}
	*m = next
}

const rootButton = "[/ Root]"

func (m browser) rootButtonAt(x, y int) bool {
	width, margin := m.viewWidth()
	return m.mode == outlineView && !m.recent.focused && m.height >= 8 && width >= 14 &&
		y == 0 && x >= margin+width-len(rootButton) && x < margin+width
}

func (m browser) outlineEntries(items []entry) []entry {
	if m.hidden {
		return items
	}
	return slices.DeleteFunc(slices.Clone(items), func(item entry) bool {
		return strings.HasPrefix(item.name, ".") && !pathWithin(m.path, item.path) && !pathWithin(m.origin, item.path)
	})
}

// Only expanded branches are flattened. Reading a child dir is an explicit
// expansion action, never a side effect of rendering or filtering the tree.
func (m *browser) rebuildOutline(focus string) {
	if m.treeRoot == "" {
		m.treeRoot = m.path
	}
	if m.branches == nil {
		m.branches = map[string]treeBranch{m.treeRoot: {items: m.entries, expanded: true}}
	}
	m.visible, m.outline = nil, nil
	appendRow := func(item entry, parent int, guide string) int {
		index := len(m.visible)
		m.visible = append(m.visible, item)
		m.outline = append(m.outline, outlineRow{parent: parent, guide: guide})
		return index
	}
	root := appendRow(entry{name: filepath.Base(m.treeRoot), path: m.treeRoot}, -1, "")
	var walk func(string, int, string)
	walk = func(path string, parent int, guide string) {
		branch := m.branches[path]
		if !branch.expanded {
			return
		}
		items := m.outlineEntries(branch.items)
		for i, item := range items {
			connector, next := "├─ ", guide+"│  "
			if i == len(items)-1 {
				connector, next = "└─ ", guide+"   "
			}
			index := appendRow(item, parent, guide+connector)
			if item.isDir() {
				walk(item.path, index, next)
			}
		}
	}
	walk(m.treeRoot, root, "")
	if m.query != "" {
		keep := make([]bool, len(m.visible))
		query := strings.ToLower(m.query)
		for i, item := range m.visible {
			if i == 0 || item.path == m.origin || strings.Contains(strings.ToLower(item.name), query) {
				keep[i] = true
			}
		}
		for i := len(keep) - 1; i >= 0; i-- {
			if keep[i] && m.outline[i].parent >= 0 {
				keep[m.outline[i].parent] = true
			}
		}
		var items []entry
		var rows []outlineRow
		indices := make([]int, len(keep))
		for i, retained := range keep {
			if !retained {
				continue
			}
			row := m.outline[i]
			if row.parent >= 0 {
				row.parent = indices[row.parent]
			}
			indices[i] = len(items)
			items, rows = append(items, m.visible[i]), append(rows, row)
		}
		m.visible, m.outline = items, rows
	}
	m.cursor = 0
	m.focusPath(m.path)
	// If filtering or hiding removes focus, keep its closest visible ancestor.
	for path := focus; path != ""; path = filepath.Dir(path) {
		m.focusPath(path)
		if m.focusedPath() == path || filepath.Dir(path) == path {
			break
		}
	}
}

func (m *browser) clearOutlineFilter() {
	focus := m.focusedPath()
	m.query, m.filtering, m.message = "", false, ""
	m.rebuildOutline(focus)
}

func (m *browser) expandOutline() {
	item := m.visible[m.cursor]
	branch, loaded := m.branches[item.path]
	if !loaded {
		items, err := readEntries(item.path)
		if err != nil {
			m.message = err.Error()
			return
		}
		branch.items = items
	}
	branch.expanded = true
	m.branches[item.path] = branch
	m.rebuildOutline(item.path)
}

func (m *browser) outlineRight() {
	m.clearOutlineFilter()
	item := m.visible[m.cursor]
	if !m.branches[item.path].expanded {
		m.expandOutline()
		return
	}
	if next := m.cursor + 1; next < len(m.outline) && m.outline[next].parent == m.cursor {
		m.cursor = next
	}
}

func (m *browser) toggleOutline() {
	m.clearOutlineFilter()
	item := m.visible[m.cursor]
	branch := m.branches[item.path]
	if !branch.expanded {
		m.expandOutline()
		return
	}
	branch.expanded = false
	m.branches[item.path] = branch
	m.rebuildOutline(item.path)
}

func (m *browser) outlineLeft() {
	m.clearOutlineFilter()
	item := m.visible[m.cursor]
	if branch := m.branches[item.path]; branch.expanded && len(m.outlineEntries(branch.items)) > 0 {
		m.toggleOutline()
		return
	}
	if parent := m.outline[m.cursor].parent; parent >= 0 {
		m.cursor = parent
	}
}

func (m *browser) refreshOutline() {
	items, err := readEntries(m.treeRoot)
	if err != nil {
		m.message = err.Error()
		return
	}
	focus, previous := m.focusedPath(), m.branches
	m.message = ""
	m.branches = make(map[string]treeBranch)
	var reload func(string, []entry)
	reload = func(path string, items []entry) {
		if path == m.path {
			m.entries = items
		}
		branch := treeBranch{items: items, expanded: previous[path].expanded}
		m.branches[path] = branch
		if !branch.expanded {
			return
		}
		for _, item := range items {
			if !item.isDir() || !previous[item.path].expanded {
				continue
			}
			children, err := readEntries(item.path)
			if err != nil {
				if m.message == "" {
					m.message = err.Error()
				}
				continue
			}
			reload(item.path, children)
		}
	}
	reload(m.treeRoot, items)
	m.rebuildOutline(focus)
}

func (m browser) creationPath() string {
	if m.mode != outlineView && !m.creatingFile {
		return m.path
	}
	return m.highlightedDirectory()
}

func (m browser) highlightedDirectory() string {
	item := m.visible[m.cursor]
	if item.isDir() {
		return item.path
	}
	return filepath.Dir(item.path)
}

func (m browser) outlineGuide(index, width int) string {
	if index >= len(m.outline) {
		return ""
	}
	return tail(m.outline[index].guide, max(0, width-16))
}

func (m browser) outlineEntryLine(index, width int) viewLine {
	item := m.visible[index]
	guide := m.outlineGuide(index, width)
	marker := "  "
	if item.isDir() {
		marker = "▸ "
		if branch, loaded := m.branches[item.path]; loaded {
			if len(m.outlineEntries(branch.items)) == 0 {
				marker = "  "
			} else if branch.expanded {
				marker = "▾ "
			}
		}
	}
	row := m.itemLine(item, max(1, width-2-textWidth(guide)), false)
	// Replace the ordinary list gutter with the tree's disclosure marker.
	row[0].text = marker + strings.TrimPrefix(row[0].text, "  ")
	if item.path == m.origin {
		row[0].style = m.styles.origin
		row = append(row, span{" (starting dir)", m.styles.origin})
	} else if item.path == m.treeRoot && m.treeRoot != string(filepath.Separator) {
		row = append(row, span{" (home)", m.styles.muted})
	}
	if branch, loaded := m.branches[item.path]; loaded && len(m.outlineEntries(branch.items)) == 0 {
		label := " (empty)"
		if len(branch.items) > 0 {
			label = " (hidden entries)"
		}
		row = append(row, span{label, m.styles.muted})
	}
	gutter := "  "
	if !m.recent.focused && index == m.cursor {
		gutter = "> "
	}
	row = append(viewLine{{gutter + guide, m.styles.muted}}, row...)
	if !m.recent.focused && index == m.cursor {
		used := 0
		for i := range row {
			if item.path != m.origin {
				row[i].style = strings.ReplaceAll(row[i].style, m.styles.muted, "")
			}
			row[i].style = m.styles.selected + row[i].style
			used += textWidth(display(row[i].text))
		}
		row = append(row, span{strings.Repeat(" ", max(0, width-used)), m.styles.selected})
	}
	return row
}

func (m browser) outlineDisclosureAt(index, x int) bool {
	item := m.visible[index]
	if !item.isDir() {
		return false
	}
	if branch, loaded := m.branches[item.path]; loaded && len(m.outlineEntries(branch.items)) == 0 {
		return false
	}
	width, margin := m.viewWidth()
	width, _ = m.outlineWidths(width)
	left := margin + 2 + textWidth(m.outlineGuide(index, width))
	return x >= left && x < min(left+2, margin+width)
}
