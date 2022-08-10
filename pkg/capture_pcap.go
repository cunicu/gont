//go:build cgo

package gont

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

const CGoPCAP = true

type pcapHandle struct {
	*pcap.Handle
}

func (c *Capture) createHandle(name string) (handle, error) {
	if c.Timeout.Microseconds() == 0 {
		c.Timeout = pcap.BlockForever
	}

	ihdl, err := pcap.NewInactiveHandle(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create PCAP handle: %w", err)
	}

	if err := ihdl.SetPromisc(c.Promiscuous); err != nil {
		return nil, fmt.Errorf("failed to set: %w", err)
	}

	if err := ihdl.SetSnapLen(c.CaptureLength); err != nil {
		return nil, fmt.Errorf("failed to set: %w", err)
	}

	if err := ihdl.SetTimeout(c.Timeout); err != nil {
		return nil, fmt.Errorf("failed to set: %w", err)
	}

	hdl, err := ihdl.Activate()
	if err != nil {
		return nil, fmt.Errorf("failed to activate PCAP handle: %w", err)
	}

	if c.FilterExpression != "" {
		if err := hdl.SetBPFFilter(c.FilterExpression); err != nil {
			return nil, fmt.Errorf("failed to set BPF filter expression: %w", err)
		}
	}

	if c.FilterInstructions != nil {
		instrs := []pcap.BPFInstruction{}
		for _, instr := range c.FilterInstructions {
			ainstr, err := instr.Assemble()
			if err != nil {
				return nil, fmt.Errorf("failed to assemble BPF bytecode: %w", err)
			}

			instrs = append(instrs, pcap.BPFInstruction{
				Code: ainstr.Op,
				Jt:   ainstr.Jt,
				Jf:   ainstr.Jf,
				K:    ainstr.K,
			})
		}

		if err := hdl.SetBPFInstructionFilter(instrs); err != nil {
			return nil, fmt.Errorf("failed to set BFP filter instructions: %w", err)
		}
	}

	return &pcapHandle{
		Handle: hdl,
	}, nil
}

func (h pcapHandle) Stats() (CaptureStats, error) {
	s, err := h.Handle.Stats()
	if err != nil {
		return CaptureStats{}, err
	}

	return CaptureStats{
		PacketsReceived: s.PacketsReceived,
		PacketsDropped:  s.PacketsDropped,
	}, nil
}
