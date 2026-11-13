// Command release packages binaries, licenses, and curated user-facing notes.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

var platforms = []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64"}

type file struct {
	name string
	data []byte
	mode int64
}

func main() {
	root := flag.String("root", "..", "application source directory")
	version := flag.String("version", "dev", "dev or a stable vX.Y.Z version")
	flag.Parse()
	if err := packageRelease(*root, *version); err != nil {
		fmt.Fprintf(os.Stderr, "release: %v\n", err)
		os.Exit(1)
	}
}

func packageRelease(root, version string) error {
	if version != "dev" && !versionPattern.MatchString(version) {
		return fmt.Errorf("invalid version %q: expected dev or stable vX.Y.Z", version)
	}
	changelog, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		return fmt.Errorf("read release notes: %w", err)
	}
	notes, err := releaseNotes(string(changelog), version)
	if err != nil {
		return err
	}
	common := []file{{name: "RELEASE_NOTES.md", data: []byte(notes), mode: 0o644}}
	for _, input := range []struct{ source, name string }{
		{"bin/LICENSE", "LICENSE"},
		{"bin/THIRD_PARTY_LICENSES.txt", "THIRD_PARTY_LICENSES.txt"},
		{"packaging/README.md", "README.md"},
	} {
		data, err := readNonempty(filepath.Join(root, input.source))
		if err != nil {
			return err
		}
		common = append(common, file{input.name, data, 0o644})
	}
	archives := make([]file, 0, len(platforms))
	var sums strings.Builder
	for _, platform := range platforms {
		binary, err := readNonempty(filepath.Join(root, "bin", "0go-"+platform))
		if err != nil {
			return err
		}
		entries := append([]file{{"0go", binary, 0o755}}, common...)
		data, err := archive(entries)
		if err != nil {
			return fmt.Errorf("archive %s: %w", platform, err)
		}
		name := "0go_" + version + "_" + strings.ReplaceAll(platform, "-", "_") + ".tar.gz"
		archives = append(archives, file{name, data, 0o644})
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(data), name)
	}
	output := filepath.Join(root, "dist", version)
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}
	outputs := append(archives,
		file{"SHA256SUMS", []byte(sums.String()), 0o644},
		file{"RELEASE_NOTES.md", []byte(notes), 0o644},
	)
	for _, item := range outputs {
		if err := os.WriteFile(filepath.Join(output, item.name), item.data, os.FileMode(item.mode)); err != nil {
			return fmt.Errorf("write %s: %w", item.name, err)
		}
	}
	return nil
}

func readNonempty(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read packaging input %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty packaging input %s", path)
	}
	return data, nil
}

func archive(files []file) ([]byte, error) {
	var out bytes.Buffer
	gz := gzip.NewWriter(&out)
	tw := tar.NewWriter(gz)
	for _, item := range files {
		header := &tar.Header{
			Name: item.name, Size: int64(len(item.data)), Mode: item.mode,
			ModTime: time.Unix(0, 0), Format: tar.FormatUSTAR,
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := tw.Write(item.data); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func releaseNotes(changelog, version string) (string, error) {
	wanted := version
	if version == "dev" {
		wanted = "Unreleased"
	}
	lines := strings.Split(changelog, "\n")
	start, matches := -1, 0
	for i, line := range lines {
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "## "))
		if len(fields) == 0 || fields[0] != wanted {
			continue
		}
		if version != "dev" {
			if len(fields) != 3 || fields[1] != "-" {
				return "", fmt.Errorf("expected CHANGELOG heading: ## %s - YYYY-MM-DD", version)
			}
			if _, err := time.Parse("2006-01-02", fields[2]); err != nil {
				return "", fmt.Errorf("invalid release date for %s: %w", version, err)
			}
		}
		start, matches = i+1, matches+1
	}
	if matches != 1 {
		return "", fmt.Errorf("expected exactly one %s section in CHANGELOG.md, found %d", wanted, matches)
	}
	end := start
	for end < len(lines) && !strings.HasPrefix(lines[end], "## ") {
		end++
	}
	body := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	hasContent := false
	for _, line := range lines[start:end] {
		text := strings.TrimSpace(line)
		if text != "" && !strings.HasPrefix(text, "#") {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return "", fmt.Errorf("empty %s release notes", wanted)
	}
	return "# 0go " + version + "\n\n" + body + "\n", nil
}
