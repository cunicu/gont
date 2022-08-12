package gont_test

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	copt "github.com/stv0g/gont/pkg/options/capture"
	"go.uber.org/zap"
)

func TestCaptureNetwork(t *testing.T) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
		sw1    *g.Switch
	)

	tmpPCAP, err := os.CreateTemp("", "gont-capture-*.pcapng")
	if err != nil {
		t.Fatalf("Failed to open temporary file: %s", err)
	}

	ch := make(chan g.CapturePacket)
	go func() {
		logger := zap.L().Named("channel")
		for p := range ch {
			layers := []string{}
			for _, layer := range p.Layers() {
				layers = append(layers, layer.LayerType().String())
			}

			logger.Info("Packet",
				zap.Strings("layers", layers),
				zap.String("node", p.Interface.Node.Name()),
				zap.String("intf", p.Interface.Name))
		}
	}()

	cb := func(p g.CapturePacket) {
		// fmt.Println("Callback", p.String())
	}

	// tcpdump -i enp0s5 'icmp6[icmp6type]=icmp6-echo || icmp6[icmp6type]=icmp6-echoreply' -d
	// instrs := []bpf.Instruction{
	// 	bpf.LoadAbsolute{Off: 12, Size: 2},                                      // 0
	// 	bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x86dd, SkipTrue: 0, SkipFalse: 6}, // 1
	// 	bpf.LoadAbsolute{Off: 20, Size: 1},                                      // 2
	// 	bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x003a, SkipTrue: 1, SkipFalse: 4}, // 3
	// 	bpf.LoadAbsolute{Off: 54, Size: 1},                                      // 4
	// 	bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x0080, SkipTrue: 1, SkipFalse: 0}, // 5
	// 	bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x0081, SkipTrue: 0, SkipFalse: 1}, // 6
	// 	bpf.RetConstant{Val: 1600},                                              // 7
	// 	bpf.RetConstant{Val: 0},                                                 // 8
	// }

	c1 := gopt.CaptureAll(
		copt.ToFile(tmpPCAP),
		copt.ToChannel(ch),
		copt.Callback(cb),
		copt.CaptureLength(1600),
		copt.Promiscuous(true),
		copt.FilterExpression("icmp6[icmp6type]=icmp6-echo || icmp6[icmp6type]=icmp6-echoreply"),
		// copt.FilterInstructions(instrs),
		copt.FilterInterfaces(func(i *g.Interface) bool {
			return strings.HasPrefix(i.Name, "veth")
		}),
		copt.FilterPackets(func(p *g.CapturePacket) bool {
			if layer := p.Layer(layers.LayerTypeICMPv6); layer != nil {
				typec := layer.(*layers.ICMPv6).TypeCode.Type()

				return typec == layers.ICMPv6TypeEchoRequest || typec == layers.ICMPv6TypeEchoReply
			} else {
				return false
			}
		}),
		copt.Comment("Some random comment which will be included in the capture file"),
	)

	opts := g.Customize(globalNetworkOptions, c1, // Also multiple capturers are supported
		gopt.CaptureAll(
			copt.ToFilename("all.pcapng"), // We can create a file
		),
	)

	if n, err = g.NewNetwork(*nname, opts...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to add switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		gopt.Interface("veth0", sw1,
			gopt.AddressIP("fc::1/64"),
			gopt.Capture(
				copt.Filename("{{ .Host }}_{{ .Interface }}.pcapng"),
			),
		),
	); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		gopt.Interface("veth0", sw1,
			gopt.AddressIP("fc::2/64"),
		),
	); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if _, err := h1.Ping(h2); err != nil {
		t.Fatalf("Failed to ping: %s", err)
	}

	// We need to wait some time until PCAP has captured the packets
	time.Sleep(1 * time.Second)

	rd, err := c1.Reader()
	if err != nil {
		t.Fatalf("Failed to get reader for PCAPng file: %s", err)
	}

	h1veth0 := h1.Interface("veth0")
	h2veth0 := h2.Interface("veth0")

	pkt, _, intf, eof := nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "h1/veth0" {
		t.Fatalf("Invalid 1st packet")
	}
	if v6, ok := pkt.NetworkLayer().(*layers.IPv6); !ok {
		t.Fatalf("Wrong network layer: %s", pkt.NetworkLayer().LayerType().String())
	} else {
		if !v6.SrcIP.Equal(h1veth0.Addresses[0].IP) {
			t.Fatalf("Invalid source IP: %s != %s",
				v6.SrcIP.String(),
				h1veth0.Addresses[0].IP.String(),
			)
		}

		if !v6.DstIP.Equal(h2veth0.Addresses[0].IP) {
			t.Fatalf("Invalid source IP: %s != %s",
				v6.SrcIP.String(),
				h1veth0.Addresses[0].IP.String(),
			)
		}
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "sw1/veth-h1" {
		t.Fatalf("Invalid 2nd packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "sw1/veth-h2" {
		t.Fatalf("Invalid 3rd packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "h2/veth0" {
		t.Fatalf("Invalid 4th packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "h2/veth0" {
		t.Fatalf("Invalid 5th packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "sw1/veth-h2" {
		t.Fatalf("Invalid 6th packet: %s", intf.Name)
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "sw1/veth-h1" {
		t.Fatalf("Invalid 7th packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Expected more packets")
	}
	if intf.Name != "h1/veth0" {
		t.Fatalf("Invalid 7th packet")
	}

	_, _, _, eof = nextPacket(t, rd)
	if eof != true {
		t.Fatalf("Did not expect EOF")
	}

	if rd.NInterfaces() != 4 {
		t.Fatalf("Invalid number of interfaces: %d != 4", rd.NInterfaces())
	}

	if c1.Count() != 8 {
		t.Fatalf("Invalid number of packets: %d != 8", c1.Count())
	}

	if err := n.Close(); err != nil {
		t.Fatalf("Failed to close network: %s", err)
	}

	if err := tmpPCAP.Close(); err != nil {
		t.Fatalf("Failed to close file: %s", err)
	}
}

func nextPacket(t *testing.T, rd *pcapgo.NgReader) (gopacket.Packet, *gopacket.CaptureInfo, *pcapgo.NgInterface, bool) {
	data, ci, err := rd.ZeroCopyReadPacketData()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, nil, true
		}

		t.Fatalf("Failed to read packet data: %s", err)
	}

	intf, err := rd.Interface(ci.InterfaceIndex)
	if err != nil {
		t.Fatalf("Received packet from unknown interface: %s", err)
	}

	return gopacket.NewPacket(data, layers.LinkTypeEthernet, gopacket.Default), &ci, &intf, false
}
