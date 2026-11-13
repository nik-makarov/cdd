package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWrapErrorPreservesCause(t *testing.T) {
	cause := &os.PathError{Op: "open", Path: "/missing", Err: os.ErrNotExist}
	err := wrapError("read directory", wrapError("load entries", cause))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("wrapped error lost its sentinel cause")
	}
	if got, ok := errors.AsType[*os.PathError](err); !ok || got != cause {
		t.Fatal("wrapped error lost its typed cause")
	}
	if err.Error() != "read directory: load entries: "+cause.Error() {
		t.Fatalf("context changed: %q", err)
	}
	if wrapError("no failure", nil) != nil {
		t.Fatal("wrapping nil must return nil")
	}
}

func TestDirectoryErrorPreservesPathAndCause(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing\n\"directory")
	_, err := newBrowser(path, false)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("directory error lost its cause: %v", err)
	}
	cause, ok := errors.AsType[*os.PathError](err)
	if !ok || cause.Path != path || err.Error() != "read directory "+strconv.Quote(path)+": "+cause.Error() {
		t.Fatalf("directory error lost its quoted path or operation: %v", err)
	}
}
