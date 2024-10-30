// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/capture"
	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCaptureNetwork(t *testing.T) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
		sw1    *g.Switch
	)

	tmpPCAP, err := os.CreateTemp(t.TempDir(), "gont-capture-*.pcapng")
	require.NoError(t, err, "Failed to open temporary file")

	ch := make(chan g.CapturePacket)
	go func() {
		logger := zap.L().Named("channel")
		for p := range ch {
			pp := p.Decode(gopacket.DecodeOptions{})

			layers := []string{}
			for _, layer := range pp.Layers() {
				layers = append(layers, layer.LayerType().String())
			}

			logger.Info("Packet",
				zap.Strings("layers", layers),
				zap.Any("intf", p.Interface))
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

	c1 := g.NewCapture(
		co.ToFile(tmpPCAP),
		co.ToChannel(ch),
		co.Callback(cb),
		co.SnapshotLength(1600),
		co.Promiscuous(true),
		co.FilterExpression("icmp6[icmp6type]=icmp6-echo || icmp6[icmp6type]=icmp6-echoreply"),
		// co.FilterInstructions(instrs),
		co.FilterInterfaces(func(i *g.Interface) bool {
			return strings.HasPrefix(i.Name, "veth")
		}),
		co.FilterPackets(func(p *g.CapturePacket) bool {
			pp := p.Decode(gopacket.DecodeOptions{})
			if layer := pp.Layer(layers.LayerTypeICMPv6); layer != nil {
				typec := layer.(*layers.ICMPv6).TypeCode.Type()

				return typec == layers.ICMPv6TypeEchoRequest || typec == layers.ICMPv6TypeEchoReply
			}

			return false
		}),
		co.Comment("Some random comment which will be included in the capture file"),
	)

	n, err = g.NewNetwork(*nname,
		g.Customize(globalNetworkOptions, c1, // Also multiple capturers are supported
			g.NewCapture(
				co.ToFilename("all.pcapng")), // We can create a file
		)...)
	require.NoError(t, err, "Failed to create network")

	sw1, err = n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to add switch")

	h1, err = n.AddHost("h1",
		g.NewInterface("veth0", sw1,
			o.AddressIP("fc::1/64"),
			g.NewCapture(
				co.Filename("{{ .Node }}_{{ .Interface }}.pcapng"))))
	require.NoError(t, err, "Failed to add host")

	h2, err = n.AddHost("h2",
		g.NewInterface("veth0", sw1,
			o.AddressIP("fc::2/64")))
	require.NoError(t, err, "Failed to add host")

	_, err = h1.Ping(h2)
	require.NoError(t, err, "Failed to ping")

	// Read-back PCAP file
	// We need to wait some time until PCAP has captured the packets
	time.Sleep(1 * time.Second)

	err = c1.Flush()
	require.NoError(t, err, "Failed to flush capture")

	_, err = tmpPCAP.Seek(0, 0)
	require.NoError(t, err, "Failed to rewind file")

	rd, err := pcapgo.NewNgReader(tmpPCAP, pcapgo.DefaultNgReaderOptions)
	require.NoError(t, err, "Failed to read PCAPng file")

	h1veth0 := h1.Interface("veth0")
	h2veth0 := h2.Interface("veth0")

	pkt, _, intf, eof := nextPacket(t, rd)
	require.Equal(t, eof, false, "Expected more packets")

	require.Equal(t, intf.Name, "h1/veth0", "Invalid 1st packet")

	v6, ok := pkt.NetworkLayer().(*layers.IPv6)
	require.True(t, ok, "Wrong network layer: %s", pkt.NetworkLayer().LayerType().String())

	require.True(t, v6.SrcIP.Equal(h1veth0.Addresses[0].IP),
		"Invalid source IP: %s != %s",
		v6.SrcIP.String(),
		h1veth0.Addresses[0].IP.String(),
	)

	require.True(t, v6.DstIP.Equal(h2veth0.Addresses[0].IP),
		"Invalid source IP: %s != %s",
		v6.SrcIP.String(),
		h1veth0.Addresses[0].IP.String(),
	)

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "sw1/veth-h1", "Invalid 2nd packet")

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "sw1/veth-h2", "Invalid 3rd packet")

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "h2/veth0", "Invalid 4th packet")

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "h2/veth0", "Invalid 5th packet")

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "sw1/veth-h2", "Invalid 6th packet: %s", intf.Name)

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "sw1/veth-h1", "Invalid 7th packet")

	_, _, intf, eof = nextPacket(t, rd)
	require.False(t, eof, "Expected more packets")
	require.Equal(t, intf.Name, "h1/veth0", "Invalid 7th packet")

	_, _, _, eof = nextPacket(t, rd)
	require.True(t, eof, "Expected EOF")

	require.Equal(t, rd.NInterfaces(), 4, "Invalid number of interfaces")
	require.EqualValues(t, c1.Count(), 8, "Invalid number of packets")

	err = n.Close()
	require.NoError(t, err, "Failed to close network")

	err = tmpPCAP.Close()
	require.NoError(t, err, "Failed to close file")
}

func nextPacket(t *testing.T, rd *pcapgo.NgReader) (gopacket.Packet, *gopacket.CaptureInfo, *pcapgo.NgInterface, bool) {
	data, ci, err := rd.ZeroCopyReadPacketData()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, nil, true
		}

		require.NoError(t, err, "Failed to read packet data")
	}

	intf, err := rd.Interface(ci.InterfaceIndex)
	require.NoError(t, err, "Received packet from unknown interface")

	return gopacket.NewPacket(data, layers.LinkTypeEthernet, gopacket.Default), &ci, &intf, false
}
