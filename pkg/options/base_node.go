package options

import (
	g "github.com/stv0g/gont/pkg"
)

type ExistingNamespace string
type DockerContainer string

func (e ExistingNamespace) Apply(n *g.BaseNode) {
	n.ExistingNamespace = string(e)
}

func (d DockerContainer) Apply(n *g.BaseNode) {
	n.DockerContainer = string(d)
}
