package options

import (
	t "github.com/stv0g/gont/pkg/topo"
)

type CreateHost t.CreateHost
type LinkHosts t.LinkHosts

func (ch CreateHost) Apply(topo *t.Topo) {
	topo.CreateHost = t.CreateHost(ch)
}

func (lh LinkHosts) Apply(topo *t.Topo) {
	topo.LinkHosts = t.LinkHosts(lh)
}
