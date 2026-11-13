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

func TestReleaseNotes(t *testing.T) {
	changelog := "# Changelog\n\n## Unreleased\n\n- Future change.\n\n## v0.1.0 - 2026-11-13\n\n### Added\n\n- Choose a directory.\n\n## v0.0.1 - 2026-11-01\n\n- Earlier change.\n"
	notes, err := releaseNotes(changelog, "v0.1.0")
	if err != nil || !strings.Contains(notes, "Choose a directory.") || strings.Contains(notes, "Future") || strings.Contains(notes, "Earlier") {
		t.Fatalf("incorrect versioned notes: %q, %v", notes, err)
	}
	for _, bad := range []string{
		"## Unreleased\n- Not published.\n",
		"## v0.1.0 - 2026-02-30\n- Invalid date.\n",
		"## v0.1.0\n- Missing date.\n",
		"## v0.1.0 - 2026-11-13\n\n",
		"## v0.1.0 - 2026-11-13\n\n### Added\n\n",
		"## v0.1.0 - 2026-11-13\n- One.\n## v0.1.0 - 2026-11-13\n- Duplicate.\n",
	} {
		if _, err := releaseNotes(bad, "v0.1.0"); err == nil {
			t.Errorf("accepted invalid release notes: %q", bad)
		}
	}
	if notes, err := releaseNotes(changelog, "dev"); err != nil || !strings.Contains(notes, "Future change.") {
		t.Fatalf("preview notes: %q, %v", notes, err)
	}
}

func releaseFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"bin", "packaging"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	inputs := map[string]string{
		"CHANGELOG.md":                 "## Unreleased\n\n- Preview.\n\n## v0.1.0 - 2026-11-13\n\n- Choose a directory.\n",
		"bin/LICENSE":                  "MIT fixture license\n",
		"bin/THIRD_PARTY_LICENSES.txt": "Upstream fixture license\n",
		"packaging/README.md":          "Run with shell integration.\n",
	}
	for _, platform := range platforms {
		inputs["bin/cdd-"+platform] = "binary for " + platform
	}
	for path, text := range inputs {
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestPackageContentsAndChecksums(t *testing.T) {
	root := releaseFixture(t)
	if err := packageRelease(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "dist", "v0.1.0")
	sums, err := os.ReadFile(filepath.Join(output, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(strings.TrimSpace(string(sums)), "\n")) != 4 {
		t.Fatal("expected exactly four release archive checksums")
	}
	for _, platform := range platforms {
		name := "cdd_v0.1.0_" + strings.ReplaceAll(platform, "-", "_") + ".tar.gz"
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(sums), fmt.Sprintf("%x  %s\n", sha256.Sum256(data), name)) {
			t.Fatalf("incorrect checksum for %s", name)
		}
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		tr := tar.NewReader(gz)
		seen := map[string]bool{}
		for {
			header, err := tr.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			seen[header.Name] = true
			if header.Name == "cdd" && (header.Mode != 0o755 || string(body) != "binary for "+platform) {
				t.Fatalf("wrong executable or mode in %s", name)
			}
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
		for _, required := range []string{"cdd", "LICENSE", "THIRD_PARTY_LICENSES.txt", "README.md", "RELEASE_NOTES.md"} {
			if !seen[required] {
				t.Errorf("missing %s from %s", required, name)
			}
		}
	}
	if err := packageRelease(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(output, "SHA256SUMS"))
	if err != nil || !bytes.Equal(sums, again) {
		t.Fatal("identical inputs must produce identical archives")
	}
}

func TestInvalidPackagingInputs(t *testing.T) {
	for _, version := range []string{"../bad", "v01.0.0", "v0.1.0-rc.1", "0.1.0"} {
		if err := packageRelease(t.TempDir(), version); err == nil {
			t.Errorf("accepted invalid version %q", version)
		}
	}
	root := releaseFixture(t)
	if err := os.Remove(filepath.Join(root, "bin", "LICENSE")); err != nil {
		t.Fatal(err)
	}
	if err := packageRelease(root, "v0.1.0"); err == nil {
		t.Fatal("packaging without the license must fail")
	}
	if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
		t.Fatal("invalid inputs must not produce partial output")
	}
}
