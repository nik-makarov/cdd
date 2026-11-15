package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Opt in explicitly: this checks the real system Trash using disposable entries
// and restores each one. The normal suite uses a fake Trash CLI instead.
func TestNativeTrash(t *testing.T) {
	if os.Getenv("CDD_TEST_NATIVE_TRASH") != "1" {
		t.Skip("set CDD_TEST_NATIVE_TRASH=1 to verify the system Trash")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"file", "dir", "link", "broken-link"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			name := filepath.Base(filepath.Dir(root)) + " ' $() \u754c " + kind
			path, trashed := filepath.Join(root, name), filepath.Join(home, ".Trash", name)
			if _, err := os.Lstat(trashed); !os.IsNotExist(err) {
				t.Fatalf("cannot reserve a unique Trash destination: %v", err)
			}
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(path, []byte("disposable native Trash fixture"), 0o600)
			case "dir":
				err = os.Mkdir(path, 0o700)
				if err == nil {
					err = os.WriteFile(filepath.Join(path, "child"), []byte("retained"), 0o600)
				}
			case "link", "broken-link":
				target := filepath.Join(root, "target")
				if kind == "link" {
					if err := os.Mkdir(target, 0o700); err != nil {
						t.Fatal(err)
					}
				}
				err = os.Symlink(target, path)
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := os.Lstat(trashed); os.IsNotExist(err) {
					return
				}
				if err := renameExclusive(trashed, path); err != nil {
					t.Errorf("restore disposable fixture %q: %v", trashed, err)
				}
			})
			m := focus(t, loadBrowser(t, root), name)
			m = press(m, "trash", "")
			m = press(m, "text", name)
			m = press(m, "enter", "")
			if err := runExternalAction(t.Context(), m.action, "darwin"); err != nil {
				t.Fatal(err)
			}
			info, err := os.Lstat(trashed)
			if err != nil || !os.SameFile(m.action.target.info, info) {
				t.Fatalf("original entry is not recoverable in Trash: %v", err)
			}
			if kind == "link" {
				if info, err := os.Stat(filepath.Join(root, "target")); err != nil || !info.IsDir() {
					t.Fatalf("symlink target was affected: %v", err)
				}
			}
		})
	}
}
