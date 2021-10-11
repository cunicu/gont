package topo

import g "github.com/stv0g/gont/pkg"

// Star creates a star topology which is identical to a
// tree topology consisting only of a single level
func Star(n *g.Network, opts ...g.Option) ([]*g.Host, error) {
	// opts.Dimensions =

	return Tree(n, opts)
}
