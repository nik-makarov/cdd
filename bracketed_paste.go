package main

import (
	"bytes"
	"strings"
	"unicode"
)

// A terminal can intercept Cmd+V and send a bracketed paste instead of the
// keystroke. Never replay its payload as browsing shortcuts or confirmation.
type bracketedPaste struct {
	active, editing bool
	text            []byte
	tooLarge        bool
}

func (p *bracketedPaste) decode(data []byte, expired bool, m *browser) (keyEvent, int) {
	const start, end = "\x1b[200~", "\x1b[201~"
	if p == nil {
		return decodeKey(data, expired)
	}
	if !p.active {
		if !bytes.HasPrefix(data, []byte(start)) {
			return decodeKey(data, expired)
		}
		p.active = true
		p.editing = m.filtering || m.creating || m.mutation != nil
		return keyEvent{}, len(start)
	}
	n := bytes.Index(data, []byte(end))
	complete := n >= 0
	if !complete {
		n = len(data)
		// Leave a partial closing marker in the terminal's pending input.
		for size := min(len(end)-1, n); size > 0; size-- {
			if bytes.HasSuffix(data, []byte(end[:size])) {
				n -= size
				break
			}
		}
	}
	if p.editing {
		const limit = 64 * 1024
		if n > limit-len(p.text) {
			p.tooLarge = true
		} else if !p.tooLarge {
			p.text = append(p.text, data[:n]...)
		}
	}
	if !complete {
		return keyEvent{}, n
	}
	key := keyEvent{name: "paste-entry"}
	if p.tooLarge {
		key = keyEvent{}
	} else if p.editing {
		key = keyEvent{name: "pasted-text", text: strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, string(p.text))}
	}
	*p = bracketedPaste{}
	return key, n + len(end)
}
