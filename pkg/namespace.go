// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"

	nft "github.com/google/nftables"
	"github.com/vishvananda/netlink"
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

var ErrNameReserved = errors.New("name 'host' is reserved")

var hostNamespace *Namespace //nolint:gochecknoglobals

type Callback func() error

type Namespace struct {
	netns.NsHandle

	nlHandle *nl.Handle
	nftConn  *nft.Conn

	Name string

	logger *zap.Logger
}

// HostNamespace creates
func HostNamespace() (ns *Namespace, err error) {
	if hostNamespace != nil {
		return hostNamespace, nil
	}

	ns = &Namespace{
		Name:    "host",
		nftConn: &nft.Conn{},
		logger:  zap.L().Named("namespace").With(zap.String("ns", "host")),
	}

	if ns.NsHandle, err = netns.Get(); err != nil {
		return nil, fmt.Errorf("failed to get network namespace handle: %w", err)
	}

	if ns.nlHandle, err = netlink.NewHandle(); err != nil {
		return nil, fmt.Errorf("failed to create netlink handle: %w", err)
	}

	hostNamespace = ns

	return ns, nil
}

// NewNamespace creates a new named network namespace.
func NewNamespace(name string) (ns *Namespace, err error) {
	if name == "host" {
		return nil, ErrNameReserved
	}

	ns = &Namespace{
		Name:   name,
		logger: zap.L().Named("namespace").With(zap.String("ns", name)),
	}

	ns.logger.Info("Creating new namespace")

	// We lock the goroutine to an OS thread for the duration while we open the netlink sockets
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save fd to current network namespace
	curNetNs, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0o777)
	if err != nil {
		return nil, fmt.Errorf("failed to open network namespace: %w", err)
	}

	// Create new named namespace
	if ns.NsHandle, err = netns.NewNamed(ns.Name); err != nil {
		return nil, fmt.Errorf("failed to create new named network namespace: %w", err)
	}

	// Create a netlink socket handle while we are in the namespace
	if ns.nlHandle, err = nl.NewHandle(); err != nil {
		return nil, fmt.Errorf("failed to create netlink handle: %w", err)
	}

	// Setup nftables connection
	ns.nftConn = &nft.Conn{
		NetNS: int(ns.NsHandle),
	}

	// Restore original netns namespace
	if err = unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
		return nil, fmt.Errorf("failed to restore original network namepsace: %w", err)
	}

	return ns, nil
}

// Close releases the network namespace.
func (ns *Namespace) Close() error {
	if ns.IsHost() {
		return nil
	}

	if ns.NsHandle >= 0 {
		if err := netns.DeleteNamed(ns.Name); err != nil {
			return fmt.Errorf("failed to delete network namespace: %w", err)
		}

		ns.logger.Info("Deleted namespace")
	}

	return nil
}

// MustClose closes the namespace like Close() but panics if an error occurs.
func (n *Namespace) MustClose() {
	if err := n.Close(); err != nil {
		panic(err)
	}
}

// RunFunc runs a Go function within the namespace.
// Note, that Goroutines started from within the passed function
// are not guaranteed to run inside the same namespace!
// This function calls runtime.{Lock|Unlock}OSThread().
func (ns *Namespace) RunFunc(cb Callback) error {
	exit, err := ns.Enter()
	if err != nil {
		return fmt.Errorf("failed to enter namespace: %w", err)
	}
	defer exit()

	return cb()
}

// Enter locks the current Goroutine to an OS thread by calling runtime.LockOSThread().
// and afterwards attaches the calling Goroutines thread to the namespace.
// The returned function should be called to move the thread back to the original namespace
// and unlock the Goroutine from the OS thread.
func (ns *Namespace) Enter() (func(), error) {
	runtime.LockOSThread()

	// Save fd to current network namespace
	curNetNs, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0o777)
	if err != nil {
		return nil, err
	}

	// Switch to network namespace
	if err := unix.Setns(int(ns.NsHandle), syscall.CLONE_NEWNET); err != nil {
		return nil, err
	}

	// ns.logger.Debug("Entered namespace")

	return func() {
		// Restore original netns namespace
		if err := unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
			panic(fmt.Errorf("failed to switch back to original netns: %w", err))
		}

		if err := syscall.Close(curNetNs); err != nil {
			panic(fmt.Errorf("failed to close netns descriptor: %w", err))
		}

		// ns.logger.Debug("Left namespace")

		runtime.UnlockOSThread()
	}, nil
}

// IsHost returns true if the namespace is representing the hosts default network namespace.
func (ns *Namespace) IsHost() bool {
	return ns.Name == "host"
}
