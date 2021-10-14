package utils

import (
	"math/rand"
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

func Touch(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return err
	}
	return f.Close()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
