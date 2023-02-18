// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "github.com/stv0g/gont/pkg"
)

// The name of an existing network namespace which is used instead of creating a new one.
type ExistingNamespace string

func (e ExistingNamespace) ApplyBaseNode(n *g.BaseNode) {
	n.ExistingNamespace = string(e)
}

// Name of an existing Docker container which is used for this node
type ExistingDockerContainer string

func (d ExistingDockerContainer) ApplyBaseNode(n *g.BaseNode) {
	n.ExistingDockerContainer = string(d)
}

// Log output of sub-processes to debug log-level
type LogToDebug bool

func (l LogToDebug) ApplyBaseNode(n *g.BaseNode) {
	n.LogToDebug = bool(l)
}

// Mount an empty dir to shadow parts of the root filesystem
type EmptyDir string

func (ed EmptyDir) ApplyBaseNode(n *g.BaseNode) {
	n.EmptyDirs = append(n.EmptyDirs, string(ed))
}

// Extra environment variable for processes started in the nodes network namespace
type ExtraEnv struct {
	Name  string
	Value any
}

func (ev ExtraEnv) ApplyBaseNode(n *g.BaseNode) {
	n.Env[ev.Name] = ev.Value
}
