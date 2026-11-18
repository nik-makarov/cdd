package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNativeFileClipboard(t *testing.T) {
	if os.Getenv("CDD_TEST_NATIVE_FILE_CLIPBOARD") != "1" {
		t.Skip("requires a macOS pasteboard server")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "-a' \u754c $() % #\n")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "directory link")
	broken := filepath.Join(dir, "broken link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", broken); err != nil {
		t.Fatal(err)
	}
	paths := []string{file, dir, link, broken}
	// Exercise the production read and write functions using a private board.
	// The user's general pasteboard is never opened.
	script, _, _ := strings.Cut(fileClipboardScript, "function run(args)")
	script += `function run(args) {
	    var board = $.NSPasteboard.pasteboardWithUniqueName;
	    function mustReject() {
	        try { fileClipboard(board, ["read"]); } catch (e) { return; }
	        throw Error("Accepted an invalid pasteboard item");
	    }
	    try {
	        board.clearContents;
	        mustReject();
	        board.setStringForType(args[0], $.NSPasteboardTypeString);
	        mustReject();
	        board.clearContents;
	        var urls = $.NSMutableArray.array;
	        args.forEach(function(p) { urls.addObject($.NSURL.fileURLWithPath(p)); });
	        if (!board.writeObjects(urls)) throw Error("Could not set up multiple items");
	        mustReject();
	        return JSON.stringify(args.map(function(p) {
	            fileClipboard(board, ["write", p]);
	            return JSON.parse(fileClipboard(board, ["read"]));
	        }));
	    } finally { board.releaseGlobally; }
	}`
	var diagnostics bytes.Buffer
	args := append([]string{"-l", "JavaScript", "-e", script, "--"}, paths...)
	cmd := exec.CommandContext(t.Context(), "/usr/bin/osascript", args...)
	cmd.Stderr = &diagnostics
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("private pasteboard: %v: %s", err, diagnostics.String())
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil || !slices.Equal(got, paths) {
		t.Fatalf("unexpected file URL round trip: %s (%v)", out, err)
	}
}
