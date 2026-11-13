package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

var sgrPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
var columnPattern = regexp.MustCompile(`\x1b\[([0-9]+)G`)

// Expand trusted horizontal cursor positions so assertions see the screen grid,
// not the byte length of its ANSI protocol. Other control sequences remain visible.
func plainFrame(view string) string {
	lines := strings.Split(sgrPattern.ReplaceAllString(view, ""), "\n")
	for i, line := range lines {
		var out strings.Builder
		start, cells := 0, 0
		for _, match := range columnPattern.FindAllStringSubmatchIndex(line, -1) {
			text := line[start:match[0]]
			out.WriteString(text)
			cells += textWidth(text)
			column, _ := strconv.Atoi(line[match[2]:match[3]])
			padding := max(0, column-1-cells)
			out.WriteString(strings.Repeat(" ", padding))
			cells += padding
			start = match[1]
		}
		out.WriteString(line[start:])
		lines[i] = out.String()
	}
	return strings.Join(lines, "\n")
}

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
	for _, compact := range []bool{false, true} {
		for _, size := range [][2]int{{120, 30}, {80, 24}, {73, 12}, {60, 18}, {56, 8}, {55, 8}, {40, 10}, {20, 8}, {12, 5}, {4, 2}, {1, 1}} {
			for _, mode := range []string{"normal", "filter", "help", "error"} {
				t.Run(fmt.Sprintf("compact=%t/%dx%d/%s", compact, size[0], size[1], mode), func(t *testing.T) {
					m := browser{path: "/long/parent/projects", compact: compact, width: size[0], height: size[1], styles: terminalStyles(true, "xterm", "")}
					for i := range 60 {
						m.visible = append(m.visible, entry{name: fmt.Sprintf("folder-%02d-\u754c", i) + strings.Repeat("long", 20)})
					}
					m.parent, m.children = m.visible, m.visible
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
					plain := plainFrame(view)
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
}

func TestEntryColorsAndMarkers(t *testing.T) {
	m := browser{styles: terminalStyles(true, "xterm", "")}
	for _, tt := range []struct {
		name, marker, color string
		kind                entryKind
	}{
		{"directory", "/", "\x1b[36m", folder},
		{"file", "", "\x1b[33m", file},
		{"directory link", "@/", "\x1b[35m", folderLink},
		{"file link", "@", "\x1b[33m", fileLink},
		{"broken link", "@!", "\x1b[33m", brokenLink},
	} {
		t.Run(tt.name, func(t *testing.T) {
			item := entry{name: "name", kind: tt.kind}
			line := m.itemLine(item, 40, false, false).render(40)
			if !strings.Contains(line, tt.color) || plainFrame(line) != "  name"+tt.marker {
				t.Fatalf("entry is not distinguishable: %q", line)
			}
			m := m
			m.styles = terminalStyles(true, "xterm", "1")
			line = m.itemLine(item, 40, true, false).render(40)
			if strings.Contains(line, "\x1b[3") || !strings.Contains(line, "\x1b[7m") || !strings.Contains(plainFrame(line), "name"+tt.marker) {
				t.Fatalf("NO_COLOR lost the entry marker or added color: %q", line)
			}
		})
	}
}

func TestStyledViewSanitizesDynamicText(t *testing.T) {
	m := browser{
		path: "/tmp/\x1b[31munsafe\n", query: "\x1b[2Jquery", message: "\x1b]2;title\a",
		width: 60, height: 18, styles: terminalStyles(true, "xterm", ""),
		visible: []entry{{name: "."}, {name: "\x1b[7mfolder\r\t"}}, cursor: 1,
	}
	for range 30 {
		m.visible = append(m.visible, entry{name: "\x1b]2;name\a\n\t"})
	}
	m.parent, m.children = m.visible, m.visible
	view := m.view()
	plain := plainFrame(view)
	if strings.ContainsAny(plain, "\x1b\a\r\t") || len(strings.Split(plain, "\n")) != m.height {
		t.Fatalf("terminal controls escaped sanitization: %q", view)
	}
	if !strings.Contains(plain, "Error:") || !strings.Contains(plain, "query") {
		t.Fatalf("error or filter disappeared: %q", plain)
	}
}

func TestTwoColumnListing(t *testing.T) {
	m := browser{path: "/projects", compact: true, width: 80, height: 12}
	m.visible = []entry{{name: "."}}
	for i := 1; i <= 20; i++ {
		m.visible = append(m.visible, entry{name: fmt.Sprintf("dir-%02d", i)})
	}
	rows := m.pageRows()
	if m.columns() != 2 || m.pageSize() != rows*2 {
		t.Fatal("long listing should use both columns")
	}
	first := strings.Split(plainFrame(m.view()), "\n")[3]
	if !strings.Contains(first, "current directory") || !strings.Contains(first, fmt.Sprintf("dir-%02d/", rows)) || !strings.Contains(first, " | ") {
		t.Fatalf("first page is not column-major: %q", first)
	}
	m.cursor = rows
	first = strings.Split(plainFrame(m.view()), "\n")[3]
	if strings.Count(first, ">") != 1 || !strings.Contains(first, fmt.Sprintf("> dir-%02d/", rows)) {
		t.Fatalf("right column selection missing: %q", first)
	}
	m.cursor = rows * 2
	view := plainFrame(m.view())
	if !strings.Contains(view, fmt.Sprintf("> dir-%02d/", m.cursor)) || strings.Contains(view, "dir-01/") || !strings.Contains(view, "more above") {
		t.Fatalf("next page did not advance both columns: %q", view)
	}
	m.width = 55
	if m.columns() != 1 || columnPattern.MatchString(m.view()) {
		t.Fatal("narrow windows should use a single column")
	}
	m.width = 80
	m.cursor = 0
	m.visible = m.visible[:rows]
	if m.columns() != 1 || strings.Contains(m.view(), "switch column") {
		t.Fatal("short listings should use a single column")
	}
}

func TestTwoColumnsAnchorUnicodeAndResetStyles(t *testing.T) {
	m := browser{path: "/projects", compact: true, width: 80, height: 12, styles: terminalStyles(true, "xterm", "")}
	for i := range 30 {
		m.visible = append(m.visible, entry{name: "\u0420\u0443\u0441\u0441\u043a\u0438\u0439-\u754c-" + strconv.Itoa(i)})
	}
	m.cursor = 1
	view := m.view()
	rows := strings.Split(view, "\n")[3 : 3+m.pageRows()]
	wantColumn := -1
	for _, row := range rows {
		positions := columnPattern.FindAllStringSubmatch(row, -1)
		if len(positions) != 1 {
			t.Fatalf("each row must anchor its divider once: %q", row)
		}
		column, _ := strconv.Atoi(positions[0][1])
		if wantColumn < 0 {
			wantColumn = column
		}
		if column != wantColumn {
			t.Fatal("Unicode names shifted the column divider")
		}
		if strings.Contains(row, "\x1b[7m") {
			left := strings.SplitN(row, positions[0][0], 2)[0]
			if !strings.HasSuffix(left, "\x1b[0m") {
				t.Fatal("selection style leaked into the second column")
			}
		}
	}
}

func TestViewStatusAndHelp(t *testing.T) {
	m := browser{path: "/projects/0go", width: 100, height: 24, visible: []entry{{name: "."}, {name: "docs"}}}
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
