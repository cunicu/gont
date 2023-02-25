// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cmd

import g "github.com/stv0g/gont/pkg"

// DisableASLR will start the sub-process with
// Address Space Layout Randomization disabled
//
// See: https://en.wikipedia.org/wiki/Address_space_layout_randomization
type DisableASLR bool

func (da DisableASLR) ApplyCmd(d *g.Cmd) {
	d.DisableASLR = bool(da)
}