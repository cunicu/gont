package options

import (
	g "github.com/stv0g/gont/pkg"
)

func Interface(name string, opts ...g.Option) *g.Interface {
	i := &g.Interface{
		Name:     name,
		Captures: []*g.Capture{},
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

// The name of an existing network namespace which is used instead of creating a new one.
type ExistingNamespace string

func (e ExistingNamespace) Apply(n *g.BaseNode) {
	n.ExistingNamespace = string(e)
}

// Name of an existing Docker container which is used for this node
type ExistingDockerContainer string

func (d ExistingDockerContainer) Apply(n *g.BaseNode) {
	n.ExistingDockerContainer = string(d)
}

// Log output of sub-processes to debug log-level
type LogToDebug bool

func (l LogToDebug) Apply(n *g.BaseNode) {
	n.LogToDebug = bool(l)
}

// Mount an empty dir to shadow parts of the root filesystem
type EmptyDir string

func (ed EmptyDir) Apply(n *g.BaseNode) {
	n.EmptyDirs = append(n.EmptyDirs, string(ed))
}
