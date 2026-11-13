package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func gitFixture(t *testing.T) string {
	t.Helper()
	root := releaseFixture(t)
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.name", "Fixture"},
		{"config", "user.email", "fixture@example.invalid"},
	} {
		gitOK(t, root, args...)
	}
	writeFixture(t, root, "VERSION", "0.1.0\n")
	// Generated build inputs are deliberately outside the source tree.
	writeFixture(t, root, ".gitignore", "/bin/\n/dist/\n")
	gitOK(t, root, "add", ".gitignore", "VERSION", "CHANGELOG.md", "packaging/README.md")
	gitOK(t, root, "commit", "-m", "Add directory selection")
	return root
}

func gitOK(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := git(root, args...)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func writeFixture(t *testing.T, root, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAutomaticVersionAndRetry(t *testing.T) {
	root := gitFixture(t)
	assertNext := func(want string) {
		t.Helper()
		got, err := nextVersion(root)
		if err != nil || got != want {
			t.Fatalf("next version = %q, %v; want %q", got, err, want)
		}
	}
	assertNext("v0.1.0")
	gitOK(t, root, "tag", "v0.1.0")
	assertNext("v0.1.0")
	writeFixture(t, root, "change.txt", "new behavior\n")
	gitOK(t, root, "add", "change.txt")
	gitOK(t, root, "commit", "-m", "Improve directory selection")
	assertNext("v0.1.1")
	gitOK(t, root, "tag", "v0.1.9")
	writeFixture(t, root, "change.txt", "another behavior\n")
	gitOK(t, root, "commit", "-am", "Preserve folder focus")
	assertNext("v0.1.10") // Numeric ordering, rather than lexical ordering.
	writeFixture(t, root, "VERSION", "0.2.0\n")
	assertNext("v0.2.0")
	gitOK(t, root, "tag", "v0.3.0-rc.1")
	assertNext("v0.2.0") // Prerelease tags do not consume stable versions.
	writeFixture(t, root, "VERSION", "0.02.0\n")
	if _, err := nextVersion(root); err == nil {
		t.Fatal("accepted a noncanonical base version")
	}
}

func TestRefuseOutOfOrderRelease(t *testing.T) {
	root := gitFixture(t)
	first := gitOK(t, root, "rev-parse", "HEAD")
	writeFixture(t, root, "change.txt", "later\n")
	gitOK(t, root, "add", "change.txt")
	gitOK(t, root, "commit", "-m", "Add later behavior")
	gitOK(t, root, "tag", "v0.1.0")
	gitOK(t, root, "checkout", "--detach", first)
	if _, err := nextVersion(root); err == nil {
		t.Fatal("an old commit must not replace the latest release")
	}
}

func TestAutomaticNotes(t *testing.T) {
	root := gitFixture(t)
	changelog := "## Unreleased\n\n- Browse folders.\n"
	notes, err := automaticNotes(root, "v0.1.0", changelog)
	if err != nil || !strings.Contains(notes, "# cdd v0.1.0\n") || !strings.Contains(notes, "Browse folders.") {
		t.Fatalf("first release notes: %q, %v", notes, err)
	}
	gitOK(t, root, "tag", "v0.1.0")
	writeFixture(t, root, "change.txt", "later\n")
	gitOK(t, root, "add", "change.txt")
	gitOK(t, root, "commit", "-m", "Preserve folder focus")
	notes, err = automaticNotes(root, "v0.1.1", changelog)
	if err != nil || !strings.Contains(notes, "- Preserve folder focus\n") || strings.Contains(notes, "Browse folders.") || !strings.Contains(notes, "v0.1.0...v0.1.1") {
		t.Fatalf("incremental release notes: %q, %v", notes, err)
	}
	gitOK(t, root, "tag", "v0.1.1")
	again, err := automaticNotes(root, "v0.1.1", changelog)
	if err != nil || notes != again {
		t.Fatalf("retry changed notes: %q, %v", again, err)
	}
	curated := changelog + "\n## v0.1.1 - 2026-11-13\n\n- Keep the highlighted folder after refresh.\n"
	notes, err = automaticNotes(root, "v0.1.1", curated)
	if err != nil || !strings.Contains(notes, "Keep the highlighted folder") || strings.Contains(notes, "Preserve folder focus") {
		t.Fatalf("curated notes not preferred: %q, %v", notes, err)
	}
	if _, err := automaticNotes(root, "v0.1.1", "## v0.1.1 - wrong-date\n- Change.\n"); err == nil {
		t.Fatal("invalid curated notes must not fall back to commit messages")
	}
}

func TestCommittedSourceAndFormulaChecksum(t *testing.T) {
	root := gitFixture(t)
	if err := packageReleaseMode(root, "v0.1.0", true); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "dist", "v0.1.0")
	source, err := os.ReadFile(filepath.Join(output, "cdd_v0.1.0_source.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	sum := fmt.Sprintf("%x", sha256.Sum256(source))
	for name, required := range map[string]string{
		"SHA256SUMS": sum + "  cdd_v0.1.0_source.tar.gz",
		"cdd.rb":     `sha256 "` + sum + `"`,
	} {
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || !strings.Contains(string(data), required) {
			t.Fatalf("incorrect source checksum in %s: %q, %v", name, data, err)
		}
	}
	gz, err := gzip.NewReader(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := gz.Close(); err != nil {
			t.Error(err)
		}
	}()
	tr := tar.NewReader(gz)
	seenVersion := false
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(header.Name, "cdd-v0.1.0/") || strings.Contains(header.Name, "/bin/") || header.Uid != 0 || header.Gid != 0 {
			t.Fatalf("unexpected source archive entry: %+v", header)
		}
		if header.Name == "cdd-v0.1.0/VERSION" {
			data, err := io.ReadAll(tr)
			if err != nil || string(data) != "0.1.0\n" {
				t.Fatalf("wrong committed version: %q, %v", data, err)
			}
			seenVersion = true
		}
	}
	if !seenVersion {
		t.Fatal("source archive is missing VERSION")
	}
	writeFixture(t, root, "VERSION", "0.2.0\n")
	again, err := sourceArchive(root, "v0.1.0")
	if err != nil || !bytes.Equal(source, again) {
		t.Fatalf("working tree changed committed source: %v", err)
	}
	if err := packageReleaseMode(root, "v0.1.0", true); err == nil {
		t.Fatal("automatic packaging accepted uncommitted source")
	}
	gitOK(t, root, "checkout", "--", "VERSION")
	gitOK(t, root, "commit", "--allow-empty", "-m", "Change metadata only")
	again, err = sourceArchive(root, "v0.1.0")
	if err != nil || !bytes.Equal(source, again) {
		t.Fatalf("commit metadata changed source checksum: %v", err)
	}
}
