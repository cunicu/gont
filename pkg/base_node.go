// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type BaseNodeOption interface {
	ApplyBaseNode(n *BaseNode)
}

type BaseNode struct {
	network *Network
	name    string

	Interfaces []*Interface
	BasePath   string

	// Options
	ConfiguredInterfaces []*Interface

	logger *zap.Logger
}

func (n *Network) addBaseNode(name string, opts ...Option) (*BaseNode, error) {
	node := &BaseNode{
		name:    name,
		network: n,
		logger:  zap.L().Named("node").With(zap.String("node", name)),
	}

	// Apply host options
	for _, opt := range opts {
		if bnOpt, ok := opt.(BaseNodeOption); ok {
			bnOpt.ApplyBaseNode(node)
		}
	}

	node.BasePath = filepath.Join(n.VarPath, "nodes", name)
	for _, path := range []string{"ns", "files"} {
		path = filepath.Join(node.BasePath, path)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
	}

	node.logger.Info("Adding new node")

	return node, nil
}

func (n *BaseNode) Name() string {
	return n.name
}

func (n *BaseNode) String() string {
	return fmt.Sprintf("%s/%s", n.Network(), n.Name())
}

// Network returns the network to which this node belongs
func (n *BaseNode) Network() *Network {
	return n.network
}

func (n *BaseNode) Interface(name string) *Interface {
	for _, i := range n.Interfaces {
		if i.Name == name {
			return i
		}
	}

	return nil
}
