//go:build !cgo

// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"

	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"golang.org/x/net/bpf"
)

const CGoPCAP = false

type pcapgoPacketSource struct {
	*pcapgo.EthernetHandle
}

func (c *Capture) createPCAPHandle(name string) (handle, error) {
	hdl, err := pcapgo.NewEthernetHandle(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open PCAP handle: %w", err)
	}

	if err := hdl.SetCaptureLength(int(c.CaptureLength)); err != nil {
		return nil, fmt.Errorf("failed to set capture length: %w", err)
	}

	if err := hdl.SetPromiscuous(c.Promiscuous); err != nil {
		return nil, fmt.Errorf("failed to set promiscuous mode: %w", err)
	}

	if c.FilterExpression != "" {
		return nil, fmt.Errorf("libpcap filter expressions require CGo")
	}

	if c.FilterInstructions != nil {
		ainstrs := []bpf.RawInstruction{}
		for _, instr := range c.FilterInstructions {
			ainstr, err := instr.Assemble()
			if err != nil {
				return nil, fmt.Errorf("failed to assemble BPF instruction: %w", err)
			}

			ainstrs = append(ainstrs, ainstr)
		}

		if err := hdl.SetBPF(ainstrs); err != nil {
			return nil, fmt.Errorf("failed to set BPF filter instructions: %w", err)
		}
	}

	return pcapgoPacketSource{
		EthernetHandle: hdl,
	}, nil
}

func (h pcapgoPacketSource) Stats() (CaptureStats, error) {
	s, err := h.EthernetHandle.Stats()
	if err != nil {
		return CaptureStats{}, err
	}

	return CaptureStats{
		PacketsReceived: int(s.Packets),
		PacketsDropped:  int(s.Drops),
	}, nil
}

func (h pcapgoPacketSource) LinkType() layers.LinkType {
	return layers.LinkTypeEthernet
}
