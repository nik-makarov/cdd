package main

import (
	"errors"
	"os/exec"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

// Pass paths as arguments, never interpolate them into JavaScript. Foundation
// moves the entry to the system Trash without involving Finder automation.
const macTrashScript = `ObjC.import("Foundation");
function run(args) {
    if (args.length !== 1) throw Error("Expected one Trash target");
    var url = $.NSURL.fileURLWithPath(args[0]);
    if (!$.NSFileManager.defaultManager.trashItemAtURLResultingItemURLError(url, null, null)) {
        throw Error("Could not move the entry to Trash; check permissions and volume support");
    }
}`

func trashCommand(platform, path string) (string, []string, error) {
	switch platform {
	case "darwin":
		return "/usr/bin/osascript", []string{"-l", "JavaScript", "-e", macTrashScript, "--", path}, nil
	case "linux":
		program, err := exec.LookPath("gio")
		if err != nil {
			return "", nil, errors.New("Trash is unavailable: install gio (GLib); nothing was deleted")
		}
		return program, []string{"trash", "--", path}, nil
	default:
		return "", nil, errors.New("System Trash is unavailable on this platform")
	}
}

func (m *browser) beginTrash() {
	edit, err := m.mutationTarget()
	if err != nil {
		m.message = err.Error()
		return
	}
	name := filepath.Base(edit.path)
	if !utf8.ValidString(name) {
		m.message = "Cannot confirm a name that is not valid UTF-8"
		return
	}
	for _, r := range name {
		if !unicode.IsPrint(r) {
			m.message = "Cannot confirm a name containing terminal control characters"
			return
		}
	}
	edit.kind = "trash"
	m.mutation, m.message, m.showHelp = edit, "", false
	m.lastClick = mouseClick{}
}

func (m *browser) confirmTrash() {
	edit := m.mutation
	if edit.input != filepath.Base(edit.path) {
		m.message = "Name does not match; nothing was moved to Trash"
		return
	}
	if err := edit.validate(); err != nil {
		m.message = err.Error()
		return
	}
	edit.busy, m.message = true, ""
	snapshot := *edit
	m.action = externalAction{kind: "trash", path: edit.resolved, target: &snapshot}
}

func (m *browser) finishTrash(action externalAction, err error) {
	edit := m.mutation
	m.mutation = nil
	if action.target != nil {
		m.refreshMutation(action.target.path, filepath.Dir(action.target.path))
	}
	if err != nil {
		if edit != nil {
			edit.busy, edit.input = false, ""
			m.mutation = edit
		}
		m.message = err.Error()
	}
}
