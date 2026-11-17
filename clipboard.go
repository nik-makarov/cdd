package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

func clipboardCommand(platform string) (string, []string, error) {
	switch platform {
	case "darwin":
		return "pbcopy", nil, nil
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if _, err := exec.LookPath("wl-copy"); err == nil {
				return "wl-copy", []string{"--type", "text/plain;charset=utf-8"}, nil
			}
		}
		if os.Getenv("DISPLAY") != "" {
			if _, err := exec.LookPath("xclip"); err == nil {
				return "xclip", []string{"-selection", "clipboard", "-in"}, nil
			}
		}
		return "", nil, errors.New("Clipboard unavailable: use wl-copy (wl-clipboard) in Wayland or xclip in X11")
	default:
		return "", nil, errors.New("Clipboard is unavailable on this platform")
	}
}

func (m *browser) copyPath() {
	if len(m.visible) == 0 {
		return
	}
	m.message, m.notice = "", ""
	m.action = externalAction{kind: "copy", path: m.focusedPath()}
}

func (m *browser) finishCopyPath(action externalAction, err error) {
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			m.message = err.Error()
		}
		return
	}
	m.notice = "Copied: " + action.path
}
