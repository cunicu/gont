package gont

import "net"

type Router struct {
	Node
}

func (n *Network) AddRouter(name string, gw net.IP, intfs ...*Interface) (*Router, error) {
	node, err := n.AddNode(name)
	if err != nil {
		return nil, err
	}

	rtr := &Router{
		Node: *node,
	}

	rtr.EnableForwarding()

	for _, intf := range intfs {
		peerDev := "br-" + rtr.Name + "-" + intf.Name

		if _, _, err = rtr.Run("ip", "link", "add", "name", intf.Name, "type", "veth", "peer", peerDev, "netns", intf.Switch.NS.String()); err != nil {
			return nil, err
		}
		if _, _, err = intf.Switch.Run("ip", "link", "set", "dev", peerDev, "master", "br0"); err != nil {
			return nil, err
		}
		if _, _, err = intf.Switch.Run("ip", "link", "set", "dev", peerDev, "up"); err != nil {
			return nil, err
		}

		rtr.ConfigureInterface(intf)

		n.AddToHostsFile(intf.Address, rtr.Name+"-"+intf.Name)
	}

	if gw != nil {
		if _, _, err = rtr.Run("ip", "route", "add", "default", "via", gw.String()); err != nil {
			return nil, err
		}
	}

	return rtr, nil
}

func (r *Router) Close() error {

	return nil
}
