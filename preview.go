package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"
)

const previewLimit = 256 * 1024

type filePreview struct {
	path      string
	lines     []string
	scroll    int
	truncated bool
}

func readFilePreview(path string) (*filePreview, error) {
	// Nonblocking open also prevents a replaced path from trapping us in a FIFO.
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("Only regular files can be previewed")
	}
	data, err := io.ReadAll(io.LimitReader(f, previewLimit+1))
	if err != nil {
		return nil, err
	}
	p := &filePreview{path: path, truncated: len(data) > previewLimit}
	data = data[:min(len(data), previewLimit)]
	if p.truncated {
		// A byte limit can cut the last UTF-8 rune; trim its partial encoding.
		for range utf8.UTFMax - 1 {
			r, n := utf8.DecodeLastRune(data)
			if r != utf8.RuneError || n != 1 {
				break
			}
			data = data[:len(data)-1]
		}
	}
	if bytes.ContainsRune(data, 0) || !utf8.Valid(data) {
		p.lines = []string{"Binary file: " + strconv.FormatInt(info.Size(), 10) + " bytes", "No text preview available."}
		return p, nil
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\t", "    ")
	p.lines = strings.Split(text, "\n")
	return p, nil
}

func (m *browser) previewFile(platform string) {
	item := m.visible[m.cursor]
	if item.isDir() {
		return
	}
	if err := regularFile(item.path); err != nil {
		m.message = err.Error()
		return
	}
	m.message, m.showHelp = "", false
	if platform == "darwin" {
		m.action = externalAction{kind: "preview", path: item.path}
		return
	}
	p, err := readFilePreview(item.path)
	if err != nil {
		m.message = err.Error()
		return
	}
	m.filePreview = p
}

func (m *browser) handleFilePreview(key keyEvent) {
	p := m.filePreview
	rows := max(1, m.height-3)
	switch key.name {
	case "esc", " ", "q":
		m.filePreview = nil
		return
	case "up", "k", "mouse-up":
		p.scroll--
	case "down", "j", "mouse-down":
		p.scroll++
	case "pgup":
		p.scroll -= rows
	case "pgdown":
		p.scroll += rows
	case "home":
		p.scroll = 0
	case "end":
		p.scroll = len(p.lines)
	}
	p.scroll = max(0, min(p.scroll, len(p.lines)-rows))
}

func (m browser) filePreviewView() string {
	p := m.filePreview
	width, margin := m.viewWidth()
	rows := max(1, m.height-3)
	start := min(p.scroll, max(0, len(p.lines)-rows))
	lines := []viewLine{{{"Preview: " + tail(display(p.path), max(1, width-9)), m.styles.accent}}, nil}
	for i := start; i < min(start+rows, len(p.lines)); i++ {
		lines = append(lines, viewLine{{p.lines[i], ""}})
	}
	for len(lines) < m.height-1 {
		lines = append(lines, nil)
	}
	footer := "Esc / Space close  |  arrows / PgUp / PgDn scroll"
	if p.truncated {
		footer = "First 256 KiB  |  " + footer
	}
	lines = append(lines, viewLine{{footer, m.styles.muted}})
	if m.height <= 2 {
		lines = []viewLine{{{p.lines[start], ""}}}
	}
	var rendered []string
	for _, line := range lines[:min(len(lines), max(1, m.height))] {
		rendered = append(rendered, strings.Repeat(" ", margin)+line.render(width))
	}
	return strings.Join(rendered, "\n")
}
