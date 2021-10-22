package gont

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

type Host struct {
	BaseNode

	// Options
	GatewayIPv4 net.IP
	GatewayIPv6 net.IP
	Forwarding  bool

	Interfaces []Interface
}

// Getter
func (h *Host) Base() *BaseNode {
	return &h.BaseNode
}

// Options
func (h *Host) Apply(i *Interface) {
	i.Node = h
}

func (n *Network) AddHost(name string, opts ...Option) (*Host, error) {
	node, err := n.AddNode(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %s", err)
	}

	host := &Host{
		BaseNode:   *node,
		Interfaces: []Interface{},
	}

	n.Nodes[name] = host // TODO: quirk to get n.UpdateHostsFile() working

	// Apply host options
	for _, opt := range opts {
		if hopt, ok := opt.(HostOption); ok {
			hopt.Apply(host)
		}
	}

	// Add interfaces
	for _, intf := range host.Interfaces {
		if err := host.AddInterface(intf); err != nil {
			return nil, fmt.Errorf("failed to add interface: %w", err)
		}
	}

	// Configure host
	if host.GatewayIPv4 != nil {
		host.AddRoute(net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.CIDRMask(0, net.IPv4len*8),
		}, host.GatewayIPv4)
	}

	if host.GatewayIPv6 != nil {
		host.AddRoute(net.IPNet{
			IP:   net.IPv6zero,
			Mask: net.CIDRMask(0, net.IPv6len*8),
		}, host.GatewayIPv6)
	}

	if host.Forwarding {
		if err := host.EnableForwarding(); err != nil {
			return nil, fmt.Errorf("failed to enable forwarding: %w", err)
		}
	}

	if err := host.ConfigureInterface(loopbackInterface); err != nil {
		return nil, fmt.Errorf("failed to configure loopback interface: %w", err)
	}

	return host, nil
}

func (h *Host) AddInterface(i Interface) error {
	peerDev := fmt.Sprintf("veth-%s", h.name)

	l := Interface{
		Port: Port{
			Name: i.Name,
			Node: h,
		},
		Addresses: i.Addresses,
	}

	r := Port{
		Name: peerDev,
		Node: i.Node,
	}

	var addrs []string
	for _, a := range l.Addresses {
		addrs = append(addrs, a.String())
	}

	log.WithFields(log.Fields{
		"intf":      l,
		"intf_peer": r,
		"addresses": addrs,
	}).Info("Adding interface")

	return h.Network.AddLink(l, r)
}

func (h *Host) ConfigureInterface(i Interface) error {
	// Disable duplicate address detection (DAD) before adding addresses
	// so we dont end up with tentative addresses and slow test executions
	if !i.EnableDAD {
		fn := filepath.Join("/proc/sys/net/ipv6/conf", i.Name, "accept_dad")
		if err := h.WriteProcFS(fn, "0"); err != nil {
			return fmt.Errorf("failed to enabled IPv6 forwarding: %s", err)
		}
	}

	for _, addr := range i.Addresses {
		if err := h.LinkAddAddress(i.Name, addr); err != nil {
			return fmt.Errorf("failed to add link address: %s", err)
		}
	}

	// Bring interface up
	if err := h.BaseNode.ConfigurePort(i.Port); err != nil {
		return err
	}

	h.Interfaces = append(h.Interfaces, i) // TODO: arent the interface already in there for some cases?

	if err := h.Network.UpdateHostsFile(); err != nil {
		return fmt.Errorf("failed to update hosts file")
	}

	return nil
}

func (h *Host) Ping(o *Host) (*ping.Statistics, error) {
	return h.PingWithOptions(o, "ip", 1, time.Second, time.Second, true)
}

func (h *Host) PingWithNetwork(o *Host, net string) (*ping.Statistics, error) {
	return h.PingWithOptions(o, net, 1, time.Second, time.Second, true)
}

func (h *Host) PingWithOptions(o *Host, net string, count int, timeout time.Duration, intv time.Duration, output bool) (*ping.Statistics, error) {
	var err error

	p := ping.New(o.Name())

	p.Count = count
	p.RecordRtts = true
	p.Timeout = timeout
	p.Interval = intv

	// Find first IP address of first interface
	ip := o.LookupAddress(net)
	if ip == nil {
		return nil, errors.New("failed to find address")
	}

	p.SetIPAddr(ip)
	p.SetLogger(log.WithField("logger", "ping"))
	p.SetPrivileged(true)

	if output {
		p.OnRecv = func(p *ping.Packet) {
			fmt.Printf("%d bytes from %s (%s): icmp_seq=%d ttl=%d time=%v\n",
				p.Nbytes,
				p.Addr,
				p.IPAddr.String(),
				p.Seq,
				p.Ttl,
				p.Rtt,
			)
		}

		p.OnFinish = func(s *ping.Statistics) {
			fmt.Printf("-- %s (%s) ping statistics ---\n"+
				"%d packets transmitted, %d received, %d duplicates, %.2f%% packet loss\n"+
				"rtt min/avg/max/mdev = %s/%s/%s/%s\n",
				o.Name(),
				s.IPAddr,
				s.PacketsSent,
				s.PacketsRecv,
				s.PacketsRecvDuplicates,
				s.PacketLoss,
				s.MinRtt,
				s.AvgRtt,
				s.MaxRtt,
				s.StdDevRtt,
			)
		}
	}

	if err = h.RunFunc(func() error {
		if output {
			fmt.Printf("PING %s(%s) %d data bytes\n",
				o.Name(),
				p.Addr(),
				p.Size,
			)
		}

		return p.Run()
	}); err != nil {
		return nil, err
	}

	return p.Statistics(), err
}

func (h *Host) Traceroute(o *Host, opts ...string) error {
	args := append([]string{o.name}, opts...)
	_, _, err := h.Run("traceroute", args...)
	return err
}

func (h *Host) LookupAddress(n string) *net.IPAddr {
	for _, i := range h.Interfaces {
		if i.Name == loopbackInterfaceName {
			continue
		}

		for _, a := range i.Addresses {
			ip := &net.IPAddr{
				IP: a.IP,
			}

			switch n {
			case "ip":
				return ip
			case "ip4":
				if a.IP.To4() != nil {
					return ip
				}
			case "ip6":
				if a.IP.To4() == nil {
					return ip
				}
			}
		}
	}

	return nil
}
