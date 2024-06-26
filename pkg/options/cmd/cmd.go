// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	g "cunicu.li/gont/v2/pkg"
)

// DisableASLR will start the sub-process with
// Address Space Layout Randomization disabled
//
// See: https://en.wikipedia.org/wiki/Address_space_layout_randomization
type DisableASLR bool

func (da DisableASLR) ApplyCmd(d *g.Cmd) {
	d.DisableASLR = bool(da)
}

// Context will start the process with the provided context
// See exec.CommandContext()
type Context struct {
	context.Context
}

func (c Context) ApplyCmd(d *g.Cmd) {
	d.Context = c
}

// A name of an environment variable which should be preserved from the parent
// process. If not provided PATH will be preserved by default.
// See gont.DefaultPreserveEnvVars
type PreserveEnv string

func (e PreserveEnv) ApplyCmd(c *g.Cmd) {
	c.PreserveEnvVars = append(c.PreserveEnvVars, string(e))
}
