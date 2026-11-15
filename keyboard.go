package main

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func changeShortcut(goos string) string {
	if goos == "darwin" {
		return "Cmd+Enter"
	}
	return "Ctrl+Enter"
}

// decodeCSIKey handles the disambiguation subset of the kitty keyboard protocol.
// Lock modifiers do not change shortcuts; key releases never trigger actions.
func decodeCSIKey(sequence, goos string) keyEvent {
	if sequence == "" {
		return keyEvent{}
	}
	final := sequence[len(sequence)-1]
	fields := strings.Split(sequence[:len(sequence)-1], ";")
	if len(fields) > 2 {
		return keyEvent{}
	}
	code := 1
	if fields[0] != "" {
		var ok bool
		code, ok = keyNumber(fields[0])
		if !ok {
			return keyEvent{}
		}
	}
	mods, event := 0, 1
	if len(fields) == 2 {
		modifier, eventText, hasEvent := strings.Cut(fields[1], ":")
		value, ok := keyNumber(modifier)
		if !ok || value < 1 || value > 256 {
			return keyEvent{}
		}
		mods = (value - 1) &^ (64 | 128)
		if hasEvent {
			event, ok = keyNumber(eventText)
			if !ok || event < 1 || event > 2 {
				return keyEvent{}
			}
		}
	}
	if final == 'u' {
		if fields[0] == "" {
			return keyEvent{}
		}
		switch code {
		case 13, 57414:
			confirmMod := 4
			if goos == "darwin" {
				confirmMod = 8
			}
			if mods == confirmMod && event == 1 {
				return keyEvent{name: "confirm"}
			}
			if mods == 0 {
				return keyEvent{name: "enter"}
			}
		case 9:
			if mods == 0 {
				return keyEvent{name: "tab"}
			}
			if mods == 1 {
				return keyEvent{name: "backtab"}
			}
		case 27:
			if mods == 0 {
				return keyEvent{name: "esc"}
			}
		case 127:
			if mods == 0 {
				return keyEvent{name: "backspace"}
			}
		default:
			if mods == 1 && ((code >= 'a' && code <= 'z') || (code >= 'A' && code <= 'Z')) {
				text := string(unicode.ToUpper(rune(code)))
				return keyEvent{name: text, text: text}
			}
			if mods == 4 {
				switch code {
				case 'c':
					return keyEvent{name: "ctrl+c"}
				case 'd':
					return keyEvent{name: "ctrl+d"}
				case 'r':
					return keyEvent{name: "ctrl+r"}
				}
			}
			if mods == 0 && code <= utf8.MaxRune && unicode.IsPrint(rune(code)) {
				text := string(rune(code))
				return keyEvent{name: text, text: text}
			}
		}
		return keyEvent{}
	}
	if final == 'Z' && code == 1 && (mods == 0 || mods == 1) {
		return keyEvent{name: "backtab"}
	}
	if mods != 0 {
		return keyEvent{}
	}
	name := ""
	if final == '~' {
		if fields[0] == "" {
			return keyEvent{}
		}
		switch code {
		case 1, 7:
			name = "home"
		case 4, 8:
			name = "end"
		case 5:
			name = "pgup"
		case 6:
			name = "pgdown"
		}
	} else if code == 1 {
		switch final {
		case 'A':
			name = "up"
		case 'B':
			name = "down"
		case 'C':
			name = "right"
		case 'D':
			name = "left"
		case 'H':
			name = "home"
		case 'F':
			name = "end"
		}
	}
	return keyEvent{name: name}
}

func keyNumber(text string) (int, bool) {
	for _, r := range text {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	value, err := strconv.Atoi(text)
	return value, err == nil
}
