package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEntryKindsAndDirectoryLinks(t *testing.T) {
	root := fixture(t)
	for name, target := range map[string]string{
		"folder-link": "Alpha", "file-link": "regular-file", "broken": "missing", "loop": "loop",
	} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	items, err := readEntries(root)
	if err != nil {
		t.Fatal(err)
	}
	byName := make(map[string]entry)
	sawFile := false
	for _, item := range items {
		byName[item.name] = item
		if !item.isDir() {
			sawFile = true
		} else if sawFile {
			t.Fatal("directories must sort before files")
		}
	}
	for _, tt := range []struct {
		name, label string
		kind        entryKind
		directory   bool
	}{
		{"Alpha", "Alpha/", folder, true},
		{"regular-file", "regular-file", file, false},
		{"folder-link", "folder-link@/", folderLink, true},
		{"file-link", "file-link@", fileLink, false},
		{"broken", "broken@!", brokenLink, false},
		{"loop", "loop@!", brokenLink, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			item, ok := byName[tt.name]
			if !ok || item.kind != tt.kind || item.isDir() != tt.directory || item.label() != tt.label || item.path != filepath.Join(root, tt.name) {
				t.Fatalf("unexpected entry: %+v", item)
			}
		})
	}
}
