package main

import (
	"path/filepath"
	"testing"
)

func TestExtendedKeyboard(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		for _, tt := range []struct {
			sequence, name, text string
		}{
			{"13u", "enter", ""}, {"13;1u", "enter", ""}, {"57414u", "enter", ""},
			{"27u", "esc", ""}, {"9u", "tab", ""}, {"9;2u", "backtab", ""},
			{"127u", "backspace", ""}, {"99;5u", "ctrl+c", ""},
			{"100;5u", "ctrl+d", ""}, {"114;5u", "ctrl+r", ""},
			{"99;69u", "ctrl+c", ""}, {"114;5:2u", "ctrl+r", ""},
			{"1;65A", "up", ""}, {"1;1:2B", "down", ""}, {"1;2Z", "backtab", ""},
			{"5;1~", "pgup", ""}, {"6;1:3~", "", ""},
			{"97u", "a", "a"}, {"1103u", "\u044f", "\u044f"},
			{"13;9:3u", "", ""}, {"13;5:3u", "", ""},
			{"13;9:2u", "", ""}, {"13;5:2u", "", ""},
			{"13;2u", "", ""}, {"13;3u", "", ""},
			{"13;10u", "", ""}, {"13;6u", "", ""},
			{"13;0u", "", ""}, {"13;+9u", "", ""},
			{"13;9:0u", "", ""}, {"13;9:4u", "", ""},
			{"13;9:1:2u", "", ""}, {"13;9999999999999999999u", "", ""},
			{"13;9;13u", "", ""}, {"?1u", "", ""}, {"u", "", ""},
			{"-1u", "", ""}, {"1114112u", "", ""}, {"55296u", "", ""},
		} {
			t.Run(goos+"/"+tt.sequence, func(t *testing.T) {
				key := decodeCSIKey(tt.sequence, goos)
				if key.name != tt.name || key.text != tt.text {
					t.Fatalf("decoded %+v, want %s/%q", key, tt.name, tt.text)
				}
			})
		}
		confirm, other := "13;5", "13;9"
		shortcut := "Ctrl+Enter"
		if goos == "darwin" {
			confirm, other, shortcut = other, confirm, "Cmd+Enter"
		}
		for _, suffix := range []string{"u", ":1u"} {
			if key := decodeCSIKey(confirm+suffix, goos); key.name != "confirm" {
				t.Fatalf("%s must confirm on %s", confirm+suffix, goos)
			}
			if key := decodeCSIKey(other+suffix, goos); key.name != "" {
				t.Fatalf("wrong platform modifier activated %+v on %s", key, goos)
			}
		}
		if changeShortcut(goos) != shortcut {
			t.Fatal("shortcut label differs from accepted modifiers")
		}
	}
}

func TestEnterBrowsesAndModifiedEnterConfirms(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		{
			root := fixture(t)
			m := focus(t, loadBrowser(t, root), "Alpha")
			m.mode = panesView
			if m.handle(keyEvent{name: "enter"}) || m.path != filepath.Join(root, "Alpha") || m.selected != "" {
				t.Fatal("Enter must only open the highlighted directory")
			}
			current := m
			if !current.handle(keyEvent{name: "enter"}) || current.selected != filepath.Join(root, "Alpha") {
				t.Fatal("Enter on . must confirm the current directory and exit")
			}
			m = focus(t, m, "..")
			if m.handle(keyEvent{name: "enter"}) || m.path != root || m.visible[m.cursor].name != "Alpha" || m.selected != "" {
				t.Fatal("Enter on .. must browse the parent and restore focus")
			}
			sequence := "13;5u"
			if goos == "darwin" {
				sequence = "13;9u"
			}
			if !m.handle(decodeCSIKey(sequence, goos)) || m.selected != filepath.Join(root, "Alpha") {
				t.Fatal("platform modifier plus Enter must confirm the destination")
			}
		}
	}
}

func TestConfirmationDoesNotSubmitTextInput(t *testing.T) {
	for _, mode := range []string{"filter", "create"} {
		m := loadBrowser(t, fixture(t))
		m.filtering, m.creating = mode == "filter", mode == "create"
		m.query, m.newName = "Alpha", "new"
		if m.handle(keyEvent{name: "confirm"}) || m.selected != "" || (!m.filtering && !m.creating) || m.query != "Alpha" || m.newName != "new" {
			t.Fatal("confirmation must not finish text input or change directory")
		}
	}
}
