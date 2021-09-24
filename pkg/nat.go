package gont

import "net"

type NAT struct {
	Router
}

func (n *Network) AddNAT(name string, gw net.IP, nb, sb *Interface) (*NAT, error) {
	r, err := n.AddRouter(name, gw, nb, sb)
	if err != nil {
		return nil, err
	}

	nat := &NAT{
		Router: *r,
	}

	if _, _, err = nat.Run("iptables", "-I", "FORWARD", "-i", sb.Name, "-d", sb.Network().String(), "-j", "DROP"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-A", "FORWARD", "-i", sb.Name, "-s", sb.Network().String(), "-j", "ACCEPT"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-A", "FORWARD", "-o", sb.Name, "-d", sb.Network().String(), "-j", "ACCEPT"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", sb.Network().String(), "!", "-d", sb.Network().String(), "-j", "MASQUERADE"); err != nil {
		return nil, err
	}

	return nat, nil
}

func (n *Network) AddHostNAT(name string, sb *Interface) (*NAT, error) {
	r, err := n.AddRouter(name, nil, sb)
	if err != nil {
		return nil, err
	}

	nat := &NAT{
		Router: *r,
	}

	if _, _, err = nat.Run("iptables", "-I", "FORWARD", "-i", sb.Name, "-d", sb.Network().String(), "-j", "DROP"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-A", "FORWARD", "-i", sb.Name, "-s", sb.Network().String(), "-j", "ACCEPT"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-A", "FORWARD", "-o", sb.Name, "-d", sb.Network().String(), "-j", "ACCEPT"); err != nil {
		return nil, err
	}
	if _, _, err = nat.Run("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", sb.Network().String(), "!", "-d", sb.Network().String(), "-j", "MASQUERADE"); err != nil {
		return nil, err
	}

	return nat, nil
}
