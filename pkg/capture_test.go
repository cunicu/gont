package gont_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func TestCaptureNetwork(t *testing.T) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
	)

	tmpPCAP, err := os.CreateTemp("", "gont-capture-*.pcapng")
	if err != nil {
		t.Fatalf("Failed to open temporary file: %s", err)
	}

	opts = append(opts,
		o.CaptureNetwork(
			o.File{tmpPCAP},
			o.CaptureLength(1600),
			o.Promisc(false),
			o.BPFilter("icmp6[icmp6type]=icmp6-echo || icmp6[icmp6type]=icmp6-echoreply"),
			o.Comment("Some random comment which will be included in the capture file"),
		),
		// Also multiple capturers are supported
		o.CaptureNetwork(
			o.Filename("all.pcapng"), // We can create a file
		),
	)

	if n, err = g.NewNetwork(*nname, opts...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if h1, err = n.AddHost("h1"); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if h2, err = n.AddHost("h2"); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if err := n.AddLink(
		o.Interface("veth0", h1,
			o.AddressIP("fc::1/64"),
			o.Capture(
				o.Filename("{{ .Host }}_{{ .Interface }}.pcapng"),
			),
		),
		o.Interface("veth0", h2,
			o.AddressIP("fc::2/64"),
		),
	); err != nil {
		t.Errorf("Failed to setup link: %s", err)
	}

	if _, err := h1.Ping(h2); err != nil {
		t.Fatalf("Failed to ping: %s", err)
	}

	for _, c := range n.Captures {
		if err := c.Flush(); err != nil {
			t.Fatalf("Failed to flush capture: %s", err)
		}
	}

	if _, err := tmpPCAP.Seek(0, 0); err != nil {
		t.Fatalf("Failed to rewind file: %s", err)
	}

	rd, err := pcapgo.NewNgReader(tmpPCAP, pcapgo.DefaultNgReaderOptions)
	if err != nil {
		t.Fatalf("Failed to read PCAPng file: %s", err)
	}

	h1veth0 := h1.Interface("veth0")
	h2veth0 := h2.Interface("veth0")

	pkt, _, intf, eof := nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Did more packets")
	}
	if intf.Name != "h1/veth0" {
		t.Fatalf("Invalid first packet")
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
		t.Fatalf("Did more packets")
	}
	if intf.Name != "h2/veth0" {
		t.Fatalf("Invalid first packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Did more packets")
	}
	if intf.Name != "h2/veth0" {
		t.Fatalf("Invalid first packet")
	}

	_, _, intf, eof = nextPacket(t, rd)
	if eof == true {
		t.Fatalf("Did more packets")
	}
	if intf.Name != "h1/veth0" {
		t.Fatalf("Invalid first packet")
	}

	_, _, _, eof = nextPacket(t, rd)
	if eof != true {
		t.Fatalf("Did not expect EOF")
	}

	if rd.NInterfaces() != 4 {
		t.Fatalf("Invalid number of interfaces: %d != 4", rd.NInterfaces())
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
