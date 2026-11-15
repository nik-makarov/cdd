package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type entryMutation struct {
	kind, path, resolved, input string
	info                        os.FileInfo
	replace                     bool
	busy                        bool
}

func withinPath(parent, path string) bool {
	rel, err := filepath.Rel(parent, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (m browser) mutationAllowed() bool {
	if len(m.visible) == 0 {
		return false
	}
	item := m.visible[m.cursor]
	return !item.isDotDir() && filepath.IsAbs(item.path) &&
		!withinPath(item.path, m.path) && !withinPath(item.path, m.origin)
}

func (m browser) mutationTarget() (*entryMutation, error) {
	if !m.mutationAllowed() {
		return nil, errors.New("Cannot change the current or starting dir, their parents, or navigation rows")
	}
	path := m.focusedPath()
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	resolved := filepath.Join(parent, filepath.Base(path))
	info, err := os.Lstat(resolved)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		for _, anchor := range []string{m.path, m.origin} {
			actual, err := filepath.EvalSymlinks(anchor)
			if err != nil {
				return nil, err
			}
			if withinPath(resolved, actual) {
				return nil, errors.New("Cannot change a dir containing the current or starting location")
			}
		}
	}
	return &entryMutation{path: path, resolved: resolved, info: info}, nil
}

func (edit entryMutation) validate() error {
	parent, err := filepath.EvalSymlinks(filepath.Dir(edit.path))
	if err != nil {
		return err
	}
	info, err := os.Lstat(edit.resolved)
	if err != nil {
		return err
	}
	if parent != filepath.Dir(edit.resolved) || !os.SameFile(edit.info, info) {
		return errors.New("Entry changed while the prompt was open; cancel and try again")
	}
	return nil
}

func (m *browser) beginRename() {
	edit, err := m.mutationTarget()
	if err != nil {
		m.message = err.Error()
		return
	}
	edit.kind, edit.input, edit.replace = "rename", filepath.Base(edit.path), true
	m.mutation, m.message, m.showHelp = edit, "", false
	m.lastClick = mouseClick{}
}

func (m *browser) handleMutationInput(key keyEvent) {
	edit := m.mutation
	if edit.busy {
		return
	}
	switch key.name {
	case "esc":
		m.mutation, m.message = nil, ""
	case "enter":
		if edit.kind == "trash" {
			m.confirmTrash()
		} else {
			m.renameEntry()
		}
	case "backspace":
		if edit.replace {
			edit.input = ""
		} else if runes := []rune(edit.input); len(runes) > 0 {
			edit.input = string(runes[:len(runes)-1])
		}
		edit.replace, m.message = false, ""
	default:
		if key.text != "" {
			if edit.replace {
				edit.input = ""
			}
			edit.input += key.text
			edit.replace, m.message = false, ""
		}
	}
}

func (m *browser) renameEntry() {
	edit := m.mutation
	name := edit.input
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\x00") {
		m.message = "Enter a single name, not a path, . or .."
		return
	}
	if err := edit.validate(); err != nil {
		m.message = err.Error()
		return
	}
	if name == filepath.Base(edit.path) {
		m.mutation, m.message = nil, ""
		return
	}
	to := filepath.Join(filepath.Dir(edit.resolved), name)
	if err := renameExclusive(edit.resolved, to); err != nil {
		m.message = wrapError("rename", err).Error()
		return
	}
	m.mutation = nil
	if strings.HasPrefix(name, ".") {
		m.hidden = true
	}
	m.refreshMutation(edit.path, filepath.Join(filepath.Dir(edit.path), name))
}

func (m *browser) refreshMutation(oldPath, focus string) {
	m.query, m.message, m.filtering = "", "", false
	if m.mode == outlineView {
		for path := range m.branches {
			if withinPath(oldPath, path) {
				delete(m.branches, path)
			}
		}
		m.refreshOutline()
		m.focusPath(focus)
		return
	}
	if err := m.open(m.path, filepath.Base(focus)); err != nil {
		m.message = "Entry changed, but refreshing failed: " + err.Error()
	}
}
