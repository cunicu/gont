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

	nft "github.com/google/nftables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
)

type NetworkOption interface {
	ApplyNetwork(n *Network)
}

type Network struct {
	Name string

	nodes     map[string]Node
	nodesLock sync.RWMutex

	hostsFileLock sync.Mutex

	HostNode *Host
	VarPath  string
	TmpPath  string

	// Options
	Persistent    bool
	NSPrefix      string
	Captures      []*Capture
	Tracer        *Tracer
	RedirectToLog bool

	keyLogPipes []*os.File
	logger      *zap.Logger
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

func NewNetwork(name string, opts ...NetworkOption) (*Network, error) {
	if err := CheckCaps(); err != nil {
		return nil, err
	}

	if name == "" {
		name = GenerateNetworkName()
	}

	varPath := filepath.Join(baseVarDir, name)
	tmpPath := filepath.Join(baseTmpDir, name)

	n := &Network{
		Name:      name,
		VarPath:   varPath,
		TmpPath:   tmpPath,
		nodes:     map[string]Node{},
		nodesLock: sync.RWMutex{},
		NSPrefix:  "gont-",
		Captures:  []*Capture{},
		logger:    zap.L().Named("network").With(zap.String("network", name)),
	}

	// Apply network specific options
	for _, opt := range opts {
		opt.ApplyNetwork(n)
	}

	if stat, err := os.Stat(varPath); err == nil && stat.IsDir() {
		return nil, syscall.EEXIST
	}

	for _, path := range []string{"files", "nodes"} {
		path = filepath.Join(varPath, path)
		if err := os.MkdirAll(path, 0o644); err != nil {
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

func (n *Network) Nodes() []Node {
	n.nodesLock.RLock()
	defer n.nodesLock.RUnlock()

	nodes := []Node{}

	for _, node := range n.nodes {
		if host, ok := node.(*Host); ok {
			nodes = append(nodes, host)
		}
	}

	return nodes
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

func (n *Network) Teardown() error {
	n.nodesLock.Lock()
	defer n.nodesLock.Unlock()

	for name, node := range n.nodes {
		if err := node.Teardown(); err != nil {
			return err
		}

		delete(n.nodes, name)
	}

	if n.VarPath != "" {
		if err := os.RemoveAll(n.VarPath); err != nil {
			return fmt.Errorf("failed to delete network var dir: %w", err)
		}
	}

	if n.TmpPath != "" {
		if err := os.RemoveAll(n.VarPath); err != nil {
			return fmt.Errorf("failed to delete network tmp dir: %w", err)
		}
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
