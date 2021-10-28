package gont

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/stv0g/gont/internal/utils"
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

type BaseNode struct {
	Node

	*Namespace
	Network *Network

	name string

	ExistingNamespace       string
	ExistingDockerContainer string

	BasePath string
}

func (n *Network) AddNode(name string, opts ...Option) (*BaseNode, error) {
	var err error

	basePath := filepath.Join(n.BasePath, "nodes", name)
	for _, path := range []string{"ns"} {
		path = filepath.Join(basePath, path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}

	node := &BaseNode{
		name:     name,
		Network:  n,
		BasePath: basePath,
	}

	log.WithField("name", name).Info("Adding new node")

	for _, opt := range opts {
		if nopt, ok := opt.(NodeOption); ok {
			nopt.Apply(node)
		}
	}

	if node.ExistingNamespace != "" {
		// Use an existing namespace created by "ip netns add"
		if nsh, err := netns.GetFromName(node.ExistingNamespace); err != nil {
			return nil, fmt.Errorf("failed to find existing network namespace %s: %w", node.ExistingNamespace, err)
		} else {
			node.Namespace = &Namespace{
				Name:     node.ExistingNamespace,
				NsHandle: nsh,
			}
		}
	} else if node.ExistingDockerContainer != "" {
		// Use an existing net namespace from a Docker container
		if nsh, err := netns.GetFromDocker(node.ExistingDockerContainer); err != nil {
			return nil, fmt.Errorf("failed to find existing docker container %s: %w", node.ExistingNamespace, err)
		} else {
			node.Namespace = &Namespace{
				Name:     node.ExistingDockerContainer,
				NsHandle: nsh,
			}
		}
	} else {
		// Create a new network namespace
		nsName := fmt.Sprintf("%s%s-%s", n.NSPrefix, n.Name, name)
		if node.Namespace, err = NewNamespace(nsName); err != nil {
			return nil, err
		}
	}

	if node.Handle == nil {
		node.Handle, err = nl.NewHandleAt(node.NsHandle)
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

	n.Nodes[name] = node

	return node, nil
}

func (n *BaseNode) Base() *BaseNode {
	return n
}

func (n *BaseNode) Name() string {
	return n.name
}

func (n *BaseNode) ConfigurePort(p Port) error {
	log.WithField("intf", n.name+"/"+p.Name).Info("Configuring port")

	link, err := n.Handle.LinkByName(p.Name)
	if err != nil {
		return err
	}

	if p.Group != DeviceGroupDefault {
		log.WithFields(log.Fields{
			"intf":  p,
			"group": p.Group,
		}).Info("Setting device group")
		if err := n.Handle.LinkSetGroup(link, int(p.Group)); err != nil {
			return err
		}
	}

	var pHandle uint32 = nl.HANDLE_ROOT
	if p.Flags&WithQdiscNetem != 0 {
		attr := nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(1, 0),
			Parent:    pHandle,
		}

		netem := nl.NewNetem(attr, p.Netem)

		if err := n.Handle.QdiscAdd(netem); err != nil {
			return err
		}

		pHandle = netem.Handle
	}
	if p.Flags&WithQdiscTbf != 0 {
		p.Tbf.LinkIndex = link.Attrs().Index
		p.Tbf.Limit = 0x7000
		p.Tbf.Minburst = 1600
		p.Tbf.Buffer = 300000
		p.Tbf.Peakrate = 0x1000000
		p.Tbf.QdiscAttrs = nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(2, 0),
			Parent:    pHandle,
		}

		if err := n.Handle.QdiscAdd(&p.Tbf); err != nil {
			return err
		}
	}

	if err := n.Handle.LinkSetUp(link); err != nil {
		return err
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

	if err := os.RemoveAll(n.BasePath); err != nil {
		return err
	}

	return nil
}

func (n *BaseNode) WriteProcFS(path, value string) error {
	_, _, err := n.Run("sh", "-c", "echo "+value+" > "+path)
	return err
}

func (n *BaseNode) EnableForwarding() error {
	if err := n.WriteProcFS("/proc/sys/net/ipv4/conf/all/forwarding", "1"); err != nil {
		return err
	}

	if err := n.WriteProcFS("/proc/sys/net/ipv6/conf/all/forwarding", "1"); err != nil {
		return err
	}

	return nil
}

func (n *BaseNode) LinkAddAddress(name string, addr net.IPNet) error {
	link, err := n.Handle.LinkByName(name)
	if err != nil {
		return err
	}

	nlAddr := &nl.Addr{
		IPNet: &addr,
	}

	log.WithFields(log.Fields{
		"intf": n.name + "/" + name,
		"addr": addr.String(),
	}).Info("Adding new address to interface")

	if err := n.Handle.AddrAdd(link, nlAddr); err != nil {
		return err
	}

	return err
}

func (n *BaseNode) AddRoute(dst net.IPNet, gw net.IP) error {
	log.WithFields(log.Fields{
		"node": n.name,
		"dst":  dst.String(),
		"gw":   gw.String(),
	}).Info("Add route")

	if err := n.Handle.RouteAdd(&nl.Route{
		Dst: &dst,
		Gw:  gw,
	}); err != nil {
		return err
	}

	return nil
}

func (n *BaseNode) AddDefaultRoute(gw net.IP) error {
	if gw.To4() != nil {
		return n.AddRoute(net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.CIDRMask(0, net.IPv6len*8),
		}, gw)
	} else {
		return n.AddRoute(net.IPNet{
			IP:   net.IPv6zero,
			Mask: net.CIDRMask(0, net.IPv6len*8),
		}, gw)
	}
}
