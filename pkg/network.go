package gont

import (
	"errors"
	"fmt"
	"syscall"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"kernel.org/pub/linux/libs/security/libcap/cap"
)

type Network struct {
	Name     string
	Nodes    map[string]Node
	HostNode *Host
	BasePath string

	Persistent bool
	NSPrefix   string

	DefaultOptions Options
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
				Handle:   baseHandle,
			},
			network: n,
		},
	}
}

func NewNetwork(name string, opts ...Option) (*Network, error) {
	// Check for required permissions
	caps := cap.GetProc()
	if val, err := caps.GetFlag(cap.Effective, cap.SYS_ADMIN); err != nil {
		return nil, err
	} else if !val {
		return nil, errors.New("missing SYS_ADMIN capability")
	}

	if name == "" {
		name = GenerateNetworkName()
	}

	basePath := filepath.Join(varDir, name)

	n := &Network{
		Name:           name,
		BasePath:       basePath,
		Nodes:          map[string]Node{},
		DefaultOptions: opts,
		NSPrefix:       "gont-",
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

	log.Infof("Created new network: %s", n.Name)

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
	for _, node := range n.Nodes {
		if err := node.Teardown(); err != nil {
			return err
		}
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

	return nil
}

func (n *Network) Register(m Node) {
	n.Nodes[m.Name()] = m
}
