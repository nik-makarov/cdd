package main

import (
	"strings"
	"testing"
)

func TestBracketedPasteNeverReplaysCommands(t *testing.T) {
	for _, editing := range []bool{false, true} {
		m := browser{creating: editing}
		var p bracketedPaste
		input := "\x1b[200~q/r\r\n\u754c\x1b[201~"
		// Split at every byte, including the multibyte name and both markers.
		var pending []byte
		var events []keyEvent
		for i := 0; i < len(input); i++ {
			pending = append(pending, input[i])
			for len(pending) > 0 {
				key, n := p.decode(pending, false, &m)
				if n == 0 {
					break
				}
				pending = pending[n:]
				if key.name != "" {
					events = append(events, key)
				}
			}
		}
		if len(pending) != 0 || p.active || len(events) != 1 {
			t.Fatal("paste was not consumed as one event", events)
		}
		want := keyEvent{name: "paste-entry"}
		if editing {
			want = keyEvent{name: "pasted-text", text: "q/r\u754c"}
		}
		if events[0] != want || strings.ContainsAny(events[0].text, "\r\n") {
			t.Fatal("unexpected paste", events[0])
		}
	}
}
