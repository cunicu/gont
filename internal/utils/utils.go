package utils

import (
	"math/rand"
	"os"
)

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
