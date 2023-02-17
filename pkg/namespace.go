// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"runtime"
	"syscall"

	nft "github.com/google/nftables"
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"

	"golang.org/x/sys/unix"
)

type Callback func() error

type Namespace struct {
	netns.NsHandle

	nlHandle *nl.Handle
	nftConn  *nft.Conn

	Name string

	logger *zap.Logger
}

func NewNamespace(name string) (*Namespace, error) {
	var err error

	ns := &Namespace{
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
		return nil, err
	}

	// Create new named namespace
	if ns.NsHandle, err = netns.NewNamed(ns.Name); err != nil {
		return nil, err
	}

	// Create a netlink socket handle while we are in the namespace
	if ns.nlHandle, err = nl.NewHandle(); err != nil {
		return nil, err
	}

	// nftables connection
	ns.nftConn = &nft.Conn{
		NetNS: int(ns.NsHandle),
	}

	// Restore original netns namespace
	return ns, unix.Setns(curNetNs, syscall.CLONE_NEWNET)
}

func (ns *Namespace) Close() error {
	if ns.NsHandle >= 0 {
		if err := netns.DeleteNamed(ns.Name); err != nil {
			return err
		}

		ns.logger.Info("Deleted namespace")
	}

	return nil
}

func (ns *Namespace) RunFunc(cb Callback) error {
	exit, err := ns.Enter()
	if err != nil {
		return fmt.Errorf("failed to enter namespace: %w", err)
	}
	defer exit()

	return cb()
}

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
