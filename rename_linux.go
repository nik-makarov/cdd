package main

import "golang.org/x/sys/unix"

func renameExclusive(from, to string) error {
	return renameExclusiveAt(unix.AT_FDCWD, from, to)
}

func renameExclusiveAt(dir int, from, to string) error {
	return unix.Renameat2(dir, from, dir, to, unix.RENAME_NOREPLACE)
}
