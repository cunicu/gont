package gont

type Router struct {
	Host
}

func (n *Network) AddRouter(name string, opts ...Option) (*Router, error) {
	host, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	if err := host.EnableForwarding(); err != nil {
		return nil, err
	}

	return &Router{
		Host: *host,
	}, nil
}
