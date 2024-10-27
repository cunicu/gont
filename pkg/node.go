// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Node interface {
	Close() error
	Teardown() error

	// Getters
	Name() string
	String() string
	Network() *Network
	Interface(name string) *Interface

	ConfigureInterface(i *Interface) error
}

type NamespacedNode interface {
	Node

	RunFunc(cb Callback) error
	NetNSHandle() netns.NsHandle
	NetlinkHandle() *nl.Handle
}
