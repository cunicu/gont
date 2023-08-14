// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"fmt"
	"syscall"

	"cunicu.li/gont/v2/internal/utils"
	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type VethOption interface {
	ApplyVeth(ve *nl.Veth)
}

type LinkOption interface {
	ApplyLink(a *nl.LinkAttrs)
}

func (n *Network) AddLink(l, r *Interface, opts ...Option) error {
	var err error

	if len(l.Name) > syscall.IFNAMSIZ-1 || len(r.Name) > syscall.IFNAMSIZ-1 {
		return fmt.Errorf("interface names are too long. max_len=%d", syscall.IFNAMSIZ-1)
	}

	if l.Node == nil || r.Node == nil {
		return errors.New("cant establish link between interfaces without node")
	}

	if l.Node == r.Node {
		return errors.New("failed to link the node with itself")
	}

	if l.Node.Network() != r.Node.Network() {
		return errors.New("nodes are belonging to different networks")
	}

	// Create Veth pair
	n.logger.Info("Adding new veth pair",
		zap.Any("left", l),
		zap.Any("right", r),
	)

	// For some unknown reason we cant create the peer interface
	// directly in the target namespace. So we create it in the same
	// and move + rename it later.

	// We also cant create the interface from the host namespace
	// as this leads to race conditions due to duplicate device names
	veth := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{
			Name:   utils.RandStringRunes(unix.IFNAMSIZ - 1), // temporary name
			TxQLen: -1,
		},
		PeerName: r.Name,
	}

	// Apply options
	for _, opt := range opts {
		switch opt := opt.(type) {
		case VethOption:
			opt.ApplyVeth(veth)
		}
	}

	lHandle := l.Node.NetlinkHandle()
	rHandle := r.Node.NetlinkHandle()

	// Create veth pair
	if err = lHandle.LinkAdd(veth); err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	rLink, err := lHandle.LinkByName(r.Name)
	if err != nil {
		return fmt.Errorf("failed to find interface %s: %w", r.Name, err)
	}

	// Move one side into the target netns
	if err := lHandle.LinkSetNsFd(rLink, int(r.Node.NetNSHandle())); err != nil {
		lHandle.LinkDel(veth) //nolint:errcheck
		return fmt.Errorf("failed to move interface to namespace: %w", err)
	}

	// Rename veth
	if err := lHandle.LinkSetName(veth, l.Name); err != nil {
		lHandle.LinkDel(veth) //nolint:errcheck
		return fmt.Errorf("failed to rename interface: %w", err)
	}

	// Update link structures
	if l.Link, err = lHandle.LinkByName(l.Name); err != nil {
		return fmt.Errorf("failed to find interface %s: %w", l.Name, err)
	}

	if r.Link, err = rHandle.LinkByName(r.Name); err != nil {
		return fmt.Errorf("failed to find interface %s: %w", r.Name, err)
	}

	// Configure interface (link state, attaching to bridge, adding addresses)
	for _, i := range []*Interface{l, r} {
		if err := i.Node.ConfigureInterface(i); err != nil {
			return fmt.Errorf("failed to configure endpoint: %w", err)
		}
	}

	return nil
}
