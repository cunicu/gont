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
		if err := host.AddDefaultRoute(host.GatewayIPv4); err != nil {
			return nil, err
		}
	}

	if host.GatewayIPv6 != nil {
		if err := host.AddDefaultRoute(host.GatewayIPv6); err != nil {
			return nil, err
		}
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
	return h.PingWithOptions(o, "ip", 1, 2*time.Second, time.Second, true)
}

func (h *Host) PingWithNetwork(o *Host, net string) (*ping.Statistics, error) {
	return h.PingWithOptions(o, net, 1, 2*time.Second, time.Second, true)
}

func (h *Host) PingWithOptions(o *Host, net string, count int, timeout time.Duration, intv time.Duration, output bool) (*ping.Statistics, error) {
	var err error

	p := ping.New(o.Name())

	p.Count = count
	p.RecordRtts = true
	p.Timeout = timeout
	p.Interval = intv

	if h.Network != o.Network {
		return nil, fmt.Errorf("hosts must be on same network")
	}

	// Find first IP address of first interface
	ip := o.LookupAddress(net)
	if ip == nil {
		return nil, errors.New("failed to find address")
	}

	logger := log.WithFields(log.Fields{
		"logger": "ping",
		"ns":     h.name,
	})

	p.SetIPAddr(ip)
	p.SetLogger(logger)
	p.SetPrivileged(true)
	p.SetNetwork(net)

	if output {
		p.OnRecv = func(p *ping.Packet) {
			logger.Printf("%d bytes from %s (%s): icmp_seq=%d ttl=%d time=%v\n",
				p.Nbytes,
				p.Addr,
				p.IPAddr.String(),
				p.Seq,
				p.Ttl,
				p.Rtt,
			)
		}

		p.OnFinish = func(s *ping.Statistics) {
			logger.Printf("-- %s (%s) ping statistics ---", o.Name(), s.IPAddr)
			logger.Printf("%d packets transmitted, %d received, %d duplicates, %.2f%% packet loss\n", s.PacketsSent, s.PacketsRecv, s.PacketsRecvDuplicates, s.PacketLoss)
			logger.Printf("rtt min/avg/max/mdev = %s/%s/%s/%s\n", s.MinRtt, s.AvgRtt, s.MaxRtt, s.StdDevRtt)
		}
	}

	if err = h.RunFunc(func() error {
		if output {
			logger.Printf("PING %s(%s) %d data bytes\n",
				o.Name(),
				p.Addr(),
				p.Size,
			)
		}

		return p.Run()
	}); err != nil {
		return nil, err
	}

	lost := p.PacketsSent - p.PacketsRecv
	if lost > 0 {
		err = fmt.Errorf("lost %d packets", lost)
	}

	return p.Statistics(), err
}

func (h *Host) Traceroute(o *Host, opts ...interface{}) error {
	if h.Network != o.Network {
		return fmt.Errorf("hosts must be on same network")
	}

	opts = append(opts, o)
	_, _, err := h.Run("traceroute", opts...)
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

			isV4 := len(a.IP.To4()) == net.IPv4len

			switch n {
			case "ip":
				return ip
			case "ip4":
				if isV4 {
					return ip
				}
			case "ip6":
				if !isV4 {
					return ip
				}
			}
		}
	}

	return nil
}
