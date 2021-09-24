package gont

import (
	"io"
	"net"
	"os/exec"
)

type Node struct {
	Name string

	NS      Namespace
	Network *Network
}

func (n *Network) AddNode(name string) (*Node, error) {
	node := &Node{
		Name:    name,
		Network: n,
	}

	ns, err := n.CreateNamespace(name)
	if err != nil {
		return nil, err
	}

	node.NS = ns

	n.Nodes = append(n.Nodes, node)

	return node, nil
}

func (n *Node) Run(cmd ...string) ([]byte, *exec.Cmd, error) {
	return n.Network.RunNS(n.NS, cmd...)
}

func (n *Node) RunAsync(cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	return n.Network.RunNSAsync(n.NS, cmd...)
}

func (n *Node) GoRunAsync(cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	return n.Network.GoRunNSAsync(n.NS, cmd...)
}

func (n *Node) ConfigureInterface(i *Interface) error {
	net := net.IPNet{
		IP:   i.Address,
		Mask: i.Mask,
	}

	if _, _, err := n.Run("ip", "address", "add", "dev", i.Name, net.String()); err != nil {
		return err
	}

	_, _, err := n.Run("ip", "link", "set", "dev", i.Name, "up")
	return err
}

func (n *Node) WriteProcFS(path, value string) error {
	// Parts of procfs are network namespaced.
	// We therefore need to write to it from a process which
	// itself runs inside the desired netns.
	_, _, err := n.Run("sh", "-c", "echo \""+value+"\" > "+path)
	return err
}

func (n *Node) EnableForwarding() error {
	if err := n.WriteProcFS("/proc/sys/net/ipv4/conf/all/forwarding", "1"); err != nil {
		return err
	}

	return n.WriteProcFS("/proc/sys/net/ipv6/conf/all/forwarding", "1")
}
