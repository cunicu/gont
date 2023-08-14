// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"cunicu.li/gont/v2/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestRandStringRunes(t *testing.T) {
	rnd := utils.RandStringRunes(16)
	require.Len(t, rnd, 16)
}

func TestTouch(t *testing.T) {
	dir, err := os.MkdirTemp("", "gont-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "test-file")

	err = utils.Touch(fn)
	require.NoError(t, err)

	fi, err := os.Stat(fn)
	require.NoError(t, err)

	require.False(t, fi.IsDir())

	require.Zero(t, fi.Size())
}
