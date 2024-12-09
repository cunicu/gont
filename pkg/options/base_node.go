// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "cunicu.li/gont/v2/pkg"
)

// HostNamespace creates the node using the network namespace of the calling process
const HostNamespace = ExistingNetworkNamespace("host")

// The name of an existing network namespace which is used instead of creating a new one.
type ExistingNetworkNamespace string

func (e ExistingNetworkNamespace) ApplyBaseNode(n *g.BaseNode) {
	n.ExistingNetworkNamespace = string(e)
}

// Name of an existing Docker container which is used for this node
type ExistingDockerContainer string

func (d ExistingDockerContainer) ApplyBaseNode(n *g.BaseNode) {
	n.ExistingDockerContainer = string(d)
}

// Mount an empty dir to shadow parts of the root filesystem
type EmptyDir string

func (ed EmptyDir) ApplyBaseNode(n *g.BaseNode) {
	n.EmptyDirs = append(n.EmptyDirs, string(ed))
}

type GoBuildFlags []string

func (bf GoBuildFlags) ApplyGoBuildFlags(d *g.GoBuildFlags) {
	*d = append(*d, bf...)
}

func BuildFlags(flags ...string) GoBuildFlags {
	return GoBuildFlags(flags)
}

// BuildFlagsDebug builds the Go binary without compiler optimizations like inlining
// to improve debugging.
var BuildFlagsDebug = GoBuildFlags{"-gcflags", "-N -l"} //nolint:gochecknoglobals
