package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// FindFiles returns a slice of all files contained in the root directory
// including its subdirectories and theirof.
func FindFiles(root string) ([]string, error) {
	files := []string{}

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			path = strings.TrimPrefix(path, root)
			files = append(files, path)
		}

		return nil
	}); err != nil {
		return []string{}, err
	}

	return files, nil
}
