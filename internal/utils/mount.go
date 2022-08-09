package utils

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func IsMountPoint(path string) (bool, error) {
	var stat, statParent unix.Stat_t

	if err := unix.Lstat(path, &stat); err != nil {
		return false, &os.PathError{Op: "stat", Path: path, Err: err}
	}

	parentPath := filepath.Dir(path)
	if err := unix.Lstat(parentPath, &statParent); err != nil {
		return false, &os.PathError{Op: "stat", Path: parentPath, Err: err}
	}

	if stat.Dev != statParent.Dev {
		return true, nil
	}

	return false, nil
}
