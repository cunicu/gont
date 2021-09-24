package gont

type Switch struct {
	Node
}

func (net *Network) AddSwitch(name string) (*Switch, error) {
	n, err := net.AddNode(name)
	if err != nil {
		return nil, err
	}

	sw := &Switch{
		Node: *n,
	}

	if _, _, err = sw.Run("ip", "link", "add", "name", "br0", "type", "bridge"); err != nil {
		return nil, err
	}
	if _, _, err = sw.Run("ip", "link", "set", "dev", "br0", "up"); err != nil {
		return nil, err
	}

	return sw, nil
}
