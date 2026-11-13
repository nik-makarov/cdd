// Command notices collects licenses for the application, excluding development tools.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var errMissingLicense = errors.New("missing primary license")

type target struct{ os, arch string }

var targets = []target{
	{"darwin", "amd64"}, {"darwin", "arm64"},
	{"linux", "amd64"}, {"linux", "arm64"},
}

type module struct {
	Path, Version, Dir string
	Main               bool
	Replace            *module
}

func main() {
	root := flag.String("root", "..", "application source directory")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := generate(ctx, *root); err != nil {
		fmt.Fprintf(os.Stderr, "notices: %v\n", err)
		os.Exit(1)
	}
}

func generate(ctx context.Context, root string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve application directory: %w", err)
	}
	text, err := collect(ctx, root, targets)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "THIRD_PARTY_LICENSES.txt"), []byte(text), 0o644); err != nil {
		return fmt.Errorf("write notices: %w", err)
	}
	return nil
}

func collect(ctx context.Context, root string, platforms []target) (string, error) {
	modules := make(map[string]module)
	for _, platform := range platforms {
		cmd := exec.CommandContext(ctx, "go", "list", "-mod=readonly", "-deps", "-json", ".")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOOS="+platform.os, "GOARCH="+platform.arch, "CGO_ENABLED=0")
		data, err := commandOutput(cmd)
		if err != nil {
			return "", fmt.Errorf("list application packages for %s/%s: %w", platform.os, platform.arch, err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		for {
			var pkg struct{ Module *module }
			if err := decoder.Decode(&pkg); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return "", fmt.Errorf("decode application package: %w", err)
			}
			if pkg.Module == nil || pkg.Module.Main {
				continue
			}
			m := *pkg.Module
			if m.Replace != nil {
				return "", fmt.Errorf("module %s has a replacement; review its licensing before packaging", m.Path)
			}
			modules[m.Path] = m
		}
	}
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	slices.Sort(names)
	var out strings.Builder
	out.WriteString("Third-party licenses for 0go\n\n")
	for _, name := range names {
		m := modules[name]
		text, err := licenseText(m.Dir)
		if err != nil {
			return "", fmt.Errorf("license for %s %s: %w", m.Path, m.Version, err)
		}
		fmt.Fprintf(&out, "=== %s %s ===\n%s\n", m.Path, m.Version, text)
	}
	cmd := exec.CommandContext(ctx, "go", "env", "-json", "GOROOT", "GOVERSION")
	cmd.Dir = root
	data, err := commandOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("locate Go toolchain: %w", err)
	}
	var env struct{ GOROOT, GOVERSION string }
	if err := json.Unmarshal(data, &env); err != nil {
		return "", fmt.Errorf("decode Go toolchain metadata: %w", err)
	}
	text, err := goLicense(env.GOROOT)
	if err != nil {
		return "", fmt.Errorf("license for Go standard library: %w", err)
	}
	fmt.Fprintf(&out, "=== Go standard library %s ===\n%s", env.GOVERSION, text)
	return out.String(), nil
}

func commandOutput(cmd *exec.Cmd) ([]byte, error) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return data, nil
}

func goLicense(goRoot string) (string, error) {
	text, err := licenseText(goRoot)
	if errors.Is(err, errMissingLicense) {
		// Homebrew keeps LICENSE alongside libexec instead of inside GOROOT.
		return licenseText(filepath.Dir(goRoot))
	}
	return text, err
}

func licenseText(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read license directory: %w", err)
	}
	var out strings.Builder
	primaryFound := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToUpper(entry.Name())
		primary := strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(name, "LICENCE") || strings.HasPrefix(name, "COPYING")
		if !primary && !strings.HasPrefix(name, "NOTICE") && !strings.HasPrefix(name, "PATENTS") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		if len(bytes.TrimSpace(data)) == 0 {
			return "", fmt.Errorf("empty license input: %s", entry.Name())
		}
		primaryFound = primaryFound || primary
		fmt.Fprintf(&out, "[%s]\n%s", entry.Name(), data)
		if !bytes.HasSuffix(data, []byte("\n")) {
			out.WriteByte('\n')
		}
		out.WriteByte('\n')
	}
	if !primaryFound {
		return "", fmt.Errorf("%w in %s", errMissingLicense, dir)
	}
	return out.String(), nil
}
