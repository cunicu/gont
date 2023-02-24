package cmd

import g "github.com/stv0g/gont/pkg"

type DisableASLR bool

func (da DisableASLR) ApplyCmd(d *g.Cmd) {
	d.DisableASLR = bool(da)
}
