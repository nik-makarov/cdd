package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

const recentLimit = 50

type recentPlaces struct {
	paths    []string
	cursor   int
	scroll   int
	focused  bool
	revealed string // Keep hidden components of this path visible in the tree.
}

func readRecent(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var places []string
	// NUL is the only byte that cannot occur in a Unix path.
	for dir := range strings.SplitSeq(string(data), "\x00") {
		if filepath.IsAbs(dir) {
			dir = filepath.Clean(dir)
			if !slices.Contains(places, dir) {
				places = append(places, dir)
				if len(places) == recentLimit {
					break
				}
			}
		}
	}
	return places, nil
}

func recordRecent(path, dir string) error {
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(dir) || strings.IndexByte(dir, 0) >= 0 {
		return errors.New("recent places require an absolute directory path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	// Lock a separate, stable inode while merging selections from other shells.
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lock.Close()
	for {
		err = unix.Flock(int(lock.Fd()), unix.LOCK_EX)
		if !errors.Is(err, unix.EINTR) {
			break
		}
	}
	if err != nil {
		return err
	}
	places, err := readRecent(path)
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)
	places = slices.DeleteFunc(places, func(p string) bool { return p == dir })
	places = append([]string{dir}, places[:min(len(places), recentLimit-1)]...)
	return writeStateFile(path, strings.Join(places, "\x00")+"\x00")
}

func (r recentPlaces) pageStart(rows int) int {
	lastPage := max(0, len(r.paths)-rows)
	start := min(r.scroll, r.cursor, lastPage)
	return max(start, r.cursor-rows+1)
}

func (m *browser) revealRecent() {
	path := m.recent.paths[m.recent.cursor]
	if err := validateDirectory(path); err != nil {
		m.message = err.Error()
		return
	}
	next := *m
	next.recent.revealed, next.recent.focused = path, false
	next.query, next.filtering, next.showHelp = "", false, false
	if err := next.prepareOutline(path); err != nil {
		m.message = err.Error()
		return
	}
	if next.focusedPath() != path || !next.visible[next.cursor].isDir() {
		// A cached parent may predate the directory or a link along its path.
		next.refreshOutline()
		next.focusPath(path)
		if next.focusedPath() != path || !next.visible[next.cursor].isDir() {
			m.message = "Could not reveal the recent directory in the tree"
			return
		}
	}
	*m = next
}

func (m *browser) handleRecent(key keyEvent) bool {
	r := &m.recent
	switch key.name {
	case "esc":
		if m.showHelp {
			m.showHelp = false
			return false
		}
		return true
	case "q":
		return true
	case "?":
		m.showHelp = !m.showHelp
	case "v":
		r.focused = false
		m.cycleView()
		r.focused = m.mode == outlineView
		return false
	case "up", "k", "down", "j", "home", "end", "pgup", "pgdown", "mouse-up", "mouse-down":
		m.move(key.name)
	case "enter", "confirm":
		if len(r.paths) > 0 {
			m.revealRecent()
		}
	}
	return false
}
