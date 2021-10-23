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

	n.Nodes[name] = sw // TODO: quirk to get n.UpdateHostsFile() working

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
			opt.Apply(sw)
		case BridgeOption:
			opt.Apply(br)
		case LinkOption:
			opt.Apply(&br.LinkAttrs)
		}
	}

	if err := nl.LinkAdd(br); err != nil {
		return nil, fmt.Errorf("failed to add bridge interface: %w", err)
	}

	log.WithFields(log.Fields{
		"node": name,
		"intf": br.LinkAttrs.Name,
	}).Infof("Adding new Linux bridge")

	if err := sw.Handle.LinkSetUp(br); err != nil {
		return nil, fmt.Errorf("failed to bring bridge up: %w", err)
	}

	return sw, nil
}

// ConfigurePort attaches an existing interface to a bridge port
func (sw *Switch) ConfigurePort(p Port) error {
	log.WithField("intf", p).Info("Connecting port to bridge master")
	br, err := sw.Handle.LinkByName(bridgeInterfaceName)
	if err != nil {
		return fmt.Errorf("failed to find bridge intf: %s", err)
	}

	l, err := sw.Handle.LinkByName(p.Name)
	if err != nil {
		return fmt.Errorf("failed to find new bridge port intf: %s", err)
	}

	// Attach port to bridge
	if err := sw.Handle.LinkSetMaster(l, br); err != nil {
		return err
	}

	// Bringing port up
	return sw.BaseNode.ConfigurePort(p)
}

func (sw Switch) Name() string {
	return sw.BaseNode.name
}
