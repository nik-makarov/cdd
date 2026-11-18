package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type entryCopy struct {
	source, destination string
	sourceInfo, dirInfo os.FileInfo
}

func copySourceInfo(path string) (os.FileInfo, error) {
	if path == "/" {
		return nil, errors.New("Cannot copy the filesystem root")
	}
	if !filepath.IsAbs(path) || !utf8.ValidString(path) || filepath.Clean(path) != path {
		return nil, errors.New("Copy requires an absolute file or folder path")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() && !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return nil, errors.New("Choose a regular file, folder, or symlink to copy")
	}
	return info, nil
}

func prepareEntryCopy(source, dir string) (*entryCopy, error) {
	info, err := copySourceInfo(source)
	if err != nil {
		return nil, err
	}
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(dir) || !dirInfo.IsDir() {
		return nil, errors.New("Choose a destination folder")
	}
	name := filepath.Base(source)
	p := &entryCopy{source, filepath.Join(dir, name), info, dirInfo}
	if _, err := p.validate(); err != nil {
		return nil, err
	}
	ext := filepath.Ext(name)
	if info.IsDir() || ext == name {
		ext = ""
	}
	stem := strings.TrimSuffix(name, ext) + " copy"
	for n := 1; ; n++ {
		_, err := os.Lstat(p.destination)
		if errors.Is(err, os.ErrNotExist) {
			return p, nil
		}
		if err != nil {
			return nil, err
		}
		suffix := ""
		if n > 1 {
			suffix = " " + strconv.Itoa(n)
		}
		p.destination = filepath.Join(dir, stem+suffix+ext)
	}
}

// Check ancestors by identity so symlinks and case-insensitive aliases cannot
// direct a recursive copy into the folder being copied.
func (p entryCopy) validate() (string, error) {
	info, err := os.Lstat(p.source)
	if err != nil {
		return "", err
	}
	dir, err := filepath.EvalSymlinks(filepath.Dir(p.destination))
	if err != nil {
		return "", err
	}
	for path := dir; ; path = filepath.Dir(path) {
		parent, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if path == dir && (!os.SameFile(p.dirInfo, parent) || !os.SameFile(p.sourceInfo, info)) {
			return "", errors.New("Source or destination changed. Paste again to review the new path")
		}
		if !info.IsDir() {
			break
		}
		if os.SameFile(info, parent) {
			return "", errors.New("Cannot paste a folder inside itself")
		}
		if path == filepath.Dir(path) {
			break
		}
	}
	return dir, nil
}

func (p entryCopy) run(ctx context.Context) (err error) {
	dir, err := p.validate()
	if err != nil {
		return err
	}
	if _, err := os.Lstat(p.destination); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("Destination already exists. Paste again to choose a free name")
		}
		return err
	}
	parent, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer parent.Close()
	fd, err := parent.Open(".")
	if err != nil {
		return err
	}
	defer fd.Close()
	info, err := fd.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(info, p.dirInfo) {
		return errors.New("Destination folder changed. Paste again to review the new path")
	}
	stage := ".cdd-paste-" + rand.Text()
	if err := parent.Mkdir(stage, 0o700); err != nil {
		return err
	}
	// Pin cleanup to this private directory even if an ancestor is moved.
	root, err := parent.OpenRoot(stage)
	if err != nil {
		return errors.Join(err, parent.Remove(stage))
	}
	defer root.Close()
	defer func() {
		if cleanupErr := errors.Join(discardStagedCopy(root), parent.Remove(stage)); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean up temporary copy: %w", cleanupErr))
		}
	}()
	// Keep the output relative to cp's working directory during the transfer.
	// -P copies symlinks as links, including broken links within a tree.
	cmd := exec.CommandContext(ctx, "/bin/cp", "-pPR", p.source, "item")
	cmd.Dir = filepath.Join(dir, stage)
	if err := runActionCommand(ctx, cmd, false); err != nil {
		return err
	}
	if _, err := p.validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return renameExclusiveAt(int(fd.Fd()), filepath.Join(stage, "item"), filepath.Base(p.destination))
}

func discardStagedCopy(stage *os.Root) error {
	err := stage.RemoveAll("item")
	if !errors.Is(err, os.ErrPermission) {
		return err
	}
	// cp -p can leave read-only directories inside an interrupted private
	// copy. Restore owner access there so cancellation can remove its children.
	err = fs.WalkDir(stage.FS(), "item", func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			return stage.Chmod(path, 0o700)
		}
		return err
	})
	if err != nil {
		return err
	}
	return stage.RemoveAll("item")
}
