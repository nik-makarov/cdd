package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
)

// Only file URL items participate. A text path copied with y is a different
// pasteboard type, even when it happens to name an existing file.
const fileClipboardScript = `ObjC.import('AppKit');
function fileClipboard(board, args) {
    if (args[0] === "write") {
        var url = $.NSURL.fileURLWithPath(args[1]);
        board.clearContents;
        if (!board.writeObjects($.NSArray.arrayWithObject(url))) throw Error("Could not copy the file to the clipboard");
        return;
    }
    var items = board.pasteboardItems;
    if (Number(items.count) !== 1) throw Error("Copy one file or folder before pasting");
    var value = items.objectAtIndex(0).stringForType($.NSPasteboardTypeFileURL);
    if (!ObjC.unwrap(value)) throw Error("The clipboard does not contain a file or folder");
    var url = $.NSURL.URLWithString(value);
    var host = ObjC.unwrap(url.host);
    if (!url.isFileURL || (host && host !== "localhost")) throw Error("Expected a local file URL");
    var path = ObjC.unwrap(url.filePathURL.path);
    if (!path || path.length > 4096) throw Error("Invalid clipboard file path");
    return JSON.stringify(path);
}
function run(args) { return fileClipboard($.NSPasteboard.generalPasteboard, args); }`

func readClipboardEntry(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "/usr/bin/osascript", "-l", "JavaScript", "-e", fileClipboardScript, "--", "read")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := runActionCommand(ctx, cmd, false); err != nil {
		return "", err
	}
	var path string
	if err := json.Unmarshal(out.Bytes(), &path); err != nil {
		return "", errors.New("Invalid file URL on the clipboard")
	}
	return path, nil
}
