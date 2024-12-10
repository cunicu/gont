// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"

	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type SwitchOption interface {
	ApplySwitch(sw *Switch)
}

type BridgeOption interface {
	ApplyBridge(br *nl.Bridge)
}

// Switch is an abstraction for a Linux virtual bridge
type Switch struct {
	*BaseNode
}

// Options

func (sw *Switch) ApplyInterface(i *Interface) {
	i.Node = sw
}

// AddSwitch adds a new Linux virtual bridge in a dedicated namespace
func (n *Network) AddSwitch(name string, opts ...Option) (*Switch, error) {
	node, err := n.AddNode(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	sw := &Switch{
		BaseNode: node,
	}

	n.Register(sw)

	br := &nl.Bridge{
		LinkAttrs: nl.LinkAttrs{
			Name:      bridgeInterfaceName,
			Namespace: nl.NsFd(sw.NsHandle),
		},
	}

	// Apply options
	for _, opt := range opts {
		switch opt := opt.(type) {
		case SwitchOption:
			opt.ApplySwitch(sw)
		case BridgeOption:
			opt.ApplyBridge(br)
		case LinkOption:
			opt.ApplyLink(&br.LinkAttrs)
		}
	}

	if err := nl.LinkAdd(br); err != nil {
		return nil, fmt.Errorf("failed to add bridge interface: %w", err)
	}

	n.logger.Info("Adding new Linux bridge",
		zap.Any("node", sw),
		zap.String("intf", br.LinkAttrs.Name),
	)

	if err := sw.nlHandle.LinkSetUp(br); err != nil {
		return nil, fmt.Errorf("failed to bring bridge up: %w", err)
	}

	// Connect host to switch interfaces
	for _, intf := range sw.Interfaces {
		peerDev := fmt.Sprintf("veth-%s", name)

		left := intf
		left.Node = sw

		right := &Interface{
			Name: peerDev,
			Node: intf.Node,
		}

		if err := n.AddLink(left, right); err != nil {
			return nil, fmt.Errorf("failed to add link: %w", err)
		}
	}

	return sw, nil
}

// ConfigureInterface attaches an existing interface to a bridge interface
func (sw *Switch) ConfigureInterface(i *Interface) error {
	sw.logger.Info("Connecting interface to bridge master", zap.Any("intf", i))
	br, err := sw.nlHandle.LinkByName(bridgeInterfaceName)
	if err != nil {
		return fmt.Errorf("failed to find bridge interface: %w", err)
	}

	l, err := sw.nlHandle.LinkByName(i.Name)
	if err != nil {
		return fmt.Errorf("failed to find new bridge interface: %w", err)
	}

	// Attach interface to bridge
	if err := sw.nlHandle.LinkSetMaster(l, br); err != nil {
		return err
	}

	return sw.BaseNode.ConfigureInterface(i)
}
