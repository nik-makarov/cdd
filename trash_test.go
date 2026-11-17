package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestTrashRequiresExactTypedName(t *testing.T) {
	root := fixture(t)
	name := " Case \u754c ' $() ; file "
	if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	m := focus(t, loadOutline(t, root), name)
	m = press(m, "trash", "")
	if m.mutation == nil || m.mutation.input != "" || m.action.kind != "" {
		t.Fatal("delete shortcut must show a blank confirmation prompt")
	}
	for _, value := range []string{"", strings.TrimSpace(name), strings.ToLower(name), name + "x", filepath.Join(root, name)} {
		m.mutation.input = value
		m = press(m, "enter", "")
		if m.action.kind != "" || m.mutation.busy || !strings.Contains(m.message, "Name does not match") {
			t.Fatal("a mismatching name must never queue deletion")
		}
	}
	m.mutation.input = ""
	m = press(m, "text", name)
	m = press(m, "enter", "")
	if m.action.kind != "trash" || !m.mutation.busy || m.selected != "" {
		t.Fatal("only the exact name may queue Trash")
	}
	for _, key := range []string{"trash", "enter", "esc", "down", "r", "mouse-left"} {
		before := *m.mutation
		m = press(m, key, "")
		if *m.mutation != before {
			t.Fatal("in-flight operation must freeze modal input")
		}
	}
	if m.action.target == m.mutation {
		t.Fatal("worker must receive an immutable snapshot")
	}
}

func TestTrashMovesEntriesAndRefreshesBothViews(t *testing.T) {
	capture := fakeActionCLI(t, "gio")
	t.Setenv("CDD_TEST_ACTION", "trash")
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, name := range []string{"Alpha", "regular-file", "dir-link", "broken-link"} {
			root, trash := fixture(t), t.TempDir()
			t.Setenv("CDD_TEST_TRASH_DIR", trash)
			for link, target := range map[string]string{"dir-link": "Alpha", "broken-link": "absent"} {
				if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
					t.Fatal(err)
				}
			}
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, name)
			if name == "Alpha" && mode == outlineView {
				m.expandOutline()
			}
			shortcut := "\x1b[127;5u"
			if runtime.GOOS == "darwin" {
				shortcut = "\x15"
			}
			key, _ := decodeKey([]byte(shortcut), false)
			m.handle(key)
			before := m.mutation.info
			m = press(m, "text", name)
			m = press(m, "enter", "")
			action := m.action
			m.action = externalAction{}
			err := runExternalAction(t.Context(), action, "linux")
			m.finishTrash(action, err)
			if err != nil || m.message != "" || m.mutation != nil || m.selected != "" {
				t.Fatalf("%s/%s: trash failed: %v, %s", mode, name, err, m.message)
			}
			moved, err := os.Lstat(filepath.Join(trash, name))
			if err != nil || !os.SameFile(before, moved) {
				t.Fatal("entry must be recoverable in Trash")
			}
			assertEditorArgs(t, capture, []string{"trash", "--", action.path})
			if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
				t.Fatal("trashed entry still appears at its original path")
			}
			for _, item := range m.visible {
				if withinPath(filepath.Join(root, name), item.path) {
					t.Fatal("deleted subtree remained visible")
				}
			}
			if name == "dir-link" {
				if _, err := os.Stat(filepath.Join(root, "Alpha", "nested")); err != nil {
					t.Fatal("trashing a link affected its target")
				}
			}
		}
	}
}

func TestTrashFailureNeverDeletesAndRequiresConfirmationAgain(t *testing.T) {
	fakeActionCLI(t, "gio")
	root := fixture(t)
	for _, behavior := range []string{"fail", "ok", "wait", "missing"} {
		m := focus(t, loadBrowser(t, root), "regular-file")
		m = press(m, "trash", "")
		m = press(m, "text", "regular-file")
		m = press(m, "enter", "")
		action := m.action
		m.action = externalAction{}
		t.Setenv("CDD_TEST_ACTION", behavior)
		if behavior == "missing" {
			t.Setenv("PATH", t.TempDir())
		}
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		err := runExternalAction(ctx, action, "linux")
		cancel()
		m.finishTrash(action, err)
		if err == nil || m.mutation == nil || m.mutation.busy || m.mutation.input != "" || m.message == "" {
			t.Fatalf("%s: failure must require a fresh confirmation: %v", behavior, err)
		}
		data, err := os.ReadFile(filepath.Join(root, "regular-file"))
		if err != nil || string(data) != "text" {
			t.Fatal("Trash failure must leave source data intact")
		}
	}
}

func TestTrashRejectsUnconfirmedOrReplacedTargets(t *testing.T) {
	capture := fakeActionCLI(t, "gio")
	root := fixture(t)
	path := filepath.Join(root, "regular-file")
	if err := runExternalAction(t.Context(), externalAction{kind: "trash", path: path}, "linux"); err == nil {
		t.Fatal("unconfirmed action was accepted")
	}
	m := focus(t, loadBrowser(t, root), "regular-file")
	m = press(m, "trash", "")
	m = press(m, "text", "regular-file")
	m = press(m, "enter", "")
	if err := os.Rename(path, filepath.Join(root, "saved")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runExternalAction(t.Context(), m.action, "linux"); err == nil {
		t.Fatal("replaced target was accepted")
	}
	if _, err := os.Stat(capture); !os.IsNotExist(err) {
		t.Fatal("invalid confirmation must not reach the Trash CLI")
	}
}

func TestTrashRejectsProtectedLocationsAndUnprintableNames(t *testing.T) {
	root := fixture(t)
	for _, m := range []browser{loadBrowser(t, root), focus(t, loadBrowser(t, root), ".."), loadOutline(t, root)} {
		m = press(m, "trash", "")
		if m.mutation != nil || m.action.kind != "" || m.message == "" {
			t.Fatal("protected locations must not open confirmation")
		}
	}
	for _, name := range []string{"line\nbreak", "escape\x1bname"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		m := press(focus(t, loadBrowser(t, root), name), "trash", "")
		if m.mutation != nil || m.action.kind != "" || m.message == "" {
			t.Fatal("unprintable names must not open confirmation")
		}
	}
}

func TestTrashShortcutsAndPromptIsolation(t *testing.T) {
	for _, platform := range []string{"darwin", "linux"} {
		sequence, other := "127;9", "127;5"
		shortcut := "Cmd+Backspace"
		if platform == "linux" {
			sequence, other, shortcut = other, sequence, "Ctrl+Backspace"
		}
		if deleteShortcut(platform) != shortcut {
			t.Fatal("incorrect shortcut label")
		}
		for _, suffix := range []string{"u", ":1u"} {
			if decodeCSIKey(sequence+suffix, platform).name != "trash" || decodeCSIKey(other+suffix, platform).name != "" {
				t.Fatal("incorrect delete modifier")
			}
		}
		for _, suffix := range []string{":2u", ":3u"} {
			if decodeCSIKey(sequence+suffix, platform).name != "" {
				t.Fatal("key repeats/releases must not trigger deletion")
			}
		}
		if decodeCSIKey("127u", platform).name != "backspace" {
			t.Fatal("plain Backspace must retain navigation behavior")
		}
	}
	root := fixture(t)
	for _, input := range []string{"/", "n", "N", "r"} {
		m := focus(t, loadBrowser(t, root), "Alpha")
		m = press(m, input, input)
		m = press(m, "trash", "")
		if m.action.kind != "" || m.mutation != nil && m.mutation.kind == "trash" {
			t.Fatal("delete shortcut must not interrupt another input mode")
		}
	}
	for _, cancel := range []string{"esc", "ctrl+c", "ctrl+d"} {
		m := focus(t, loadBrowser(t, root), "Alpha")
		m = press(m, "trash", "")
		m = press(m, "text", "Alpha")
		if done := m.handle(keyEvent{name: cancel}); done != (cancel != "esc") || m.action.kind != "" {
			t.Fatal("cancel must not queue Trash")
		}
	}
}

func TestTrashLegacyMacShortcut(t *testing.T) {
	key, consumed := decodeKey([]byte("\x15next"), false)
	if consumed != 1 || key.text != "" {
		t.Fatal("legacy shortcut must consume only its control byte")
	}
	if runtime.GOOS != "darwin" {
		if key.name != "" {
			t.Fatal("legacy Ctrl+U must not become a deletion shortcut outside macOS")
		}
		return
	}
	if key.name != "trash" {
		t.Fatal("Ghostty/agterm Cmd+Backspace must open Trash confirmation")
	}
	root := fixture(t)
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, name := range []string{"Alpha", "regular-file"} {
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
			}
			m = focus(t, m, name)
			if m.handle(key) || m.mutation == nil || m.mutation.input != "" || m.action.kind != "" {
				t.Fatal("legacy shortcut must open a blank prompt without deleting or exiting")
			}
			m.handle(key)
			if m.mutation.input != "" || m.action.kind != "" {
				t.Fatal("repeated legacy shortcuts must not confirm deletion")
			}
			m = press(m, "text", name+"wrong")
			m = press(m, "enter", "")
			if m.action.kind != "" || m.mutation.busy {
				t.Fatal("legacy shortcut must retain exact name confirmation")
			}
			m = press(m, "esc", "")
			if _, err := os.Lstat(filepath.Join(root, name)); err != nil {
				t.Fatal("cancelled legacy shortcut affected the entry")
			}
		}
	}
	for _, input := range []string{"/", "n", "N", "r"} {
		m := press(focus(t, loadBrowser(t, root), "Alpha"), input, input)
		m.handle(key)
		if m.action.kind != "" || m.mutation != nil && m.mutation.kind == "trash" {
			t.Fatal("legacy shortcut must not interrupt text input")
		}
	}
	if key := decodeCSIKey("117;5u", "darwin"); key.name != "" {
		t.Fatal("unambiguous Ctrl+U must not be treated as Cmd+Backspace")
	}
}

func TestTrashCommandArgumentsAndButton(t *testing.T) {
	path := "/tmp/-\u754c ' $() `literal`; file"
	program, args, err := trashCommand("darwin", path)
	if err != nil || program != "/usr/bin/osascript" || !slices.Equal(args, []string{"-l", "JavaScript", "-e", macTrashScript, "--", path}) || strings.Contains(macTrashScript, path) {
		t.Fatal("macOS must use a fixed Trash script with one literal path argument")
	}
	root := fixture(t)
	for _, size := range [][2]int{{80, 24}, {40, 20}, {20, 16}} {
		m := focus(t, loadBrowser(t, root), "regular-file")
		m.width, m.height = size[0], size[1]
		found := false
		for y := 0; y < m.height && !found; y++ {
			for x := 0; x < m.width; x++ {
				if m.entryButtonAt(x, y) == "trash" {
					m.handle(keyEvent{name: "mouse-left", x: x, y: y})
					found = true
					break
				}
			}
		}
		if !found || m.mutation == nil || m.mutation.input != "" {
			t.Fatal("visible Delete button must open blank confirmation")
		}
		for _, height := range []int{4, 8, 24} {
			m.height = height
			view := plainFrame(m.view())
			if len(strings.Split(view, "\n")) != height || !strings.Contains(view, "Confirm name:") {
				t.Fatal("confirmation input must remain visible at small sizes")
			}
		}
	}
	if _, _, err := trashCommand("unsupported", path); err == nil {
		t.Fatal("unsupported platforms must fail without deletion")
	}
	if err := validateActionPath(externalAction{kind: "trash"}); err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatal("confirmation must be checked before filesystem access")
	}
}
