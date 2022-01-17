package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stv0g/gont/internal/utils"
)

func TestRandStringRunes(t *testing.T) {
	rnd := utils.RandStringRunes(16)

	if len(rnd) != 16 {
		t.Fail()
	}
}

func TestTouch(t *testing.T) {
	dir, err := os.MkdirTemp("", "gont-test")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test-file")

	if err := utils.Touch(fn); err != nil {
		t.Fail()
	}

	fns, err := utils.FindFiles(dir)
	if err != nil {
		t.Fail()
	}

	found := false
	for _, f := range fns {
		if f == "/test-file" {
			found = true
		}
	}

	if !found {
		t.Fail()
	}
}
