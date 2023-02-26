// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"io"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/stv0g/gont/pkg/trace"
)

const (
	LinkTypeUser0 = 147
	LinkTypeTrace = LinkTypeUser0
)

var _ PacketSource = (*traceEventPacketSource)(nil)

type traceEventPacketSource struct {
	tracepoints chan trace.Event
	count       uint64
}

func newTracepointPacketSource() *traceEventPacketSource {
	return &traceEventPacketSource{
		tracepoints: make(chan trace.Event),
	}
}

func (tps *traceEventPacketSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	tps.count++
	tp, ok := <-tps.tracepoints
	if !ok {
		return nil, gopacket.CaptureInfo{}, io.EOF
	}

	return SerializePacket(&tp)
}

func (tps *traceEventPacketSource) Stats() (captureStats, error) {
	return captureStats{
		PacketsReceived: tps.count,
	}, nil
}

func (tps *traceEventPacketSource) LinkType() layers.LinkType {
	// TODO: Register our own DLT value?
	// https://www.tcpdump.org/linktypes.html
	return LinkTypeTrace
}

func (tps *traceEventPacketSource) SourceTracepoint(tp trace.Event) {
	tps.tracepoints <- tp
}

func (tps *traceEventPacketSource) Close() error {
	close(tps.tracepoints)
	return nil
}

func SerializePacket(t *trace.Event) (data []byte, ci gopacket.CaptureInfo, err error) {
	buf, err := t.MarshalCBOR()
	if err != nil {
		return nil, gopacket.CaptureInfo{}, err
	}

	return buf, gopacket.CaptureInfo{
		Timestamp:     t.Timestamp,
		Length:        len(buf),
		CaptureLength: len(buf),
	}, nil
}
