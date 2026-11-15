package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func viewPreferencePath() string {
	dir, err := os.UserConfigDir()
	if err != nil || !filepath.IsAbs(dir) {
		return ""
	}
	return filepath.Join(dir, "cdd", "last-view")
}

func readViewPreference(path string) viewMode {
	data, err := os.ReadFile(path)
	if err != nil {
		return outlineView
	}
	switch strings.TrimSpace(string(data)) {
	case "tree":
		return outlineView
	case "panes":
		return panesView
	default:
		return outlineView
	}
}

func writeViewPreference(path string, mode viewMode) error {
	if mode != panesView && mode != outlineView {
		return errors.New("unknown view mode")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".last-view-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(mode.String() + "\n")
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func (m *browser) restoreViewPreference(path string) {
	m.viewFile = path
	switch readViewPreference(path) {
	case outlineView:
		if err := m.prepareOutline(m.path); err != nil {
			m.message = "Could not restore tree view: " + err.Error()
		}
	}
}
