// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"cunicu.li/gont/v2/internal/utils"
	nft "github.com/google/nftables"
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type BaseNodeOption interface {
	ApplyBaseNode(n *BaseNode)
}

type BaseNode struct {
	*Namespace

	isHostNode bool
	network    *Network
	name       string

	BasePath string

	Interfaces []*Interface

	// Options
	ConfiguredInterfaces    []*Interface
	Tracer                  *Tracer
	Debugger                *Debugger
	ExistingNamespace       string
	ExistingDockerContainer string
	RedirectToLog           bool
	EmptyDirs               []string
	Captures                []*Capture

	logger *zap.Logger
}

func (n *Network) AddNode(name string, opts ...Option) (*BaseNode, error) {
	var err error

	basePath := filepath.Join(n.VarPath, "nodes", name)
	for _, path := range []string{"ns", "files"} {
		path = filepath.Join(basePath, path)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
	}

	node := &BaseNode{
		name:     name,
		network:  n,
		BasePath: basePath,
		logger:   zap.L().Named("node").With(zap.String("node", name)),
	}

	node.logger.Info("Adding new node")

	for _, opt := range opts {
		if nOpt, ok := opt.(BaseNodeOption); ok {
			nOpt.ApplyBaseNode(node)
		}
	}

	// Create mount point directories
	for _, ed := range node.EmptyDirs {
		path := filepath.Join(basePath, "files", ed)

		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Create non-existing empty dirs
		// TODO: Should we cleanup in Close()?
		if _, err := os.Stat(ed); err != nil && errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(ed, 0o755); err != nil {
				return nil, fmt.Errorf("failed to create directory '%s': %w", ed, err)
			}
		}

		// Directories containing a hidden .mount file will be bind mounted
		// as a whole rather than just the files it contains.
		hfn := filepath.Join(path, ".mount")
		if err := utils.Touch(hfn); err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
	}

	switch {
	case node.ExistingNamespace != "":
		// Use an existing namespace created by "ip netns add"
		nsh, err := netns.GetFromName(node.ExistingNamespace)
		if err != nil {
			return nil, fmt.Errorf("failed to find existing network namespace %s: %w", node.ExistingNamespace, err)
		}

		node.Namespace = &Namespace{
			Name:     node.ExistingNamespace,
			NsHandle: nsh,
		}

	case node.ExistingDockerContainer != "":
		// Use an existing net namespace from a Docker container
		nsh, err := netns.GetFromDocker(node.ExistingDockerContainer)
		if err != nil {
			return nil, fmt.Errorf("failed to find existing docker container %s: %w", node.ExistingNamespace, err)
		}

		node.Namespace = &Namespace{
			Name:     node.ExistingDockerContainer,
			NsHandle: nsh,
		}

	default:
		// Create a new network namespace
		nsName := fmt.Sprintf("%s%s-%s", n.NSPrefix, n.Name, name)
		if node.Namespace, err = NewNamespace(nsName); err != nil {
			return nil, err
		}
	}

	if node.nlHandle == nil {
		node.nlHandle, err = nl.NewHandleAt(node.NsHandle)
		if err != nil {
			return nil, err
		}
	}

	src := fmt.Sprintf("/proc/self/fd/%d", int(node.NsHandle))
	dst := filepath.Join(basePath, "ns", "net")
	if err := utils.Touch(dst); err != nil {
		return nil, err
	}
	if err := unix.Mount(src, dst, "", syscall.MS_BIND, ""); err != nil {
		return nil, fmt.Errorf("failed to bind mount netns fd: %s", err)
	}

	n.Register(node)

	return node, nil
}

// Getter

func (n *BaseNode) NetNSHandle() netns.NsHandle {
	return n.NsHandle
}

func (n *BaseNode) NetlinkHandle() *nl.Handle {
	return n.nlHandle
}

func (n *BaseNode) NftConn() *nft.Conn {
	return n.nftConn
}

func (n *BaseNode) Name() string {
	return n.name
}

func (n *BaseNode) String() string {
	return fmt.Sprintf("%s/%s", n.Network(), n.Name())
}

// Network returns the network to which this node belongs
func (n *BaseNode) Network() *Network {
	return n.network
}

func (n *BaseNode) Interface(name string) *Interface {
	for _, i := range n.Interfaces {
		if i.Name == name {
			return i
		}
	}

	return nil
}

func (n *BaseNode) ConfigureInterface(i *Interface) error {
	logger := n.logger.With(zap.Any("intf", i))
	logger.Info("Configuring interface")

	// Set MTU
	if i.LinkAttrs.MTU != 0 {
		logger.Info("Setting interface MTU",
			zap.Int("mtu", i.LinkAttrs.MTU),
		)
		if err := n.nlHandle.LinkSetMTU(i.Link, i.LinkAttrs.MTU); err != nil {
			return err
		}
	}

	// Set L2 (MAC) address
	if i.LinkAttrs.HardwareAddr != nil {
		logger.Info("Setting interface MAC address",
			zap.Any("mac", i.LinkAttrs.HardwareAddr),
		)
		if err := n.nlHandle.LinkSetHardwareAddr(i.Link, i.LinkAttrs.HardwareAddr); err != nil {
			return err
		}
	}

	// Set transmit queue length
	if i.LinkAttrs.TxQLen > 0 {
		logger.Info("Setting interface transmit queue length",
			zap.Int("txqlen", i.LinkAttrs.TxQLen),
		)
		if err := n.nlHandle.LinkSetTxQLen(i.Link, i.LinkAttrs.TxQLen); err != nil {
			return err
		}
	}

	// Set interface group
	if i.LinkAttrs.Group != 0 {
		logger.Info("Setting interface group",
			zap.Uint32("group", i.LinkAttrs.Group),
		)
		if err := n.nlHandle.LinkSetGroup(i.Link, int(i.LinkAttrs.Group)); err != nil {
			return err
		}
	}

	// Setup netem Qdisc
	var pHandle uint32 = nl.HANDLE_ROOT
	if i.Flags&WithQdiscNetem != 0 {
		attr := nl.QdiscAttrs{
			LinkIndex: i.Link.Attrs().Index,
			Handle:    nl.MakeHandle(1, 0),
			Parent:    pHandle,
		}

		netem := nl.NewNetem(attr, i.Netem)

		logger.Info("Adding Netem qdisc to interface")
		if err := n.nlHandle.QdiscAdd(netem); err != nil {
			return err
		}

		pHandle = netem.Handle
	}

	// Setup tbf Qdisc
	if i.Flags&WithQdiscTbf != 0 {
		i.Tbf.LinkIndex = i.Link.Attrs().Index
		i.Tbf.Limit = 0x7000
		i.Tbf.Minburst = 1600
		i.Tbf.Buffer = 300000
		i.Tbf.Peakrate = 0x1000000
		i.Tbf.QdiscAttrs = nl.QdiscAttrs{
			LinkIndex: i.Link.Attrs().Index,
			Handle:    nl.MakeHandle(2, 0),
			Parent:    pHandle,
		}

		logger.Info("Adding TBF qdisc to interface")
		if err := n.nlHandle.QdiscAdd(&i.Tbf); err != nil {
			return err
		}
	}

	// Setting link up
	if err := n.nlHandle.LinkSetUp(i.Link); err != nil {
		return err
	}

	// Start packet capturing if requested on network or host level
	captures := []*Capture{}
	captures = append(captures, n.network.Captures...)
	captures = append(captures, n.Captures...)
	captures = append(captures, i.Captures...)

	for _, c := range captures {
		if c != nil && (c.FilterInterface == nil || c.FilterInterface(i)) {
			if _, err := c.startInterface(i); err != nil {
				return fmt.Errorf("failed to capture interface: %w", err)
			}
		}
	}

	n.Interfaces = append(n.Interfaces, i)

	if err := n.network.GenerateHostsFile(); err != nil {
		return fmt.Errorf("failed to update hosts file")
	}

	return nil
}

func (n *BaseNode) Close() error {
	for _, i := range n.Interfaces {
		if err := i.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (n *BaseNode) Teardown() error {
	if err := n.Namespace.Close(); err != nil {
		return err
	}

	nsMount := filepath.Join(n.BasePath, "ns", "net")
	if err := unix.Unmount(nsMount, 0); err != nil {
		return err
	}

	return os.RemoveAll(n.BasePath)
}

// WriteProcFS write a value to a path within the ProcFS by entering the namespace of this node.
func (n *BaseNode) WriteProcFS(path, value string) error {
	n.logger.Info("Updating procfs",
		zap.String("path", path),
		zap.String("value", value),
	)

	return n.RunFunc(func() error {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.WriteString(value)

		return err
	})
}

// EnableForwarding enables forwarding for both IPv4 and IPv6 protocols in the kernel for all interfaces
func (n *BaseNode) EnableForwarding() error {
	if err := n.WriteProcFS("/proc/sys/net/ipv4/conf/all/forwarding", "1"); err != nil {
		return err
	}

	return n.WriteProcFS("/proc/sys/net/ipv6/conf/all/forwarding", "1")
}

// AddRoute adds a route to the node.
func (n *BaseNode) AddRoute(r *nl.Route) error {
	n.logger.Info("Add route",
		zap.Any("dst", r.Dst),
		zap.Any("gw", r.Gw),
	)

	return n.nlHandle.RouteAdd(r)
}

// AddDefaultRoute adds a default route for this node by providing a default gateway.
func (n *BaseNode) AddDefaultRoute(gw net.IP) error {
	if gw.To4() != nil {
		return n.AddRoute(&nl.Route{
			Dst: &DefaultIPv4Mask,
			Gw:  gw,
		})
	}

	return n.AddRoute(&nl.Route{
		Dst: &DefaultIPv6Mask,
		Gw:  gw,
	})
}

// AddInterface adds an interface to the list of configured interfaces
func (n *BaseNode) AddInterface(i *Interface) {
	n.ConfiguredInterfaces = append(n.ConfiguredInterfaces, i)
}
