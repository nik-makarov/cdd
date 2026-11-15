package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDecodeMouseReports(t *testing.T) {
	for _, tt := range []struct {
		name, input, event string
		x, y               int
	}{
		{"left", "\x1b[<0;23;5M", "mouse-left", 22, 4},
		{"right", "\x1b[<2;23;5M", "mouse-right", 22, 4},
		{"release", "\x1b[<0;23;5m", "mouse-release", 22, 4},
		{"wheel up", "\x1b[<64;23;5M", "mouse-up", 22, 4},
		{"wheel down", "\x1b[<65;23;5M", "mouse-down", 22, 4},
		{"large coordinates", "\x1b[<0;500;300M", "mouse-left", 499, 299},
		{"middle ignored", "\x1b[<1;1;1M", "mouse-ignore", 0, 0},
		{"modifier ignored", "\x1b[<4;1;1M", "mouse-ignore", 0, 0},
		{"drag ignored", "\x1b[<32;1;1M", "mouse-ignore", 0, 0},
		{"horizontal wheel ignored", "\x1b[<66;1;1M", "mouse-ignore", 0, 0},
		{"legacy left", "\x1b[M !!", "mouse-left", 0, 0},
		{"legacy right", "\x1b[M\"%&", "mouse-right", 4, 5},
		{"legacy release", "\x1b[M#!!", "mouse-release", 0, 0},
		{"legacy wheel", "\x1b[M`!!", "mouse-up", 0, 0},
		{"legacy shortcut bytes", "\x1b[Mnqv", "mouse-ignore", 80, 85},
		{"legacy invalid coordinates", "\x1b[M   ", "", 0, 0},
		{"zero coordinate", "\x1b[<0;0;1M", "", 0, 0},
		{"negative coordinate", "\x1b[<0;-1;1M", "", 0, 0},
		{"missing coordinate", "\x1b[<0;1M", "", 0, 0},
		{"empty coordinate", "\x1b[<0;;1M", "", 0, 0},
		{"extra coordinate", "\x1b[<0;1;1;1M", "", 0, 0},
		{"overflow", "\x1b[<0;999999999999999999999999;1M", "", 0, 0},
		{"wrong terminator", "\x1b[<0;1;1Z", "", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.input)
			for n := 1; n < len(data); n++ {
				if _, consumed := decodeKey(data[:n], false); consumed != 0 {
					t.Fatalf("consumed incomplete mouse report at byte %d", n)
				}
			}
			key, consumed := decodeKey(append(data, 'q'), false)
			if consumed != len(data) || key.name != tt.event || key.text != "" || key.x != tt.x || key.y != tt.y {
				t.Fatalf("decoded %+v (%d bytes), want %s at (%d,%d)", key, consumed, tt.event, tt.x, tt.y)
			}
		})
	}
}

func TestMouseEntryMatchesRenderedRows(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, size := range [][2]int{{120, 24}, {80, 12}, {58, 8}, {40, 7}, {20, 5}} {
			for _, mode := range []string{"normal", "filter", "help", "error"} {
				t.Run(fmt.Sprintf("compact=%t/%dx%d/%s", compact, size[0], size[1], mode), func(t *testing.T) {
					m := browser{path: "/projects", compact: compact, width: size[0], height: size[1], cursor: 37}
					for i := range 60 {
						name := fmt.Sprintf("dir-%02d", i)
						m.visible = append(m.visible, entry{name: name, path: filepath.Join(m.path, name)})
					}
					m.filtering, m.showHelp = mode == "filter", mode == "help"
					if mode == "error" {
						m.message = "Example error"
					}
					matched := 0
					for y, line := range strings.Split(plainFrame(m.view()), "\n") {
						for i, item := range m.visible {
							if x := strings.Index(line, item.label()); x >= 0 {
								matched++
								if got := m.mouseEntry(x, y); got != i {
									t.Fatalf("click on displayed %s at (%d,%d) maps to entry %d", item.name, x, y, got)
								}
							}
						}
					}
					if matched == 0 {
						t.Fatal("test did not find any displayed entries")
					}
				})
			}
		}
	}
}

func TestMouseIgnoresNonEntryAreas(t *testing.T) {
	m, err := newBrowser(fixture(t), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, cell := range [][2]int{{-1, 4}, {22, -1}, {22, 0}, {22, 1}, {22, 2}, {1, 4}, {3, 4}, {19, 4}, {20, 4}, {21, 4}, {58, 4}, {60, 4}, {77, 4}, {79, 4}, {22, 15}, {22, 21}, {22, 30}} {
		if got := m.mouseEntry(cell[0], cell[1]); got != -1 {
			t.Fatalf("non-entry cell %v maps to %d", cell, got)
		}
	}
	for _, compact := range []bool{false, true} {
		m.compact, m.width, m.height = compact, 1, 1
		m.cursor = 2
		if got := m.mouseEntry(0, 0); got != m.cursor {
			t.Fatalf("single-cell terminal maps to %d, want %d", got, m.cursor)
		}
	}
}

func mouseOn(t *testing.T, m browser, label, button string, when time.Time) keyEvent {
	t.Helper()
	for y, line := range strings.Split(plainFrame(m.view()), "\n") {
		if x := strings.Index(line, label); x >= 0 && m.mouseEntry(x, y) >= 0 {
			return keyEvent{name: button, x: textWidth(line[:x]), y: y, when: when}
		}
	}
	t.Fatalf("entry %q is not displayed", label)
	return keyEvent{}
}

func TestMouseHighlightsAndOpensDirectory(t *testing.T) {
	for _, mode := range []string{"tree", "list"} {
		t.Run(mode, func(t *testing.T) {
			root := fixture(t)
			m := loadBrowser(t, root)
			m.compact = mode == "list"
			m.refilter()
			click := mouseOn(t, m, "Alpha/", "mouse-left", time.Unix(100, 0))
			if m.handle(click) || m.selected != "" || m.path != root || m.visible[m.cursor].name != "Alpha" {
				t.Fatal("single click must only highlight the directory")
			}
			if !m.compact && m.preview != filepath.Join(root, "Alpha") {
				t.Fatal("click did not refresh the directory preview")
			}
			selected := m
			if !selected.handle(keyEvent{name: "confirm"}) || selected.selected != filepath.Join(root, "Alpha") {
				t.Fatal("confirmation must choose the highlighted directory")
			}
			m.handle(keyEvent{name: "mouse-release", x: click.x, y: click.y})
			click.when = click.when.Add(200 * time.Millisecond)
			click.x += 2
			if m.handle(click) || m.selected != "" || m.path != filepath.Join(root, "Alpha") {
				t.Fatal("double click must browse the highlighted directory without selecting it")
			}
		})
	}
}

func TestMouseKeepsScrolledRowsStable(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := browser{path: "/projects", compact: compact, width: 80, height: 16, cursor: 37}
		for i := range 60 {
			m.visible = append(m.visible, entry{name: fmt.Sprintf("dir-%02d", i), path: fmt.Sprintf("/projects/dir-%02d", i)})
		}
		start := m.pageStart()
		click := mouseOn(t, m, m.visible[start+1].label(), "mouse-left", time.Unix(100, 0))
		m.handle(click)
		if m.pageStart() != start || m.mouseEntry(click.x, click.y) != m.cursor || m.cursor != start+1 {
			t.Fatal("click shifted the list underneath the pointer")
		}
	}
}

func TestMouseOpensCurrentAndSecondColumn(t *testing.T) {
	root := t.TempDir()
	for i := range 40 {
		if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("dir-%02d", i)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range []string{".", "dir-20", "dir-39"} {
		t.Run(target, func(t *testing.T) {
			m := loadBrowser(t, root)
			m.compact = true
			if target == "dir-39" {
				m.cursor = len(m.visible) - 1
			}
			width, margin := m.viewWidth()
			panes := m.paneWidths(width, m.columns())
			header, _ := m.frame()
			if target == "dir-39" && m.mouseEntry(margin+panes[0]+3, len(header)+m.pageRows()-1) != -1 {
				t.Fatal("empty second column must not select a directory")
			}
			var click keyEvent
			if target == "." {
				click = mouseOn(t, m, ".  (current dir)", "mouse-left", time.Unix(100, 0))
			} else {
				click = mouseOn(t, m, target+"/", "mouse-left", time.Unix(100, 0))
			}
			if target == "dir-20" && click.x < margin+panes[0]+3 {
				t.Fatal("expected a directory in the second column")
			}
			if m.handle(click) {
				t.Fatal("single click selected a directory")
			}
			click.when = click.when.Add(200 * time.Millisecond)
			if m.handle(click) || m.selected != "" || m.path != filepath.Join(root, target) {
				t.Fatalf("double click opened %q, want %q without selecting", m.path, filepath.Join(root, target))
			}
			if !m.handle(keyEvent{name: "confirm"}) || m.selected != filepath.Join(root, target) {
				t.Fatal("confirmation must choose the opened directory")
			}
		})
	}
}

func TestMouseDoubleClickDoesNotOpenFilesOrStaleDirectories(t *testing.T) {
	for _, mode := range []string{"file", "removed directory", "replaced directory"} {
		t.Run(mode, func(t *testing.T) {
			root := fixture(t)
			m, err := newBrowser(root, false)
			if err != nil {
				t.Fatal(err)
			}
			label := "beta/"
			if mode == "file" {
				label = "regular-file"
			}
			click := mouseOn(t, m, label, "mouse-left", time.Unix(100, 0))
			m.handle(click)
			if mode != "file" {
				if err := os.Remove(filepath.Join(root, "beta")); err != nil {
					t.Fatal(err)
				}
				if mode == "replaced directory" {
					if err := os.WriteFile(filepath.Join(root, "beta"), []byte("unchanged"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			click.when = click.when.Add(200 * time.Millisecond)
			if m.handle(click) || m.selected != "" || m.message == "" || m.path != root {
				t.Fatal("double click bypassed directory navigation validation")
			}
		})
	}
}

func TestMouseDoubleClickResets(t *testing.T) {
	for _, action := range []string{"keyboard", "blank click", "wheel", "modifier", "timeout", "large pointer drift", "different entry"} {
		t.Run(action, func(t *testing.T) {
			root := fixture(t)
			m := loadBrowser(t, root)
			click := mouseOn(t, m, "Alpha/", "mouse-left", time.Unix(100, 0))
			m.handle(click)
			click.when = click.when.Add(200 * time.Millisecond)
			switch action {
			case "keyboard":
				m.handle(keyEvent{name: "home"})
			case "blank click":
				m.handle(keyEvent{name: "mouse-left", x: 0, y: 0})
			case "wheel":
				m.handle(keyEvent{name: "mouse-down", x: click.x, y: click.y})
			case "modifier":
				m.handle(keyEvent{name: "mouse-ignore", x: click.x, y: click.y})
			case "timeout":
				click.when = click.when.Add(time.Second)
			case "large pointer drift":
				click.x += 3
			case "different entry":
				click = mouseOn(t, m, "beta/", "mouse-left", click.when)
			}
			if m.handle(click) || m.selected != "" || m.path != root {
				t.Fatal("unrelated clicks were treated as a double click")
			}
		})
	}
}

func TestMouseBrowsingAndInputModes(t *testing.T) {
	root := fixture(t)
	m := loadBrowser(t, root)
	click := mouseOn(t, m, "Alpha/", "mouse-right", time.Unix(100, 0))
	m.handle(click)
	if m.path != root || m.selected != "" || m.visible[m.cursor].name != "." {
		t.Fatal("right click must not open, highlight, or select a directory")
	}
	m = loadBrowser(t, root)
	click = mouseOn(t, m, "Alpha/", "mouse-down", time.Unix(100, 0))
	m.handle(click)
	if m.visible[m.cursor].name != "Alpha" {
		t.Fatal("wheel down must move one entry")
	}
	click.name = "mouse-up"
	m.handle(click)
	if m.visible[m.cursor].name != "." {
		t.Fatal("wheel up must move one entry")
	}
	m = press(m, "/", "/")
	m = press(m, "text", "Alpha")
	click = mouseOn(t, m, "Alpha/", "mouse-left", time.Unix(100, 0))
	m.handle(click)
	click.when = click.when.Add(200 * time.Millisecond)
	if m.handle(click) || m.filtering || m.selected != "" || m.query != "" || m.path != filepath.Join(root, "Alpha") {
		t.Fatal("double click while filtering must open the directory and clear the filter")
	}
	m = press(m, "n", "n")
	m = press(m, "text", "new directory")
	for _, name := range []string{"mouse-left", "mouse-right", "mouse-up", "mouse-down", "mouse-release"} {
		click.name = name
		if m.handle(click) || !m.creating || m.newName != "new directory" || m.path != filepath.Join(root, "Alpha") || m.selected != "" {
			t.Fatal("mouse input interrupted directory-name editing")
		}
	}
}

func TestMouseOpensParent(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, name := range []string{"Alpha", "link"} {
			for _, query := range []string{"", "missing"} {
				root := fixture(t)
				if err := os.Symlink(t.TempDir(), filepath.Join(root, "link")); err != nil {
					t.Fatal(err)
				}
				start := filepath.Join(root, name)
				m := loadBrowser(t, start)
				m.compact, m.query, m.filtering = compact, query, query != ""
				m.refilter()
				click := mouseOn(t, m, "\u2190 Back to parent", "mouse-left", time.Unix(100, 0))
				if m.handle(click) || m.path != start || m.selected != "" || m.visible[m.cursor].name != ".." {
					t.Fatal("single click must only highlight ..")
				}
				click.when = click.when.Add(200 * time.Millisecond)
				if m.handle(click) || m.path != root || m.selected != "" || m.visible[m.cursor].name != name || m.query != "" || m.filtering {
					t.Fatal("double click must open the logical parent and focus the previous directory")
				}
			}
		}
	}
}

func TestSelectButtonMatchesRenderedPosition(t *testing.T) {
	root := fixture(t)
	for _, compact := range []bool{false, true} {
		for _, size := range [][2]int{{120, 24}, {80, 12}, {40, 7}, {20, 5}, {12, 5}, {20, 2}, {4, 2}} {
			for _, mode := range []string{"normal", "filter", "completed filter", "help", "error", "create"} {
				t.Run(fmt.Sprintf("compact=%t/%dx%d/%s", compact, size[0], size[1], mode), func(t *testing.T) {
					m := loadBrowser(t, root)
					m.compact, m.width, m.height = compact, size[0], size[1]
					m.filtering, m.creating, m.showHelp = mode == "filter", mode == "create", mode == "help"
					if mode == "completed filter" {
						m.query = "missing"
						m.refilter()
					}
					if mode == "error" {
						m.message = "Example error"
					}
					button := m.selectButton()
					buttonX, buttonY, buttonWidth := -1, -1, 0
					if len(button) > 0 {
						for y, line := range strings.Split(plainFrame(m.view()), "\n") {
							if x := strings.Index(line, button[0][1].text); x >= 0 {
								buttonX, buttonY, buttonWidth = textWidth(line[:x]), y, textWidth(button[0][1].text)
							}
						}
					}
					for y := range m.height {
						for x := range m.width {
							want := buttonY >= 0 && y >= buttonY && y < buttonY+len(button) && x >= buttonX && x < buttonX+buttonWidth
							if m.selectButtonAt(x, y) != want {
								t.Fatalf("button hit test differs from display at (%d,%d)", x, y)
							}
						}
					}
					if buttonY >= 0 {
						for _, name := range []string{"mouse-right", "mouse-release", "mouse-down"} {
							if m.handle(keyEvent{name: name, x: buttonX, y: buttonY}) || m.selected != "" || m.path != root {
								t.Fatal("only a left click may activate the selection button")
							}
						}
						for row := range len(button) {
							copy := m
							if !copy.handle(keyEvent{name: "mouse-left", x: buttonX, y: buttonY + row}) || copy.selected != root {
								t.Fatal("each row of the button must choose the highlighted directory")
							}
						}
					} else if mode == "normal" && m.width >= 40 && m.height >= 7 {
						t.Fatal("selection button should be visible")
					}
				})
			}
		}
	}
}

func TestSelectButtonUsesDirectoryGuards(t *testing.T) {
	for _, target := range []string{"Alpha", "..", "regular-file", "removed"} {
		root := fixture(t)
		m, err := newBrowser(root, false)
		if err != nil {
			t.Fatal(err)
		}
		name := target
		if target == "removed" {
			name = "beta"
			if err := os.Remove(filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
		}
		m = focus(t, m, name)
		button := m.selectButton()[0][1].text
		found := false
		for y, line := range strings.Split(plainFrame(m.view()), "\n") {
			if x := strings.Index(line, button); x >= 0 {
				found = true
				key := keyEvent{name: "mouse-left", x: textWidth(line[:x]), y: y}
				if target == "removed" || target == "regular-file" {
					if m.handle(key) || m.selected != "" || m.message == "" || m.path != root {
						t.Fatal("button must reject files and removed directories")
					}
				} else if !m.handle(key) || m.selected != filepath.Join(root, target) {
					t.Fatal("button must select the highlighted directory, like the confirmation shortcut")
				}
				break
			}
		}
		if !found {
			t.Fatal("selection button is missing")
		}
	}
}
