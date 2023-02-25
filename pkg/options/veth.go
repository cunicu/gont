// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"net"

	nl "github.com/vishvananda/netlink"
)

type PeerHardwareAddress net.HardwareAddr

func (p PeerHardwareAddress) ApplyVeth(v *nl.Veth) {
	v.PeerHardwareAddr = net.HardwareAddr(p)
}
