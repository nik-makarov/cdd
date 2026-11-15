package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type externalAction struct {
	kind    string
	path    string
	command string
}

func regularFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("This action requires a regular file")
	}
	return nil
}

func actionCommand(action externalAction, platform string) (string, []string, error) {
	if !filepath.IsAbs(action.path) {
		return "", nil, errors.New("File actions require an absolute path")
	}
	if action.kind == "preview" && platform == "darwin" {
		return "qlmanage", []string{"-p", action.path}, nil
	}
	if action.kind == "open" {
		switch platform {
		case "darwin":
			return "open", []string{action.path}, nil
		case "linux":
			return "xdg-open", []string{action.path}, nil
		}
	}
	if action.kind == "editor" {
		return editorCommand(action.command, action.path)
	}
	return "", nil, errors.New("Unsupported file action")
}

// Limit diagnostics and keep child output away from the terminal protocol and
// selected-path stdout. The command is invoked directly, never through a shell.
type actionOutput struct{ text []byte }

func (b *actionOutput) Write(p []byte) (int, error) {
	b.text = append(b.text, p[:min(len(p), 4096-len(b.text))]...)
	return len(p), nil
}

func runExternalAction(ctx context.Context, action externalAction, platform string) error {
	program, args, err := actionCommand(action, platform)
	if err != nil {
		return err
	}
	if program == "" {
		return nil
	}
	if err := validateActionPath(action); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, program, args...)
	var output actionOutput
	cmd.Stderr = &output
	cmd.WaitDelay = 500 * time.Millisecond
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if detail := strings.TrimSpace(string(output.text)); detail != "" {
			return fmt.Errorf("%s: %w: %s", program, err, detail)
		}
		return fmt.Errorf("%s: %w", program, err)
	}
	return nil
}

func validateActionPath(action externalAction) error {
	if action.kind != "editor" {
		return regularFile(action.path)
	}
	info, err := os.Stat(action.path)
	if err != nil {
		return err
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return errors.New("Editor requires a dir or regular file")
	}
	return nil
}

func (m *browser) openInEditor() {
	action := externalAction{kind: "editor", path: m.focusedPath(), command: m.config.EditorCommand}
	program, _, err := editorCommand(action.command, action.path)
	if err != nil {
		m.message = err.Error()
		return
	}
	if program == "" {
		return
	}
	if err := validateActionPath(action); err != nil {
		m.message = err.Error()
		return
	}
	m.message = ""
	m.action = action
}

func (m *browser) openFile() {
	path := m.focusedPath()
	if err := regularFile(path); err != nil {
		m.message = err.Error()
		return
	}
	m.message = ""
	m.action = externalAction{kind: "open", path: path}
}
