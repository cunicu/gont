// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package options contains all the general options for configuring the general objects like hosts, switches, routers and interfaces
package options

// Customize clones and extends a list of options without altering the list of base options.
func Customize[T any](opts []T, extraOptions ...T) []T {
	new := make([]T, len(opts)+len(extraOptions))
	new = append(new, opts...)
	new = append(new, extraOptions...)

	return new
}
