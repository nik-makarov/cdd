package main

import (
	"testing"
	"unicode/utf8"
)

func TestDecodeKeys(t *testing.T) {
	for _, tt := range []struct {
		input, name, text string
	}{
		{"\x1b[A", "up", ""}, {"\x1bOB", "down", ""},
		{"\x1b[C", "right", ""}, {"\x1b[D", "left", ""},
		{"\x1b[H", "home", ""}, {"\x1b[4~", "end", ""},
		{"\x1b[5~", "pgup", ""}, {"\x1b[6~", "pgdown", ""},
		{"\r", "enter", ""}, {"\x03", "ctrl+c", ""},
		{"\x04", "ctrl+d", ""}, {"\x12", "ctrl+r", ""},
		{"\x7f", "backspace", ""}, {"\u044f", "\u044f", "\u044f"},
		{"\t", "tab", ""}, {"\x1b[Z", "backtab", ""},
		{"\u754c", "\u754c", "\u754c"}, {"/", "/", "/"},
		{"\x1b[999~", "", ""}, {"\x00", "", ""},
	} {
		t.Run(tt.input, func(t *testing.T) {
			data := []byte(tt.input)
			for n := 1; n < len(data); n++ {
				if _, consumed := decodeKey(data[:n], false); consumed != 0 {
					t.Fatalf("consumed fragmented key %q at byte %d", tt.input, n)
				}
			}
			key, consumed := decodeKey(data, false)
			if consumed != len(data) || key.name != tt.name || key.text != tt.text {
				t.Fatalf("decoded %q: %+v, consumed %d", tt.input, key, consumed)
			}
		})
	}
}

func TestEscapeAndMalformedInput(t *testing.T) {
	if _, consumed := decodeKey([]byte("\x1b"), false); consumed != 0 {
		t.Fatal("escape must wait briefly for a possible arrow sequence")
	}
	if key, consumed := decodeKey([]byte("\x1b"), true); key.name != "esc" || consumed != 1 {
		t.Fatal("standalone escape did not cancel")
	}
	for _, data := range [][]byte{[]byte("\x1b["), {0xd0}, {0xff}, []byte("\x1b\u044f")} {
		if _, consumed := decodeKey(data, true); consumed < 1 || consumed > len(data) {
			t.Fatalf("expired input makes no progress: %q", data)
		}
	}
}

func TestTruncateUnicode(t *testing.T) {
	for _, tt := range []struct {
		input string
		width int
		want  string
	}{
		{"abcdef", 3, "abc"}, {"\u0420\u0443\u0441\u0441\u043a\u0438\u0439", 5, "\u0420\u0443"},
		{"a\u754cb", 3, "a\u754c"}, {"x\U0001f642y", 3, "x\U0001f642"},
		{"anything", 0, ""},
	} {
		if got := truncate(tt.input, tt.width); got != tt.want || !utf8.ValidString(got) {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

func FuzzDecodeKey(f *testing.F) {
	f.Add([]byte("\x1b[A"), false)
	f.Add([]byte("\u044f"), true)
	f.Add([]byte("\xff\x1b["), true)
	f.Fuzz(func(t *testing.T, data []byte, expired bool) {
		for len(data) > 0 {
			_, consumed := decodeKey(data, expired)
			if consumed < 0 || consumed > len(data) {
				t.Fatalf("invalid consumption %d for %d bytes", consumed, len(data))
			}
			if consumed == 0 {
				if expired {
					t.Fatal("expired input did not make progress")
				}
				return
			}
			data = data[consumed:]
		}
	})
}
