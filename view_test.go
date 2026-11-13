package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

var sgrPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestStyledLineClipsBeforeStyling(t *testing.T) {
	line := viewLine{{"a", "\x1b[2m"}, {"\u754c", "\x1b[36m"}, {"/", "\x1b[1m"}}
	for _, tt := range []struct {
		name, want string
		width      int
	}{
		{"wide rune cannot fit", "a", 2},
		{"wide rune fits", "a\u754c", 3},
		{"whole label fits", "a\u754c/", 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := line.render(tt.width)
			plain := sgrPattern.ReplaceAllString(got, "")
			if plain != tt.want || strings.ContainsRune(plain, '\x1b') || !strings.HasSuffix(got, "\x1b[0m") {
				t.Fatalf("clipping corrupted text or styling: %q", got)
			}
		})
	}
}

func TestTerminalStyles(t *testing.T) {
	for _, tt := range []struct {
		name, terminal, noColor      string
		tty, wantColor, wantEmphasis bool
	}{
		{"interactive", "xterm-256color", "", true, true, true},
		{"empty term", "", "", true, true, true},
		{"no color", "xterm", "1", true, false, true},
		{"no color zero", "xterm", "0", true, false, true},
		{"dumb", "dumb", "", true, false, false},
		{"non terminal", "xterm", "", false, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			styles := terminalStyles(tt.tty, tt.terminal, tt.noColor)
			line := viewLine{{"path", styles.accent}, {"error", styles.failure}, {"/", styles.directory}, {"> selected", styles.selected}}.render(80)
			if color := strings.Contains(line, "36m") || strings.Contains(line, "31m"); color != tt.wantColor {
				t.Fatalf("color policy: %q", line)
			}
			if emphasis := strings.Contains(line, "\x1b["); emphasis != tt.wantEmphasis {
				t.Fatalf("emphasis policy: %q", line)
			}
			if sgrPattern.ReplaceAllString(line, "") != "patherror/> selected" {
				t.Fatalf("styling changed content: %q", line)
			}
		})
	}
}

func TestViewFitsTerminalAndRetainsSelection(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 18}, {40, 10}, {20, 8}, {12, 5}, {4, 2}, {1, 1}} {
		for _, mode := range []string{"normal", "filter", "help", "error"} {
			t.Run(fmt.Sprintf("%dx%d/%s", size[0], size[1], mode), func(t *testing.T) {
				m := browser{path: "/long/parent/projects", width: size[0], height: size[1], styles: terminalStyles(true, "xterm", "")}
				for i := range 60 {
					m.visible = append(m.visible, directory{name: fmt.Sprintf("folder-%02d-\u754c", i)})
				}
				m.cursor = 59
				m.filtering = mode == "filter"
				m.showHelp = mode == "help"
				if m.filtering {
					m.query = "folder"
				}
				if mode == "error" {
					m.message = "Permission denied"
				}
				view := m.view()
				plain := sgrPattern.ReplaceAllString(view, "")
				lines := strings.Split(plain, "\n")
				if len(lines) > size[1] || !utf8.ValidString(view) {
					t.Fatalf("invalid frame: %q", view)
				}
				for _, line := range lines {
					if textWidth(line) > max(1, size[0]-1) {
						t.Fatalf("frame wraps: %q", line)
					}
				}
				if !strings.Contains(plain, ">") || !strings.Contains(view, "\x1b[7m") {
					t.Fatalf("selection missing: %q", view)
				}
				if size[0] >= 40 && !strings.Contains(plain, "folder-59-") {
					t.Fatalf("selected directory scrolled out of view: %q", view)
				}
			})
		}
	}
}

func TestStyledViewSanitizesDynamicText(t *testing.T) {
	m := browser{
		path: "/tmp/\x1b[31munsafe\n", query: "\x1b[2Jquery", message: "\x1b]2;title\a",
		width: 60, height: 18, styles: terminalStyles(true, "xterm", ""),
		visible: []directory{{name: "."}, {name: "\x1b[7mfolder\r\t"}}, cursor: 1,
	}
	view := m.view()
	plain := sgrPattern.ReplaceAllString(view, "")
	if strings.ContainsAny(plain, "\x1b\a\r\t") || len(strings.Split(plain, "\n")) != m.height {
		t.Fatalf("terminal controls escaped sanitization: %q", view)
	}
	if !strings.Contains(plain, "Error:") || !strings.Contains(plain, "query") {
		t.Fatalf("error or filter disappeared: %q", plain)
	}
}

func TestViewStatusAndHelp(t *testing.T) {
	m := browser{path: "/projects/0go", width: 100, height: 24, visible: []directory{{name: "."}, {name: "docs"}}}
	plain := m.view()
	if !strings.Contains(plain, "1 directory") || strings.Contains(plain, "1/2") || strings.Contains(plain, "Ctrl+R") {
		t.Fatalf("default view is not compact: %q", plain)
	}
	m.showHelp = true
	view := m.view()
	for _, key := range []string{"Backspace", "Ctrl+R", "Ctrl+D", "Home / End", "close help"} {
		if !strings.Contains(view, key) {
			t.Errorf("help missing %q: %q", key, view)
		}
	}
	m.query = "missing"
	m.visible = m.visible[:1]
	if !strings.Contains(m.view(), "0 matches") || !strings.Contains(m.view(), "current directory") {
		t.Fatal("empty filter must retain current-directory selection")
	}
}

func TestTailKeepsCurrentDirectory(t *testing.T) {
	for _, tt := range []struct {
		name, path, want string
		width            int
	}{
		{"fits", "/tmp/project", "/tmp/project", 20},
		{"deep path", "/very/long/parent/project", ".../project", 11},
		{"wide name", "/parent/\u754c\u754c", "...\u754c\u754c", 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tail(tt.path, tt.width); got != tt.want {
				t.Fatalf("tail = %q, want %q", got, tt.want)
			}
		})
	}
}
