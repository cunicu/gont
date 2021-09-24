package gont

type Namespace struct {
	Name string
	NSID uint

	Network *Network
}

func (n Namespace) String() string {
	return n.Name
}

func (n Namespace) Close() error {
	_, _, err := n.Network.Run("ip", "netns", "delete", n.Name)
	return err
}
