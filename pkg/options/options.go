// Package options contains all the general options for configuring the general objects like hosts, switches, routers and interfaces
package options

import g "github.com/stv0g/gont/pkg"

// Customize clones and extends a list of options without altering the list of base options.
func Customize(opts []g.Option, extraOptions ...g.Option) []g.Option {
	new := []g.Option{}

	new = append(new, opts...)
	new = append(new, extraOptions...)

	return new
}
