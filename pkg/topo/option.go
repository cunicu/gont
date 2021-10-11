package topo

import g "github.com/stv0g/gont/pkg"

type TopoOption interface {
	g.Option
	apply(t *Topo)
}
