package gont

type Router struct {
	*Host
}

func (h *Router) Apply(i *Interface) {
	i.Node = h
}

func (n *Network) AddRouter(name string, opts ...Option) (*Router, error) {
	host, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	if err := host.EnableForwarding(); err != nil {
		return nil, err
	}

	rtr := &Router{
		Host: host,
	}

	n.Register(rtr)

	return rtr, nil
}
