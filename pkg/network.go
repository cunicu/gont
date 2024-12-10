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
	"slices"
	"strings"
	"sync"
	"syscall"

	"cunicu.li/gont/v2/internal/utils"
	"go.uber.org/zap"
)

var errNoKeyLogs = errors.New("no captures with keylogs")

type NetworkOption interface {
	ApplyNetwork(n *Network)
}

type Network struct {
	*CGroup

	Name    string
	VarPath string
	TmpPath string // For storing temporart Go build artifacts (see RunGo())

	nodes     map[string]Node
	nodesLock sync.RWMutex

	hostsFileLock sync.Mutex

	// Options
	Persistent    bool
	Captures      []*Capture
	Debugger      *Debugger
	Tracer        *Tracer
	RedirectToLog bool
	Slice         string

	keyLogPipes []*os.File
	logger      *zap.Logger
}

func NewNetwork(name string, opts ...Option) (n *Network, err error) {
	if err := CheckCaps(); err != nil {
		return nil, err
	}

	if name == "" {
		name = GenerateNetworkName()
	} else {
		if strings.Contains(name, "-") {
			return nil, fmt.Errorf("%w: malformed name: %s", ErrInvalidName, name)
		}

		if slices.Contains(NetworkNames(), name) {
			return nil, fmt.Errorf("%w: network already exists: %s", ErrInvalidName, name)
		}
	}

	varPath := filepath.Join(baseVarDir, name)
	tmpPath := filepath.Join(baseTmpDir, name)

	n = &Network{
		Name:    name,
		VarPath: varPath,
		TmpPath: tmpPath,
		Slice:   fmt.Sprintf("gont-%s", name),
		nodes:   map[string]Node{},
		logger:  zap.L().Named("network").With(zap.String("network", name)),
	}

	// Apply network specific options
	for _, opt := range slices.Concat(GlobalOptions, opts) {
		if opt, ok := opt.(NetworkOption); ok {
			opt.ApplyNetwork(n)
		}
	}

	// Setup directories
	if stat, err := os.Stat(varPath); err == nil && stat.IsDir() {
		return nil, syscall.EEXIST
	}

	for _, path := range []string{"files", "nodes"} {
		path = filepath.Join(varPath, path)
		if err := os.MkdirAll(path, 0o644); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Setup CGroup slice
	if n.CGroup, err = NewCGroup(nil, "slice", n.Slice, opts...); err != nil {
		return nil, fmt.Errorf("failed to create cgroup: %w", err)
	}

	if err := n.CGroup.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cgroup: %w", err)
	}

	if err := os.Symlink(
		filepath.Join(cgroupDir, "gont.slice", n.Slice+".slice"),
		filepath.Join(n.VarPath, "cgroup"),
	); err != nil {
		return nil, fmt.Errorf("failed to link cgroup: %w", err)
	}

	// Setup files
	if err := n.generateHostsFile(); err != nil {
		return nil, fmt.Errorf("failed to update hosts file: %w", err)
	}

	if err := n.generateConfigFiles(); err != nil {
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

	for name, node := range n.nodes {
		if err := node.Teardown(); err != nil {
			return fmt.Errorf("failed to teardown node %s: %w", name, err)
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

// MustClose closes the network like Close() but panics if an error occurs.
func (n *Network) MustClose() {
	if err := n.Close(); err != nil {
		panic(err)
	}
}

func (n *Network) Register(m Node) {
	n.nodesLock.Lock()
	defer n.nodesLock.Unlock()

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
		return nil, errNoKeyLogs
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
