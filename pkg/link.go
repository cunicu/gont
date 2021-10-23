package gont

import (
	"errors"
	"fmt"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/stv0g/gont/internal/utils"
	nl "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func (n *Network) AddLink(l, r Endpoint, opts ...Option) error {
	var err error

	lPort := l.port()
	rPort := r.port()

	lNode := lPort.Node.Base()
	rNode := rPort.Node.Base()

	if len(lPort.Name) > syscall.IFNAMSIZ-1 || len(rPort.Name) > syscall.IFNAMSIZ-1 {
		return fmt.Errorf("interface names are too long. max_len=%d", syscall.IFNAMSIZ-1)
	}

	if lPort.Node == rPort.Node {
		return errors.New("failed to link the node with itself")
	}

	if lPort.Node == nil || rPort.Node == nil {
		return errors.New("cant establish link between interfaces without node")
	}

	log.WithFields(log.Fields{
		"left":  lPort,
		"right": rPort,
	}).Info("Adding new veth pair")

	// Create Veth pair

	// For some unknown reason we cant create the peer interface
	// directly in the target namespace. So we create it in the same
	// and move + rename it later.

	// We also cant create the interface from the host namespace
	// as this leads to race conditions due to dupplicate device names
	veth := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{
			Name:   utils.RandStringRunes(unix.IFNAMSIZ - 1), // temporary name
			TxQLen: -1,
		},
		PeerName: rPort.Name,
	}

	// Apply options
	for _, opt := range opts {
		switch opt := opt.(type) {
		case VethOption:
			opt.Apply(veth)
		case LinkOption:
			opt.Apply(&veth.LinkAttrs)
		}
	}

	if err = lNode.Handle.LinkAdd(veth); err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	rLink, err := lNode.Handle.LinkByName(rPort.Name)
	if err != nil {
		return err
	}

	if err := lNode.Handle.LinkSetNsFd(rLink, int(rNode.NsHandle)); err != nil {
		return err
	}

	if err := lNode.Handle.LinkSetName(veth, lPort.Name); err != nil {
		return err
	}

	// Configuring endpoints (link state, attaching to bridge, adding addresses)
	for _, e := range []Endpoint{l, r} {
		if err := e.Configure(); err != nil {
			return fmt.Errorf("failed to configure endpoint: %w", err)
		}
	}

	return nil
}
