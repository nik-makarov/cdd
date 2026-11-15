package main

import (
	"errors"
	"os/exec"
	"strings"
)

func (m browser) selectedEditorCommand() string {
	if len(m.visible) == 0 {
		return ""
	}
	if m.visible[m.cursor].isDir() {
		return strings.TrimSpace(m.config.EditorDirCommand)
	}
	return strings.TrimSpace(m.config.EditorFileCommand)
}

// parseEditorCommand accepts quoted arguments and || alternatives, without shell
// expansion. Alternatives are used only when an executable is unavailable.
func parseEditorCommand(command string) ([][]string, error) {
	if strings.ContainsAny(command, "\x00\r\n") {
		return nil, errors.New("command must be a single line without NUL")
	}
	if strings.TrimSpace(command) == "" {
		return nil, nil
	}
	var commands [][]string
	var args []string
	var word strings.Builder
	var quote rune
	started := false
	flush := func() {
		if started {
			args = append(args, word.String())
			word.Reset()
			started = false
		}
	}
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case r == '\\' && quote != '\'':
			i++
			if i == len(runes) {
				return nil, errors.New("unfinished escape")
			}
			word.WriteRune(runes[i])
			started = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				word.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == ' ' || r == '\t':
			flush()
		case r == '|' && i+1 < len(runes) && runes[i+1] == '|':
			flush()
			if len(args) == 0 || args[0] == "" {
				return nil, errors.New("empty command alternative")
			}
			commands, args = append(commands, args), nil
			i++
		case strings.ContainsRune("|;&<>()$`", r):
			return nil, errors.New("only arguments, quotes, escapes, and || alternatives are supported")
		default:
			word.WriteRune(r)
			started = true
		}
	}
	if quote != 0 {
		return nil, errors.New("unclosed quote")
	}
	flush()
	if len(args) == 0 || args[0] == "" {
		return nil, errors.New("empty command alternative")
	}
	return append(commands, args), nil
}

func editorCommand(command, path string) (string, []string, error) {
	commands, err := parseEditorCommand(command)
	if err != nil {
		return "", nil, err
	}
	for _, args := range commands {
		program, err := exec.LookPath(args[0])
		if err != nil {
			continue
		}
		return program, append(args[1:], "--", path), nil
	}
	return "", nil, nil
}
