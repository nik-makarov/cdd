package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCopySelectedPath(t *testing.T) {
	root := fixture(t)
	for link, target := range map[string]string{"dir-link": "Alpha", "broken-link": "missing"} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, path := range []string{root, filepath.Dir(root), filepath.Join(root, "Alpha"), filepath.Join(root, "regular-file"), filepath.Join(root, "dir-link"), filepath.Join(root, "broken-link")} {
			m := loadBrowser(t, root)
			if mode == outlineView {
				m = loadOutline(t, root)
				m.showRootTree()
			}
			m.focusPath(path)
			before := m
			if m.focusedPath() != path || m.handle(keyEvent{name: "y", text: "y"}) {
				t.Fatal("copy must keep the picker open")
			}
			if m.action != (externalAction{kind: "copy", path: path}) || m.message != "" || m.notice != "" {
				t.Fatal("copy must queue the logical highlighted path without claiming success")
			}
			if m.path != before.path || m.cursor != before.cursor || m.query != before.query || m.selected != "" {
				t.Fatal("copy must preserve browsing and shell selection")
			}
			m = press(m, "down", "")
			m.finishCopyPath(m.action, nil)
			if m.notice != "Copied: "+path || m.message != "" {
				t.Fatal("completion must refer to the original copied path, even after moving focus")
			}
		}
	}
}

func TestClipboardCommandsPreserveLiteralInput(t *testing.T) {
	for _, tt := range []struct {
		platform, program string
		args              []string
	}{
		{"darwin", "pbcopy", nil},
		{"linux", "wl-copy", []string{"--type", "text/plain;charset=utf-8"}},
		{"linux", "xclip", []string{"-selection", "clipboard", "-in"}},
	} {
		t.Run(tt.program, func(t *testing.T) {
			capture := fakeActionCLI(t, tt.program)
			t.Setenv("WAYLAND_DISPLAY", "wayland-test")
			t.Setenv("DISPLAY", ":99")
			t.Setenv("LC_ALL", "C")
			t.Setenv("GORACE", "atexit_sleep_ms=0")
			t.Setenv("CDD_TEST_ACTION", "copy")
			// The path need not exist: copying must not read or resolve the entry.
			path := filepath.Join(t.TempDir(), "-quote' $() `literal`; \u754c\n")
			action := externalAction{kind: "copy", path: path}
			if err := runExternalAction(t.Context(), action, tt.platform); err != nil {
				t.Fatal(err)
			}
			assertEditorArgs(t, capture, tt.args)
			data, err := os.ReadFile(capture + ".stdin")
			if err != nil || string(data) != path {
				t.Fatalf("clipboard input changed: %q, %v", data, err)
			}
			if tt.platform == "darwin" {
				locale, err := os.ReadFile(capture + ".locale")
				if err != nil || string(locale) != "en_US.UTF-8" {
					t.Fatal("pbcopy must interpret Unicode paths as UTF-8")
				}
			}
			if _, _, err := actionCommand(externalAction{kind: "copy", path: "relative"}, tt.platform); err == nil {
				t.Fatal("relative clipboard paths must be rejected")
			}
		})
	}
}

func TestClipboardAvailabilityAndOwnership(t *testing.T) {
	capture := fakeActionCLI(t, "wl-copy")
	if err := os.Symlink(filepath.Join(filepath.Dir(capture), "wl-copy"), filepath.Join(filepath.Dir(capture), "xclip")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WAYLAND_DISPLAY", "wayland-test")
	t.Setenv("DISPLAY", ":99")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	program, _, err := clipboardCommand("linux")
	if err != nil || program != "wl-copy" {
		t.Fatal("Wayland must take priority in a Wayland session")
	}
	action := externalAction{kind: "copy", path: filepath.Join(t.TempDir(), "path")}
	t.Setenv("CDD_TEST_ACTION", "copy-fork")
	if err := runExternalAction(t.Context(), action, "linux"); err != nil {
		t.Fatalf("a successful clipboard owner retaining stderr must not report failure: %v", err)
	}
	t.Setenv("WAYLAND_DISPLAY", "")
	program, _, err = clipboardCommand("linux")
	if err != nil || program != "xclip" {
		t.Fatal("X11 must use xclip")
	}
	t.Setenv("DISPLAY", "")
	if _, _, err := clipboardCommand("linux"); err == nil {
		t.Fatal("a headless session must report clipboard unavailability")
	}
	t.Setenv("WAYLAND_DISPLAY", "wayland-test")
	t.Setenv("PATH", t.TempDir())
	if _, _, err := clipboardCommand("linux"); err == nil || !strings.Contains(err.Error(), "wl-clipboard") {
		t.Fatal("missing clipboard utilities need an actionable error")
	}
}

func TestCopyFailureAndFeedback(t *testing.T) {
	fakeActionCLI(t, "pbcopy")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	m := loadBrowser(t, fixture(t))
	m.copyPath()
	for _, behavior := range []string{"fail", "wait"} {
		t.Setenv("CDD_TEST_ACTION", behavior)
		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		err := runExternalAction(ctx, m.action, "darwin")
		cancel()
		m.finishCopyPath(m.action, err)
		if err == nil || len(err.Error()) > 4300 || m.notice != "" || m.message == "" || m.selected != "" {
			t.Fatal("failed copies must report bounded errors without success or selection")
		}
		if behavior == "wait" && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("copy cancellation must propagate")
		}
	}
	m.message = ""
	m.finishCopyPath(m.action, context.Canceled)
	if m.notice != "" || m.message != "" {
		t.Fatal("canceled copy must not claim success")
	}
	m.styles = terminalStyles(true, "xterm", "")
	m.finishCopyPath(externalAction{kind: "copy", path: "/tmp/\u754c\x1b[31m\nname"}, nil)
	frame := plainFrame(m.view())
	if !strings.Contains(frame, "Copied:") || strings.Contains(frame, "Error:") || strings.ContainsRune(frame, '\x1b') {
		t.Fatal("successful copy must render safe, non-error feedback")
	}
	for _, key := range []string{"mouse-release", "mouse-ignore"} {
		m = press(m, key, "")
		if m.notice == "" {
			t.Fatal("releasing a fast copy click must not immediately dismiss feedback")
		}
	}
	m = press(m, "down", "")
	if m.notice != "" {
		t.Fatal("the next action must dismiss previous feedback")
	}
}

func TestCopyButtonAndInput(t *testing.T) {
	root := fixture(t)
	for _, mode := range []viewMode{outlineView, panesView} {
		for _, name := range []string{"Alpha", "regular-file"} {
			for _, width := range []int{20, 40, 80} {
				m := loadBrowser(t, root)
				if mode == outlineView {
					m = loadOutline(t, root)
				}
				m = focus(t, m, name)
				m.width = width
				m.notice = "Copied: previous path"
				found := false
				for y := 0; y < m.height && !found; y++ {
					for x := 0; x < m.width; x++ {
						if m.entryButtonAt(x, y) == "y" {
							if m.handle(keyEvent{name: "mouse-left", x: x, y: y}) {
								t.Fatal("copy button exited cdd")
							}
							found = true
							break
						}
					}
				}
				if !found || m.action != (externalAction{kind: "copy", path: filepath.Join(root, name)}) || m.selected != "" {
					t.Fatal("copy button must work for files and dirs in both layouts")
				}
			}
		}
	}
	for _, input := range []string{"/", "n", "N", "r", "trash"} {
		m := press(focus(t, loadBrowser(t, root), "Alpha"), input, input)
		m = press(m, "y", "y")
		if m.action.kind != "" || len(m.entryButtons()) != 0 {
			t.Fatal("copy must not interrupt text input")
		}
		if m.mutation != nil && m.mutation.input != "y" || input == "/" && m.query != "y" || (input == "n" || input == "N") && m.newName != "y" {
			t.Fatal("y must remain literal while naming or filtering")
		}
	}
	m := focus(t, loadBrowser(t, root), "regular-file")
	m.height = 2
	if strings.Contains(m.view(), "[ Copy Path") {
		t.Fatal("clipped copy button must be hidden")
	}
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.entryButtonAt(x, y) == "y" {
				t.Fatal("clipped copy button must not be clickable")
			}
		}
	}
	if !slices.ContainsFunc(m.entryActions("linux"), func(h keyHint) bool { return h.key == "y" }) {
		t.Fatal("copy button must be available on Linux too")
	}
}
