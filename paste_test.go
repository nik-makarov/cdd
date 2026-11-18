package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestPasteRequiresReviewInBothViews(t *testing.T) {
	for _, mode := range []viewMode{outlineView, panesView} {
		t.Run(mode.String(), func(t *testing.T) {
			root := fixture(t)
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, "Alpha")
			dir := m.focusedPath()
			m.clipboardEntry("paste-entry")
			if m.action.kind != "prepare-paste" || m.action.path != dir {
				t.Fatal("paste must target the highlighted folder")
			}
			read := m.action
			m.action = externalAction{}
			m = press(m, "enter", "")
			if m.action.kind != "" || m.selected != "" {
				t.Fatal("Enter acted before the clipboard was read")
			}
			p, err := prepareEntryCopy(filepath.Join(root, "regular-file"), dir)
			if err != nil {
				t.Fatal(err)
			}
			m.finishAction(read, actionResult{entry: p})
			m.width, m.height = 32, 20
			view := strings.Join(strings.Fields(m.view()), "")
			if !strings.Contains(view, strconv.Quote(p.destination)) || !strings.Contains(view, "Enterpaste") {
				t.Fatal("confirmation must show the complete wrapped destination", m.view())
			}
			m = press(m, "esc", "")
			if m.action.kind != "" || m.paste != nil || m.focusedPath() != dir {
				t.Fatal("escape must cancel and keep focus")
			}
			if _, err := os.Lstat(p.destination); !os.IsNotExist(err) {
				t.Fatal("paste wrote before confirmation")
			}
			m.clipboardEntry("paste-entry")
			m.finishAction(m.action, actionResult{entry: p})
			m.action = externalAction{}
			m.height = 2
			m = press(m, "enter", "")
			if m.action.kind != "" {
				t.Fatal("accepted a destination that cannot be displayed")
			}
			m.height = 3
			pathRows := m.pastePathRows()
			for range pathRows {
				m = press(m, "down", "")
			}
			if m.paste.scroll != len(pathRows)-1 || !strings.Contains(m.view(), pathRows[len(pathRows)-1][0].text) {
				t.Fatal("long destination could not be reviewed by scrolling")
			}
			m.height = 20
			m = press(m, "enter", "")
			a := m.action
			if a.kind != "paste-entry" || a.paste != p {
				t.Fatal("confirmation did not queue the displayed copy")
			}
			m.action = externalAction{}
			m = press(m, "enter", "")
			if m.action.kind != "" {
				t.Fatal("confirmed twice while busy")
			}
			err = runExternalAction(t.Context(), a, "darwin")
			m.finishAction(a, actionResult{err: err})
			if err != nil || m.paste != nil || m.notice != "Pasted: "+p.destination || m.focusedPath() != dir || m.selected != "" {
				t.Fatal("paste did not finish in place", err)
			}
			m = focus(t, m, "regular-file")
			m.clipboardEntry("paste-entry")
			if m.action.path != root {
				t.Fatal("pasting on a file must use its parent")
			}
		})
	}
}

func TestFileClipboardShortcuts(t *testing.T) {
	for _, tt := range []struct{ seq, key string }{
		{"99;9u", "copy-entry"}, {"118;9u", "paste-entry"}, {"118;73:1u", "paste-entry"},
		{"99;9:2u", ""}, {"118;9:3u", ""}, {"99;10u", ""}, {"118;5u", ""},
	} {
		key := decodeCSIKey(tt.seq, "darwin")
		if key.name != tt.key || key.text != "" || decodeCSIKey(tt.seq, "linux").name != "" {
			t.Fatalf("unexpected shortcut for %q: %+v", tt.seq, key)
		}
	}
	m := focus(t, loadOutline(t, fixture(t)), "regular-file")
	if !m.handle(decodeCSIKey("99;5u", "darwin")) || m.action.kind != "" {
		t.Fatal("Ctrl+C must still cancel")
	}
	if runtime.GOOS != "darwin" {
		return
	}
	for _, shortcut := range []string{"99;9u", "118;9u"} {
		key := decodeCSIKey(shortcut, "darwin")
		clicked := m
		clicked.width, clicked.height = 42, 24
		found := false
		for y := 0; y < clicked.height && !found; y++ {
			for x := 0; x < clicked.width; x++ {
				if clicked.entryButtonAt(x, y) == key.name {
					clicked.handle(keyEvent{name: "mouse-left", x: x, y: y})
					found = true
					break
				}
			}
		}
		keyed := m
		keyed.handle(key)
		if !found || keyed.action.kind == "" || keyed.action != clicked.action {
			t.Fatalf("keyboard and button disagree for %q", key.name)
		}
		for _, field := range []string{"/", "n", "N", "r", "trash"} {
			modal := press(m, field, field)
			modal.handle(key)
			if modal.action.kind != "" || modal.paste != nil {
				t.Fatalf("clipboard shortcut interrupted %q", field)
			}
		}
	}
}
