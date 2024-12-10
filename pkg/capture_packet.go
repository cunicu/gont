// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"github.com/gopacket/gopacket"
)

type CapturePacket struct {
	gopacket.CaptureInfo
	Data []byte

	Interface *captureInterface
}

func (p CapturePacket) Decode(dOpts gopacket.DecodeOptions) gopacket.Packet {
	return gopacket.NewPacket(p.Data, p.Interface.pcapInterface.LinkType, dOpts)
}
