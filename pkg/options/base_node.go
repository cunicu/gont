package options

import (
	g "github.com/stv0g/gont/pkg"
)

func Interface(name string, opts ...g.Option) *g.Interface {
	i := &g.Interface{
		Name: name,
	}

	for _, o := range opts {
		switch opt := o.(type) {
		case g.InterfaceOption:
			opt.Apply(i)
		case g.LinkOption:
			opt.Apply(&i.LinkAttrs)
		}
	}

	return i
}

type ExistingNamespace string
type ExistingDockerContainer string

func (e ExistingNamespace) Apply(n *g.BaseNode) {
	n.ExistingNamespace = string(e)
}

func (d ExistingDockerContainer) Apply(n *g.BaseNode) {
	n.ExistingDockerContainer = string(d)
}
