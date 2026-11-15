package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func actionTestHelper() int {
	data, _ := json.Marshal(os.Args[1:])
	if err := os.WriteFile(os.Getenv("CDD_TEST_ACTION_ARGS"), data, 0o600); err != nil {
		return 2
	}
	fmt.Print("child stdout must not reach the selected path")
	switch os.Getenv("CDD_TEST_ACTION") {
	case "fail":
		fmt.Fprint(os.Stderr, strings.Repeat("diagnostic ", 1000))
		return 1
	case "wait":
		time.Sleep(time.Hour)
	}
	return 0
}

func fakeActionCLI(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(executable, filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("CDD_TEST_ACTION", "ok")
	capture := filepath.Join(dir, "args.json")
	t.Setenv("CDD_TEST_ACTION_ARGS", capture)
	return capture
}

func TestQuickLookCommandPreservesLiteralPath(t *testing.T) {
	capture := fakeActionCLI(t, "qlmanage")
	path := filepath.Join(t.TempDir(), "-quote' $() `literal`; file")
	if err := os.WriteFile(path, []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}
	action := externalAction{kind: "preview", path: path}
	if err := runExternalAction(t.Context(), action, "darwin"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(args, []string{"-p", path}) {
		t.Fatalf("arguments changed: %q", args)
	}
	t.Setenv("CDD_TEST_ACTION", "fail")
	if err := runExternalAction(t.Context(), action, "darwin"); err == nil || len(err.Error()) > 4300 {
		t.Fatal("failure diagnostics must be bounded")
	}
	t.Setenv("CDD_TEST_ACTION", "wait")
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if err := runExternalAction(ctx, action, "darwin"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancellation not propagated: %v", err)
	}
}
