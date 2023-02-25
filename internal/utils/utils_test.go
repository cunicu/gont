// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stv0g/gont/internal/utils"
)

func TestRandStringRunes(t *testing.T) {
	rnd := utils.RandStringRunes(16)
	assert.Len(t, rnd, 16)
}

func TestTouch(t *testing.T) {
	dir, err := os.MkdirTemp("", "gont-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test-file")

	err = utils.Touch(fn)
	assert.NoError(t, err)

	fi, err := os.Stat(fn)
	assert.NoError(t, err)

	assert.False(t, fi.IsDir())

	assert.Equal(t, fi.Size(), 0)
}
