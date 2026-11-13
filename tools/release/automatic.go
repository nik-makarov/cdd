package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type releaseVersion [3]uint64

func parseVersion(tag string) (releaseVersion, error) {
	var version releaseVersion
	if !versionPattern.MatchString(tag) {
		return version, fmt.Errorf("invalid stable version %q", tag)
	}
	for i, part := range strings.Split(tag[1:], ".") {
		number, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return version, fmt.Errorf("parse version %q: %w", tag, err)
		}
		version[i] = number
	}
	return version, nil
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func stableTags(root string) ([]string, error) {
	out, err := git(root, "tag", "--list", "v*")
	if err != nil {
		return nil, err
	}
	var tags []string
	for tag := range strings.FieldsSeq(out) {
		if versionPattern.MatchString(tag) {
			if _, err := parseVersion(tag); err != nil {
				return nil, err
			}
			tags = append(tags, tag)
		}
	}
	slices.SortFunc(tags, func(a, b string) int {
		va, _ := parseVersion(a) // Validated above.
		vb, _ := parseVersion(b)
		return slices.Compare(va[:], vb[:])
	})
	return tags, nil
}

func nextVersion(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("read base version: %w", err)
	}
	base := "v" + strings.TrimSpace(string(data))
	baseVersion, err := parseVersion(base)
	if err != nil {
		return "", err
	}
	tags, err := stableTags(root)
	if err != nil || len(tags) == 0 {
		return base, err
	}
	latest := tags[len(tags)-1]
	if _, err := git(root, "merge-base", "--is-ancestor", latest, "HEAD"); err != nil {
		return "", fmt.Errorf("latest release %s is not an ancestor of HEAD; refuse an out-of-order release: %w", latest, err)
	}
	head, err := git(root, "tag", "--points-at", "HEAD")
	if err != nil {
		return "", err
	}
	if slices.Contains(strings.Fields(head), latest) {
		return latest, nil // Retry the same commit without assigning a new version.
	}
	previous, err := parseVersion(latest)
	if err != nil {
		return "", err
	}
	if slices.Compare(baseVersion[:], previous[:]) > 0 {
		return base, nil
	}
	if previous[2] == ^uint64(0) {
		return "", fmt.Errorf("patch version overflow")
	}
	return fmt.Sprintf("v%d.%d.%d", previous[0], previous[1], previous[2]+1), nil
}

func automaticNotes(root, version, changelog string) (string, error) {
	// An explicitly curated section takes precedence, and must remain valid.
	for line := range strings.SplitSeq(changelog, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "##" && fields[1] == version {
			return releaseNotes(changelog, version)
		}
	}
	tags, err := stableTags(root)
	if err != nil {
		return "", err
	}
	previous := ""
	for _, tag := range tags {
		if tag != version {
			previous = tag
		}
	}
	if previous == "" {
		notes, err := releaseNotes(changelog, "dev")
		return strings.Replace(notes, "# cdd dev\n", "# cdd "+version+"\n", 1), err
	}
	if _, err := git(root, "merge-base", "--is-ancestor", previous, "HEAD"); err != nil {
		return "", err
	}
	log, err := git(root, "log", "--reverse", "--no-merges", "--format=%s", previous+"..HEAD")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(log) == "" {
		return "", fmt.Errorf("no changes since %s", previous)
	}
	var notes strings.Builder
	fmt.Fprintf(&notes, "# cdd %s\n\n## Changes\n\n", version)
	for subject := range strings.SplitSeq(strings.TrimSpace(log), "\n") {
		fmt.Fprintf(&notes, "- %s\n", subject)
	}
	fmt.Fprintf(&notes, "\n[Full comparison](https://github.com/nik-makarov/cdd/compare/%s...%s)\n", previous, version)
	return notes.String(), nil
}

func sourceArchive(root, version string) ([]byte, error) {
	// Re-encode Git's committed tree to avoid timestamps, ownership, PAX commit
	// comments, or compression metadata depending on the host and Git version.
	data, err := git(root, "archive", "--format=tar", "HEAD")
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	gz := gzip.NewWriter(&output)
	tw := tar.NewWriter(gz)
	tr := tar.NewReader(strings.NewReader(data))
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read Git archive: %w", err)
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeSymlink {
			return nil, fmt.Errorf("unsupported source entry %q", header.Name)
		}
		if !filepath.IsLocal(header.Name) {
			return nil, fmt.Errorf("unsafe source entry %q", header.Name)
		}
		entry := &tar.Header{
			Name: "cdd-" + version + "/" + header.Name, Mode: header.Mode,
			Typeflag: header.Typeflag, Size: header.Size, Linkname: header.Linkname,
			ModTime: time.Unix(0, 0), Format: tar.FormatPAX,
		}
		if err := tw.WriteHeader(entry); err != nil {
			return nil, fmt.Errorf("write source header: %w", err)
		}
		if _, err := io.Copy(tw, tr); err != nil {
			return nil, fmt.Errorf("write source content: %w", err)
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func formula(version, sum string) string {
	return fmt.Sprintf(`class Cdd < Formula
  desc "Interactive directory picker with shell integration"
  homepage "https://github.com/nik-makarov/cdd"
  url "https://github.com/nik-makarov/cdd/releases/download/%s/cdd_%s_source.tar.gz"
  version "%s"
  sha256 "%s"
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build", "VERSION=v#{version}"
    bin.install "bin/cdd"
    (pkgshare/"licenses").install "bin/LICENSE", "bin/THIRD_PARTY_LICENSES.txt"
  end

  def caveats
    <<~EOS
      Add to ~/.zshrc:
        eval "$(#{HOMEBREW_PREFIX}/bin/cdd --init zsh)"
      For bash, use --init bash in ~/.bashrc.
      After brew upgrade, rerun the init line in existing shells.
    EOS
  end

  test do
    assert_equal "cdd v#{version}\n", shell_output("#{bin}/cdd --version")
    assert_match "cdd() {", shell_output("#{bin}/cdd --init zsh")
    assert_match "cdd() {", shell_output("#{bin}/cdd --init bash")
    assert_match "Usage:", shell_output("#{bin}/cdd --help 2>&1")
  end
end
`, version, version, strings.TrimPrefix(version, "v"), sum)
}
