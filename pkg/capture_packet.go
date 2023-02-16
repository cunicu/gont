package gont

import (
	"time"

	"github.com/gopacket/gopacket"
)

type CapturePacket struct {
	gopacket.CaptureInfo
	Data []byte

	Decoded gopacket.Packet

	Interface *captureInterface
}

func (p CapturePacket) Time() time.Time {
	return p.Timestamp
}

func (p CapturePacket) Decode(dOpts gopacket.DecodeOptions) gopacket.Packet {
	return gopacket.NewPacket(p.Data, p.Interface.pcapInterface.LinkType, dOpts)
}
