// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"cunicu.li/gont/v2/internal/utils"
	nft "github.com/google/nftables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
)

type NetworkOption interface {
	ApplyNetwork(n *Network)
}

type Network struct {
	*CGroup

	Name string

	nodes     map[string]Node
	nodesLock sync.RWMutex

	hostsFileLock sync.Mutex

	HostNode *Host
	VarPath  string
	TmpPath  string // For Go builds (see RunGo())

	// Options
	Persistent    bool
	Captures      []*Capture
	Debugger      *Debugger
	Tracer        *Tracer
	RedirectToLog bool

	keyLogPipes []*os.File
	logger      *zap.Logger
}

func HostNode(n *Network) (h *Host) {
	baseNs, err := netns.Get()
	if err != nil {
		return nil
	}

	baseHandle, err := netlink.NewHandle()
	if err != nil {
		return nil
	}

	h = &Host{
		BaseNode: &BaseNode{
			name:       "host",
			isHostNode: true,
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

	cgroupName := fmt.Sprintf("gont-%s-%s", n.Name, h.name)
	h.CGroup, err = NewCGroup(n.sdConn, "slice", cgroupName)
	if err != nil {
		return nil
	}

	if err := h.CGroup.Start(); err != nil {
		return nil
	}

	return h
}

func NewNetwork(name string, opts ...Option) (n *Network, err error) {
	if err := CheckCaps(); err != nil {
		return nil, err
	}

	if name == "" {
		name = GenerateNetworkName()
	}

	varPath := filepath.Join(baseVarDir, name)
	tmpPath := filepath.Join(baseTmpDir, name)

	n = &Network{
		Name:     name,
		VarPath:  varPath,
		TmpPath:  tmpPath,
		nodes:    map[string]Node{},
		Captures: []*Capture{},
		logger:   zap.L().Named("network").With(zap.String("network", name)),
	}

	// Apply network specific options
	for _, opt := range opts {
		switch opt := opt.(type) {
		case NetworkOption:
			opt.ApplyNetwork(n)
		}
	}

	// Setup CGroup slice
	cgroupName := fmt.Sprintf("gont-%s", name)
	if n.CGroup, err = NewCGroup(nil, "slice", cgroupName, opts...); err != nil {
		return nil, fmt.Errorf("failed to create CGroup slice: %w", err)
	}

	// Setup files
	if stat, err := os.Stat(varPath); err == nil && stat.IsDir() {
		return nil, syscall.EEXIST
	}

	for _, path := range []string{"files", "nodes"} {
		path = filepath.Join(varPath, path)
		if err := os.MkdirAll(path, 0o644); err != nil {
			return nil, err
		}
	}

	if n.HostNode = HostNode(n); n.HostNode == nil {
		return nil, errors.New("failed to create host node")
	}

	if err := n.generateHostsFile(); err != nil {
		return nil, fmt.Errorf("failed to update hosts file: %w", err)
	}

	if err := n.generateConfigFiles(); err != nil {
		return nil, fmt.Errorf("failed to generate configuration files: %w", err)
	}

	if err := n.CGroup.Start(); err != nil {
		return nil, fmt.Errorf("failed to start CGroup slice: %w", err)
	}

	if err := os.Symlink(
		filepath.Join(cgroupDir, "gont.slice", cgroupName+".slice"),
		filepath.Join(n.VarPath, "cgroup"),
	); err != nil {
		return nil, fmt.Errorf("failed to link cgroup: %w", err)
	}

	n.logger.Info("Created new network")

	return n, nil
}

// Getter

func (n *Network) String() string {
	return n.Name
}

func (n *Network) Nodes() []Node {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	return utils.MapValues(n.nodes)
}

func (n *Network) Hosts() []*Host {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	hosts := []*Host{}

	for _, node := range n.nodes {
		if host, ok := node.(*Host); ok {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

func (n *Network) Switches() []*Switch {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	switches := []*Switch{}

	for _, node := range n.nodes {
		if sw, ok := node.(*Switch); ok {
			switches = append(switches, sw)
		}
	}

	return switches
}

func (n *Network) Routers() []*Router {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	routers := []*Router{}

	for _, node := range n.nodes {
		if router, ok := node.(*Router); ok {
			routers = append(routers, router)
		}
	}

	return routers
}

// Iterators

func (n *Network) ForEachHost(cb func(h *Host)) {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	for _, node := range n.nodes {
		if host, ok := node.(*Host); ok {
			cb(host)
		}
	}
}

func (n *Network) Teardown() error {
	n.nodesLock.Lock()
	defer n.nodesLock.Unlock()

	if err := n.HostNode.Teardown(); err != nil {
		return err
	}

	for name, node := range n.nodes {
		if err := node.Teardown(); err != nil {
			return err
		}

		delete(n.nodes, name)
	}

	if n.VarPath != "" {
		if err := os.RemoveAll(n.VarPath); err != nil {
			return fmt.Errorf("failed to delete network directory: %w", err)
		}
	}

	if n.TmpPath != "" {
		if err := os.RemoveAll(n.TmpPath); err != nil {
			return fmt.Errorf("failed to delete temporary network directory: %w", err)
		}
	}

	if err := n.CGroup.Stop(); err != nil {
		return fmt.Errorf("failed to stop cgroup: %w", err)
	}

	return nil
}

func (n *Network) Close() error {
	if !n.Persistent {
		if err := n.Teardown(); err != nil {
			return err
		}
	}

	for name, node := range n.nodes {
		if err := node.Close(); err != nil {
			return fmt.Errorf("failed to close node '%s': %w", name, err)
		}

		delete(n.nodes, name)
	}

	for _, p := range n.keyLogPipes {
		if err := p.Close(); err != nil {
			return fmt.Errorf("failed to close keylog pipe: %w", err)
		}
	}

	for _, c := range n.Captures {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close packet capture: %w", err)
		}
	}

	if t := n.Tracer; t != nil {
		if err := t.Close(); err != nil {
			return fmt.Errorf("failed to close packet capture: %w", err)
		}
	}

	return nil
}

func (n *Network) Register(m Node) {
	n.nodesLock.Lock()
	defer n.nodesLock.Unlock()

	// TODO: Handle name collisions
	n.nodes[m.Name()] = m
}

func (n *Network) KeyLogPipe(secretsType uint32) (*os.File, error) {
	capturesWithKeys := []*Capture{}
	for _, c := range n.Captures {
		if c.LogKeys {
			capturesWithKeys = append(capturesWithKeys, c)
		}
	}

	if len(capturesWithKeys) == 0 {
		return nil, nil
	}

	rd, wr, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	go func() {
		b := &bytes.Buffer{}

		if _, err := io.Copy(b, rd); err != nil && !errors.Is(err, os.ErrClosed) {
			n.logger.Error("Failed to read key log data", zap.Error(err))
			return
		}

		for _, c := range capturesWithKeys {
			if err := c.writeDecryptionSecret(secretsType, b.Bytes()); err != nil {
				n.logger.Error("Failed to decryption secret", zap.Error(err))
			}
		}
	}()

	n.keyLogPipes = append(n.keyLogPipes, rd)

	return wr, nil
}
