package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLicenseInputs(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "NOTICE", "Upstream attribution\n")
	if _, err := licenseText(dir); !errors.Is(err, errMissingLicense) {
		t.Fatalf("NOTICE without a license must fail: %v", err)
	}
	writeFixture(t, dir, "LICENSE", "  Full license text\n")
	writeFixture(t, dir, "PATENTS", "Patent grant\n")
	writeFixture(t, dir, "README.md", "Do not redistribute this as license text")
	text, err := licenseText(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"  Full license text\n", "Upstream attribution\n", "Patent grant\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing complete input %q", want)
		}
	}
	if strings.Contains(text, "Do not redistribute") {
		t.Fatal("unrelated file included in notices")
	}
	writeFixture(t, dir, "LICENSE", " \n\t")
	if _, err := licenseText(dir); err == nil {
		t.Fatal("empty primary license must fail")
	}
}

func TestGoLicenseHomebrewLayout(t *testing.T) {
	parent := t.TempDir()
	goRoot := filepath.Join(parent, "libexec")
	if err := os.Mkdir(goRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, parent, "LICENSE", "Distribution license\n")
	text, err := goLicense(goRoot)
	if err != nil || !strings.Contains(text, "Distribution license\n") {
		t.Fatalf("parent license fallback failed: %q, %v", text, err)
	}
	writeFixture(t, goRoot, "LICENSE", "\n")
	if _, err := goLicense(goRoot); err == nil {
		t.Fatal("empty GOROOT license must fail instead of falling back")
	}
}

func TestApplicationNoticesExcludeTools(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	text, err := collect(t.Context(), root, []target{{runtime.GOOS, runtime.GOARCH}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"golang.org/x/term v", "golang.org/x/sys v", "Go standard library go", "Copyright", "Redistribution"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing runtime notice %q", want)
		}
	}
	if strings.Contains(text, "golang.org/x/vuln") || strings.Contains(text, "golang.org/x/tools") {
		t.Fatal("development-only module included in application notices")
	}
	if strings.Contains(text, root) {
		t.Fatal("local filesystem path leaked into notices")
	}
}
