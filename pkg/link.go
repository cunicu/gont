package gont

import (
	"errors"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
)

func (n *Network) AddLink(l, r Endpoint, opts ...Option) error {
	lPort := l.port()
	rPort := r.port()

	if lPort.Node == nil || rPort.Node == nil {
		return errors.New("cant establish link between interfaces without node")
	}

	if lPort.Node == rPort.Node {
		return errors.New("failed to link the node with itself")
	}

	log.WithFields(log.Fields{
		"left":  lPort,
		"right": rPort,
	}).Info("Adding new veth pair")

	// Create Veth pair
	veth := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{
			Name:      lPort.Name,
			Namespace: nl.NsFd(lPort.Node.Base().NsHandle),
			TxQLen:    -1,
		},
		PeerName:      rPort.Name,
		PeerNamespace: nl.NsFd(rPort.Node.Base().NsHandle),
	}
	if err := nl.LinkAdd(veth); err != nil {
		return err
	}

	// Configuring endpoints (link state, attaching to bridge, adding addresses)
	for _, e := range []Endpoint{l, r} {
		if err := e.Configure(); err != nil {
			return err
		}
	}

	return nil
}
