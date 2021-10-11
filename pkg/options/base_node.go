package options

import (
	g "github.com/stv0g/gont/pkg"
)

type ExistingNamespace string
type ExistingDockerContainer string

func (e ExistingNamespace) Apply(n *g.BaseNode) {
	n.ExistingNamespace = string(e)
}

func (d ExistingDockerContainer) Apply(n *g.BaseNode) {
	n.ExistingDockerContainer = string(d)
}
