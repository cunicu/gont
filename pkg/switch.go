package gont

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
)

// Switch is an abstraction for a Linux virtual bridge
type Switch struct {
	BaseNode

	Ports []Port
}

// Options

func (sw Switch) Apply(p *Port) {
	p.Node = &sw
}

// Getter

func (sw *Switch) Base() *BaseNode {
	return &sw.BaseNode
}

// AddSwitch adds a new Linux virtual bridge in a dedicated namespace
func (n *Network) AddSwitch(name string, opts ...Option) (*Switch, error) {
	node, err := n.AddNode(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	sw := &Switch{
		BaseNode: *node,
	}

	// Apply switch options
	for _, opt := range opts {
		if swopt, ok := opt.(SwitchOption); ok {
			swopt.Apply(sw)
		}
	}

	link := &nl.Bridge{
		LinkAttrs: nl.LinkAttrs{
			Name:      bridgeInterfaceName,
			Namespace: nl.NsFd(sw.NsHandle),
		},
	}
	if err := nl.LinkAdd(link); err != nil {
		return nil, fmt.Errorf("failed to add bridge interface: %w", err)
	}

	log.WithFields(log.Fields{
		"node": name,
		"intf": link.LinkAttrs.Name,
	}).Infof("Adding new Linux bridge")

	if err := sw.RunFunc(func() error {
		return nl.LinkSetUp(link)
	}); err != nil {
		return nil, fmt.Errorf("failed to bring bridge up: %w", err)
	}

	return sw, nil
}

// ConfigurePort attaches an existing interface to a bridge port
func (sw *Switch) ConfigurePort(p Port) error {
	log.WithField("intf", p).Info("Connecting port to bridge master")
	if err := sw.RunFunc(func() error {
		br, err := nl.LinkByName(bridgeInterfaceName)
		if err != nil {
			return fmt.Errorf("failed to find bridge intf: %s", err)
		}

		l, err := nl.LinkByName(p.Name)
		if err != nil {
			return fmt.Errorf("failed to find new bridge port intf: %s", err)
		}

		// Attach port to bridge
		if err := nl.LinkSetMaster(l, br); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Bringing port up
	return sw.BaseNode.ConfigurePort(p)
}

func (sw Switch) Name() string {
	return sw.BaseNode.name
}
