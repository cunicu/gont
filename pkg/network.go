package gont

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	"os"
	"path/filepath"

	nft "github.com/google/nftables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
)

type Network struct {
	Name string

	Nodes     map[string]Node
	NodesLock sync.RWMutex

	HostNode *Host
	BasePath string

	// Options
	Persistent bool
	NSPrefix   string
	Captures   []*Capture

	logger *zap.Logger
}

func HostNode(n *Network) *Host {
	baseNs, err := netns.Get()
	if err != nil {
		return nil
	}

	baseHandle, err := netlink.NewHandle()
	if err != nil {
		return nil
	}

	return &Host{
		BaseNode: &BaseNode{
			name: "host",
			Namespace: &Namespace{
				Name:     "base",
				NsHandle: baseNs,
				nlHandle: baseHandle,
				nftConn:  &nft.Conn{},
				logger:   zap.L().Named("namespace"),
			},
			network: n,
			logger:  zap.L().Named("host"),
		},
	}
}

func NewNetwork(name string, opts ...Option) (*Network, error) {
	if err := CheckCaps(); err != nil {
		return nil, err
	}

	if name == "" {
		name = GenerateNetworkName()
	}

	basePath := filepath.Join(varDir, name)

	n := &Network{
		Name:      name,
		BasePath:  basePath,
		Nodes:     map[string]Node{},
		NodesLock: sync.RWMutex{},
		NSPrefix:  "gont-",
		Captures:  []*Capture{},
		logger:    zap.L().Named("network").With(zap.String("network", name)),
	}

	// Apply network specific options
	for _, opt := range opts {
		if nopt, ok := opt.(NetworkOption); ok {
			nopt.Apply(n)
		}
	}

	if stat, err := os.Stat(basePath); err == nil && stat.IsDir() {
		return nil, syscall.EEXIST
	}

	for _, path := range []string{"files", "nodes"} {
		path = filepath.Join(basePath, path)
		if err := os.MkdirAll(path, 0644); err != nil {
			return nil, err
		}
	}

	n.HostNode = HostNode(n)
	if n.HostNode == nil {
		return nil, errors.New("failed to create host node")
	}

	if err := n.GenerateHostsFile(); err != nil {
		return nil, fmt.Errorf("failed to update hosts file: %w", err)
	}

	if err := n.GenerateConfigFiles(); err != nil {
		return nil, fmt.Errorf("failed to generate configuration files: %w", err)
	}

	n.logger.Info("Created new network")

	return n, nil
}

// Getter

func (n *Network) String() string {
	return n.Name
}

func (n *Network) Hosts() []*Host {
	hosts := []*Host{}

	for _, node := range n.Nodes {
		if host, ok := node.(*Host); ok {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

func (n *Network) Switches() []*Switch {
	switches := []*Switch{}

	for _, node := range n.Nodes {
		if sw, ok := node.(*Switch); ok {
			switches = append(switches, sw)
		}
	}

	return switches
}

func (n *Network) Routers() []*Router {
	routers := []*Router{}

	for _, node := range n.Nodes {
		if router, ok := node.(*Router); ok {
			routers = append(routers, router)
		}
	}

	return routers
}

func (n *Network) Teardown() error {
	n.NodesLock.Lock()
	defer n.NodesLock.Unlock()

	for name, node := range n.Nodes {
		if err := node.Teardown(); err != nil {
			return err
		}

		delete(n.Nodes, name)
	}

	if n.BasePath != "" {
		os.RemoveAll(n.BasePath)
	}

	return nil
}

func (n *Network) Close() error {
	if !n.Persistent {
		if err := n.Teardown(); err != nil {
			return err
		}
	}

	for name, node := range n.Nodes {
		if err := node.Close(); err != nil {
			return err
		}

		delete(n.Nodes, name)
	}

	for _, c := range n.Captures {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (n *Network) Register(m Node) {
	n.NodesLock.Lock()
	defer n.NodesLock.Unlock()

	// TODO handle name collisions

	n.Nodes[m.Name()] = m
}
