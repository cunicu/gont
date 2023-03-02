// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"io"
	"time"

	"github.com/gopacket/gopacket/pcapgo"
	"go.uber.org/zap"
)

type captureStats struct {
	PacketsReceived uint64
	PacketsDropped  uint64
}

type captureInterface struct {
	*Interface

	pcapInterfaceIndex int
	pcapInterface      pcapgo.NgInterface

	StartTime time.Time

	source packetSource
	logger *zap.Logger
}

func (ci *captureInterface) Flush() {
}

func (ci *captureInterface) readPackets(c *Capture) {
	for {
		if err := ci.readPacket(c); err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				c.logger.Error("Failed to read packet data", zap.Error(err))
				continue
			}
		}
	}
}

func (ci *captureInterface) readPacket(c *Capture) error {
	var err error
	var cp CapturePacket

	if cp.Data, cp.CaptureInfo, err = ci.source.ReadPacketData(); err != nil {
		return err
	}

	cp.Interface = ci

	c.newPacket(cp)

	return nil
}
