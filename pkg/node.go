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
	NetNSHandle() netns.NsHandle
	NetlinkHandle() *nl.Handle

	RunFunc(cb Callback) error

	ConfigureInterface(i *Interface) error
}
