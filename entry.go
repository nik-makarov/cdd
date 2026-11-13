package main

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type entryKind uint8

const (
	folder entryKind = iota
	file
	folderLink
	fileLink
	brokenLink
)

type entry struct {
	name, path string
	kind       entryKind
}

func (e entry) isDir() bool { return e.kind == folder || e.kind == folderLink }

func (e entry) label() string {
	switch e.kind {
	case folder:
		return e.name + "/"
	case folderLink:
		return e.name + "@/"
	case fileLink:
		return e.name + "@"
	case brokenLink:
		return e.name + "@!"
	default:
		return e.name
	}
}

func readEntries(path string) ([]entry, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, wrapError("read directory "+strconv.Quote(path), err)
	}
	entries := make([]entry, 0, len(items))
	for _, item := range items {
		e := entry{name: item.Name(), path: filepath.Join(path, item.Name()), kind: file}
		if item.IsDir() {
			e.kind = folder
		}
		if item.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(e.path)
			switch {
			case err != nil:
				e.kind = brokenLink
			case info.IsDir():
				e.kind = folderLink
			default:
				e.kind = fileLink
			}
		}
		entries = append(entries, e)
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if a.isDir() != b.isDir() {
			if a.isDir() {
				return -1
			}
			return 1
		}
		return cmp.Or(strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name)), strings.Compare(a.name, b.name))
	})
	return entries, nil
}
