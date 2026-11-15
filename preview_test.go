package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestFilePreviewNavigationAndSanitization(t *testing.T) {
	root := fixture(t)
	path := filepath.Join(root, "regular-file")
	if err := os.WriteFile(path, []byte("unsafe\x1b[31m\ttext\r\n"+strings.Repeat("line\n", 60)), 0o600); err != nil {
		t.Fatal(err)
	}
	m := focus(t, loadOutline(t, root), "regular-file")
	m.previewFile("linux")
	if m.filePreview == nil || m.selected != "" {
		t.Fatal("preview must keep selection in the picker")
	}
	if strings.ContainsAny(plainFrame(m.view()), "\x1b\r\t") {
		t.Fatal("file controls reached the terminal")
	}
	if m.handle(keyEvent{name: "pgdown"}) || m.filePreview.scroll == 0 {
		t.Fatal("preview must scroll without browsing")
	}
	for _, size := range [][2]int{{80, 24}, {20, 8}, {4, 2}, {1, 1}} {
		m.width, m.height = size[0], size[1]
		for _, line := range strings.Split(plainFrame(m.view()), "\n") {
			if textWidth(line) > max(1, m.width-1) {
				t.Fatalf("preview exceeds width: %q", line)
			}
		}
	}
	if m.handle(keyEvent{name: "esc"}) || m.filePreview != nil || m.focusedPath() != path || m.selected != "" {
		t.Fatal("closing preview must retain browser focus")
	}
	m.previewFile("darwin")
	if m.action.kind != "preview" || m.action.path != path || m.filePreview != nil {
		t.Fatal("macOS must request Quick Look")
	}
}

func TestFilePreviewLimitsAndSpecialFiles(t *testing.T) {
	root := t.TempDir()
	for _, tt := range []struct {
		name, text        string
		binary, truncated bool
	}{
		{"empty", "", false, false}, {"binary", "a\x00b", true, false},
		{"large", strings.Repeat("x", previewLimit-1) + "\u754c", false, true},
	} {
		path := filepath.Join(root, tt.name)
		if err := os.WriteFile(path, []byte(tt.text), 0o600); err != nil {
			t.Fatal(err)
		}
		p, err := readFilePreview(path)
		if err != nil {
			t.Fatal(err)
		}
		if p.truncated != tt.truncated || strings.HasPrefix(p.lines[0], "Binary file:") != tt.binary {
			t.Fatalf("incorrect preview for %s", tt.name)
		}
	}
	fifo := filepath.Join(root, "fifo")
	if err := unix.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, fifo, filepath.Join(root, "missing")} {
		if _, err := readFilePreview(path); err == nil {
			t.Fatalf("special path accepted: %q", path)
		}
	}
}

func TestSpaceRemainsLiteralDuringInput(t *testing.T) {
	for _, key := range []string{"/", "n"} {
		m := loadOutline(t, fixture(t))
		m = press(m, key, key)
		m = press(m, " ", " ")
		if m.filePreview != nil || m.action.kind != "" || (m.query != " " && m.newName != " ") {
			t.Fatal("space must remain literal while editing")
		}
	}
}
