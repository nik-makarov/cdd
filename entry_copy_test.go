package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPasteCopiesEntriesExclusively(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "-folder '$() \u754c#")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(source, "a\n.txt")
	if err := os.WriteFile(file, []byte{0, 7, 255}, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(source, "link")
	if err := os.Symlink("missing", link); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, path, want string
		existing         []string
	}{
		{"folder", source, filepath.Base(source) + " copy", nil},
		{"file", file, "a\n.txt", nil},
		{"broken link", link, "link", nil},
		{"repeated file", file, "a\n copy.txt", []string{"a\n.txt"}},
		{"numbered file", file, "a\n copy 2.txt", []string{"a\n.txt", "a\n copy.txt"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			if tt.path == source {
				parent = dir
			}
			for _, name := range tt.existing {
				if err := os.Symlink("untouched", filepath.Join(parent, name)); err != nil {
					t.Fatal(err)
				}
			}
			p, err := prepareEntryCopy(tt.path, parent)
			if err != nil || p.destination != filepath.Join(parent, tt.want) {
				t.Fatalf("prepare: %+v, %v", p, err)
			}
			if err := p.run(t.Context()); err != nil {
				t.Fatal(err)
			}
			if err := p.run(t.Context()); err == nil {
				t.Fatal("overwrote existing copy on retry")
			}
			path := p.destination
			if tt.path == source {
				if target, err := os.Readlink(filepath.Join(path, "link")); err != nil || target != "missing" {
					t.Fatalf("nested symlink: %q, %v", target, err)
				}
				path = filepath.Join(path, filepath.Base(file))
			}
			if tt.path == link {
				if target, err := os.Readlink(path); err != nil || target != "missing" {
					t.Fatalf("symlink: %q, %v", target, err)
				}
			} else if data, err := os.ReadFile(path); err != nil || string(data) != "\x00\x07\xff" {
				t.Fatalf("copied file bytes: %x, %v", data, err)
			}
		})
	}
}

func TestPasteRejectsRecursiveAndChangedPaths(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(sub, alias); err != nil {
		t.Fatal(err)
	}
	for _, dest := range []string{dir, sub, alias} {
		if _, err := prepareEntryCopy(dir, dest); err == nil {
			t.Fatalf("allowed pasting folder inside itself at %q", dest)
		}
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := prepareEntryCopy(file, alias)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := p.run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation must propagate", err)
	}
	if items, err := os.ReadDir(sub); err != nil || len(items) != 0 {
		t.Fatal("cancelled paste left files behind", items, err)
	}
	if err := os.Rename(file, file+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := p.run(t.Context()); err == nil {
		t.Fatal("accepted a replaced source")
	}
	p, err = prepareEntryCopy(file, alias)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(sub, sub+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := p.run(t.Context()); err == nil {
		t.Fatal("accepted a replaced destination folder")
	}
}

func TestConcurrentPasteKeepsOneCompleteCopy(t *testing.T) {
	root := fixture(t)
	source, dir := filepath.Join(root, "Alpha"), filepath.Join(root, "beta")
	p, err := prepareEntryCopy(source, dir)
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	for range 2 {
		go func() { results <- p.run(t.Context()) }()
	}
	first, second := <-results, <-results
	if (first == nil) == (second == nil) {
		t.Fatalf("exactly one paste should succeed: %v, %v", first, second)
	}
	items, err := os.ReadDir(dir)
	if err != nil || len(items) != 1 || items[0].Name() != "Alpha" {
		t.Fatalf("unexpected destination entries: %v, %v", items, err)
	}
	if info, err := os.Stat(filepath.Join(p.destination, "nested")); err != nil || !info.IsDir() {
		t.Fatal("published an incomplete copy", err)
	}
}

func TestDiscardReadOnlyStagedCopy(t *testing.T) {
	dir := t.TempDir()
	item := filepath.Join(dir, "item")
	if err := os.Mkdir(item, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(item, "file"), nil, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(item, 0o500); err != nil {
		t.Fatal(err)
	}
	stage, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer stage.Close()
	if err := discardStagedCopy(stage); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(item); !os.IsNotExist(err) {
		t.Fatal("read-only staging data survived cleanup", err)
	}
}
